package downloader

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	"github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/bytebufferpool"
	"go.uber.org/ratelimit"
	"io"
	"net/http"
	"net/url"
	//"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DownloaderGo struct {
	Downloader
	cookie string
	proxy  string
}

func doDownloadHttp(entry *log.Entry, output string, url string, headers map[string]string, needMove bool) error {
	// Create the file
	/*out, err := os.Create(output)
	if err != nil {
		return err
	}
	if !needMove {
		defer func () {
			go out.Close()
		}()
	} else {
		defer out.Close()
	}*/
	out := utils.GetWriter(output)
	defer out.Close()

	transport := &http.Transport{}

	client := &http.Client{
		Transport: transport,
	}
	// Get the data
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloader got bad status: %s", resp.Status)
	}

	buf := make([]byte, 1024*1024*3) // 1M buffer
	src := resp.Body
	dst := out
	for {
		// Writer the body to file
		written := int64(0)
		for {
			nr, er := src.Read(buf)
			if nr > 0 {
				nw, ew := dst.Write(buf[0:nr])
				if nw > 0 {
					written += int64(nw)
				}
				if ew != nil {
					err = ew
					break
				}
				if nr != nw {
					err = io.ErrShortWrite
					break
				}
			}
			if er != nil {
				err = er
				break
			}
		}

		//written, err := io.CopyBuffer(out, resp.Body, buf)
		entry.Infof("Wrote %s, err: %s", written, err)
		if err == nil {
			return nil
		} else if err == io.EOF {
			entry.Info("Stream ended")
			return nil
		} else {
			return err
		}
	}

	return nil
}

type HLSSegment struct {
	SegNo         int
	SegArriveTime time.Time
	Url           string
	//Data          []byte
	Data *bytes.Buffer
}

type HLSDownloader struct {
	Logger         *log.Entry
	HLSUrl         string
	HLSHeader      map[string]string
	UrlUpdating    sync.Mutex
	AltHLSUrl      string
	AltHLSHeader   map[string]string
	AltUrlUpdating sync.Mutex
	Clients        []*http.Client
	AltClients     []*http.Client
	allClients     []*http.Client
	//bakclient        []*http.Client
	Video               *interfaces.VideoInfo
	SeqMap              sync.Map
	AltSeqMap           *lru.Cache
	OutPath             string
	Output              io.Writer
	errChan             chan error
	alterrChan          chan error
	firstSeqChan        chan int
	forceRefreshChan    chan int
	altforceRefreshChan chan int
	FinishSeq           int
	Stopped             bool
	AltStopped          bool
	Cookie              string
	segRl               ratelimit.Limiter
}

var bufPool bytebufferpool.Pool

func (d *HLSDownloader) handleSegment(segData *HLSSegment, isAlt bool) bool {
	d.segRl.Take()
	logger := d.Logger.WithField("alt", isAlt)
	//downChan := make(chan []byte)
	downChan := make(chan *bytes.Buffer)
	defer func() {
		defer func() {
			recover()
		}()
		close(downChan)
	}()
	doDownload := func(client *http.Client) {
		//buf := bufPool.Get()
		newbuf, err := utils.HttpGetBuffer(client, segData.Url, d.HLSHeader, nil)
		if err != nil {
			logger.Infof("Err when download segment %s: %s", segData.Url, err)
			if strings.HasSuffix(err.Error(), "404") {
				func() {
					defer func() {
						recover()
					}()
					downChan <- nil
				}()
			}
			//bufPool.Put(buf)
		} else {
			func() {
				defer func() {
					recover()
				}()
				downChan <- newbuf
			}()
		}
	}
	onlyAlt := false
	/*if strings.Contains(segData.Url, "gotcha104") {
		onlyAlt = true
	}*/
	i := 0
	clients := d.allClients
	if onlyAlt {
		clients = d.AltClients
		if len(clients) == 0 {
			clients = d.allClients
		}
	}
breakout:
	for {
		/*select {
		case ret := <-downChan:
			close(downChan)
			segData.Data = ret
			if ret == nil {
				return false
			}
			break breakout
		default:
			//
		}*/
		i %= len(clients)
		go doDownload(clients[i])
		//go d.downloadWorker(d.allClients[i], segData.Url, downChan)
		i += 1
		/*
			data, err := utils.HttpGet(d.allClients[0], segData.Url, d.HLSHeader)
			if err != nil {
				d.Logger.Warnf("Err when download segment %s: %s", segData.Url, err)
			} else {
				segData.Data = data
				//f, _ := os.Create("test1")
				//f.Write(segData.Data)
				//f.Close()
				break
			}*/
		select {
		case ret := <-downChan:
			close(downChan)
			segData.Data = ret
			if ret == nil {
				return false
			}
			break breakout
		case <-time.After(10 * time.Second):
			//
		}
		if i == len(clients) {
			logger.Warnf("Failed all-clients to download segment %d", segData.SegNo)
		}
		if time.Now().Sub(segData.SegArriveTime) > 300*time.Second {
			logger.Warnf("Failed to download segment %d within timeout...", segData.SegNo)
			return false
		}
	}
	if isAlt {
		logger.Infof("Downloaded segment %d: len %v", segData.SegNo, segData.Data.Len())
	} else {
		logger.Debugf("Downloaded segment %d: len %v", segData.SegNo, segData.Data.Len())
	}
	return true
}

func (d *HLSDownloader) m3u8Parser(logger *log.Entry, parsedurl *url.URL, m3u8 string, isAlt bool) bool {
	//baseUrl := parsedurl.Scheme + "://" + parsedurl.Host + path.Dir(parsedurl.Path)
	//var baseUrl string
	relaUrl := "http" + "://" + parsedurl.Host + path.Dir(parsedurl.Path)
	hostUrl := "http" + "://" + parsedurl.Host
	getSegUrl := func(url string) string {
		if strings.HasPrefix(url, "http") {
			return url
		} else if url[0:1] == "/" {
			return hostUrl + url
		} else {
			return relaUrl + "/" + url
		}
	}

	m3u8lines := strings.Split(m3u8, "\n")
	if m3u8lines[0] != "#EXTM3U" {
		logger.Warnf("Failed to parse m3u8, expected %s, got %s", "#EXTM3U", m3u8lines[0])
		return false
	}

	curseq := -1
	segs := make([]string, 0)
	i := 0
	finished := false
	for {
		i += 1
		if i >= len(m3u8lines) {
			break
		}
		line := m3u8lines[i]
		if strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE") {
			_, _, val := utils.RPartition(line, ":")
			_seq, err := strconv.Atoi(val)
			if err != nil {
				logger.Warnf("EXT-X-MEDIA-SEQUENCE malformed: %s", line)
				continue
			}
			curseq = _seq
		} else if strings.HasPrefix(line, "#EXTINF:") {
			logger.Debugf("Got seg %d %s", curseq+len(segs), m3u8lines[i+1])
			segs = append(segs, m3u8lines[i+1])
			i += 1
		} else if strings.HasPrefix(line, "#EXT-X-ENDLIST") {
			logger.Debug("Got HLS end mark!")
			finished = true
		} else if line == "" || strings.HasPrefix(line, "#EXT-X-VERSION") ||
			strings.HasPrefix(line, "#EXT-X-ALLOW-CACHE") ||
			strings.HasPrefix(line, "#EXT-X-TARGETDURATION") {
		} else {
			logger.Debugf("Ignored line: %s", line)
		}
	}

	if curseq == -1 {
		// curseq parse failed
		logger.Warnf("curseq parse failed!!!")
		return false
	}
	if !isAlt && d.firstSeqChan != nil {
		d.firstSeqChan <- curseq
		d.firstSeqChan = nil
	}
	for i, seg := range segs {
		if !isAlt {
			segData, loaded := d.SeqMap.LoadOrStore(curseq+i, &HLSSegment{SegNo: curseq + i, SegArriveTime: time.Now(), Url: getSegUrl(seg)})
			if !loaded {
				go d.handleSegment(segData.(*HLSSegment), false)
			}
		} else {
			d.AltSeqMap.PeekOrAdd(curseq+i, &HLSSegment{SegNo: curseq + i, SegArriveTime: time.Now(), Url: getSegUrl(seg)})
		}
	}
	if finished {
		d.FinishSeq = curseq + len(segs) - 1
	}
	return true
}

func (d *HLSDownloader) forceRefresh(isAlt bool) {
	defer func() {
		recover()
	}()
	if !isAlt {
		d.forceRefreshChan <- 1
	} else {
		d.altforceRefreshChan <- 1
	}
}

func (d *HLSDownloader) sendErr(err error) {
	defer func() {
		recover()
	}()
	d.errChan <- err
}

func (d *HLSDownloader) m3u8Handler(isAlt bool) error {
	logger := d.Logger.WithField("alt", isAlt)
	m3u8retry := 0
	retchan := make(chan []byte, 1)
	defer func() {
		defer func() {
			recover()
		}()
		close(retchan)
	}()

	for {
		if retchan == nil {
			retchan = make(chan []byte, 1)
		}
		if m3u8retry >= 1 {
			logger.Infof("m3u8 download retry %d", m3u8retry)
			if d.Stopped {
				return nil
			}
			if m3u8retry%5 == 0 {
				logger.Infof("refreshing m3u8url...", m3u8retry)
				d.forceRefresh(isAlt)
				time.Sleep(10 * time.Second)
			}
		}
		m3u8retry += 1
		if m3u8retry > 5 {
			return fmt.Errorf("Still failed to get valid m3u8 after 15 attempts!")
		}

		var curUrl string
		var curHeader map[string]string
		if !isAlt {
			d.UrlUpdating.Lock()
			curUrl = d.HLSUrl
			curHeader = d.HLSHeader
			d.UrlUpdating.Unlock()
		} else {
			d.AltUrlUpdating.Lock()
			curUrl = d.AltHLSUrl
			curHeader = d.AltHLSHeader
			d.AltUrlUpdating.Unlock()
		}

		if curUrl == "" {
			logger.Infof("got empty m3u8 url", curUrl)
			d.forceRefresh(isAlt)
			time.Sleep(10 * time.Second)
			return nil
		}

		if strings.Contains(curUrl, "gotcha103") {
			//fuck qiniu
			logger.Infof("We got qiniu cdn... %s", curUrl)
			d.forceRefresh(isAlt)
			time.Sleep(300 * time.Second) // alt refresh is expensive
			return nil                    // exit peacefully
		}

		// Get the data
		var err error
		var _m3u8 []byte

		parsedurl, err := url.Parse(curUrl)
		if err != nil {
			logger.Warnf("m3u8 url parse fail: %s", err)
			d.forceRefresh(isAlt)
			time.Sleep(10 * time.Second)
			continue
		}

		if strings.Contains(curUrl, "gotcha104") {
			curUrl = strings.Replace(curUrl, "d1--cn-gotcha104.bilivideo.com", "3hq4yf8r2xgz9.cfc-execute.su.baidubce.com", 1)
		}

		doQuery := func(client *http.Client) {
			//start := time.Now()
			if _, ok := curHeader["Accept-Encoding"]; ok {
				delete(curHeader, "Accept-Encoding")
			}
			_m3u8, err := utils.HttpGet(client, curUrl, curHeader)
			if err != nil {
				if strings.HasSuffix(err.Error(), "404") {
					func() {
						defer func() {
							recover()
						}()
						retchan <- nil // abort!
					}()
				}
				logger.Warnf("Download m3u8 failed with %s", err)
			} else {
				func() {
					defer func() {
						recover()
					}()
					//logger.Debugf("Downloaded m3u8 in %s", time.Now().Sub(start))
					retchan <- _m3u8
				}()
			}
		}

	breakout:
		for i, client := range d.allClients {
			go doQuery(client)
			select {
			case ret := <-retchan:
				close(retchan)
				retchan = nil
				if ret == nil {
					//logger.Info("Unrecoverable m3u8 download err, aborting")
					return fmt.Errorf("Unrecoverable m3u8 download err, aborting")
				}
				_m3u8 = ret
				break breakout
			case <-time.After(time.Millisecond * 1800):
				logger.Debugf("Download m3u8 %s timeout with client %d", curUrl, i)
			}
		}

		if _m3u8 == nil {
			logger.Warnf("m3u8 all-client http get failed: %s, retrying next round", curUrl)
			continue
		}

		m3u8 := string(_m3u8)
		ret := d.m3u8Parser(logger, parsedurl, m3u8, isAlt)
		if ret {
			m3u8retry = 0
		} else {
			continue
		}
		break
	}
	return nil
}

func (d *HLSDownloader) Downloader() {
	ticker := time.NewTicker(time.Second * 3)
	breakflag := false
	for {
		go func() {
			err := d.m3u8Handler(false)
			if err != nil {
				d.sendErr(err)
				breakflag = true
				return
			}
		}()
		if breakflag {
			return
		}
		if d.FinishSeq > 0 {
			d.Stopped = true
		}
		if d.Stopped {
			break
		}
		<-ticker.C
	}
}

func (d *HLSDownloader) AltDownloader() {
	ticker := time.NewTicker(time.Second * 3)
	for {
		err := d.m3u8Handler(true)
		if err != nil {
			d.Logger.Infof("Alt m3u8 download failed, err: %s", err)
		}
		if d.AltStopped {
			break
		}
		<-ticker.C
	}
}

func (d *HLSDownloader) Worker(isAlt bool) {
	ticker := time.NewTicker(time.Minute * 40)
	var pStopped *bool
	if isAlt {
		pStopped = &d.AltStopped
	} else {
		pStopped = &d.Stopped
	}
	for {
		if isAlt {
			<-ticker.C
		} else {
			select {
			case _ = <-ticker.C:

			case _ = <-d.forceRefreshChan:
				d.Logger.Info("Got forceRefresh signal, refresh at once!")
			}
		}
		retry := 0
		for {
			retry += 1
			if retry > 1 {
				time.Sleep(30 * time.Second)
				if retry > 20 {
					if !isAlt {
						d.sendErr(fmt.Errorf("failed to update playlist in 20 attempts"))
					} else {
						d.Logger.Warnf("failed to update playlist in 20 attempts")
					}
					return
				}
				if *pStopped {
					return
				}
			}
			err, infoJson := updateInfo(d.Video, "", d.Cookie, isAlt)
			if err != nil {
				d.Logger.Warnf("Failed to update playlist: %s", err)
				continue
			}
			m3u8url, headers, err := parseHttpJson(infoJson)
			if err != nil {
				d.Logger.Warnf("Failed to parse json ret: %s", err)
				continue
			}

			d.Logger.Infof("Got new m3u8url: %s", m3u8url)
			if m3u8url == "" {
				d.Logger.Warnf("Got empty m3u8 url...: %s", infoJson)
				continue
			}
			if !isAlt {
				d.UrlUpdating.Lock()
				d.HLSUrl = m3u8url
				d.HLSHeader = headers
				d.UrlUpdating.Unlock()
			} else {
				d.AltUrlUpdating.Lock()
				d.AltHLSUrl = m3u8url
				d.AltHLSHeader = headers
				d.AltUrlUpdating.Unlock()
			}
			break
		}
		if *pStopped {
			return
		}
	}
}

func (d *HLSDownloader) AltWorker() {
	logger := d.Logger.WithField("alt", true)
	ticker := time.NewTicker(time.Minute * 40)

	if d.AltHLSUrl == "" {
		d.AltUrlUpdating.Lock()
		d.AltHLSUrl = d.HLSUrl
		d.AltHLSHeader = d.HLSHeader
		d.AltUrlUpdating.Unlock()
	}

	for {
		select {
		case _ = <-ticker.C:

		case _ = <-d.altforceRefreshChan:
			logger.Info("Got forceRefresh signal, refresh at once!")
		}
		retry := 0
		for {

			retry += 1
			if retry > 1 {
				time.Sleep(30 * time.Second)
				if retry > 5 {
					logger.Warnf("failed to update playlist in 20 attempts, fallback to main hls")
					d.AltUrlUpdating.Lock()
					d.AltHLSUrl = d.HLSUrl
					d.AltHLSHeader = d.HLSHeader
					d.AltUrlUpdating.Unlock()
					return
				}
				if d.AltStopped {
					return
				}
			}
			err, infoJson := updateInfo(d.Video, "", d.Cookie, true)
			if err != nil {
				logger.Warnf("Failed to update playlist: %s", err)
				continue
			}
			m3u8url, headers, err := parseHttpJson(infoJson)
			if err != nil {
				logger.Warnf("Failed to parse json ret: %s, rawData: %s", err, infoJson)
				continue
			}

			logger.Infof("Got new m3u8url: %s", m3u8url)
			if m3u8url == "" {
				logger.Warnf("Got empty m3u8 url...: %s", infoJson)
				continue
			}
			d.AltUrlUpdating.Lock()
			d.AltHLSUrl = m3u8url
			d.AltHLSHeader = headers
			d.AltUrlUpdating.Unlock()
			break
		}
		if d.AltStopped {
			return
		}
	}
}

func (d *HLSDownloader) Writer() {
	curSeq := <-d.firstSeqChan
	//firstSeq := curSeq
	for {
		loadTime := time.Second * 0
		//d.Logger.Debugf("Loading segment %d", curSeq)
		for {
			_val, ok := d.SeqMap.Load(curSeq)
			if ok {
				val := _val.(*HLSSegment)
				if curSeq >= 10 {
					d.SeqMap.Delete(curSeq - 10)
				}

				if val.Data != nil {
					timeoutChan := make(chan int, 1)
					go func(timeoutChan chan int, startTime time.Time, segNo int) {
						timer := time.NewTimer(15 * time.Second)
						select {
						case <-timeoutChan:
							d.Logger.Debugf("Wrote segment %d in %s", segNo, time.Now().Sub(startTime))
						case <-timer.C:
							d.Logger.Warnf("Write segment %d too slow...", curSeq)
							timer2 := time.NewTimer(60 * time.Second)
							select {
							case <-timeoutChan:
								d.Logger.Debugf("Wrote segment %d in %s", segNo, time.Now().Sub(startTime))
							case <-timer2.C:
								d.Logger.Errorf("Write segment %d timeout!!!!!!!", curSeq)
							}
						}
					}(timeoutChan, time.Now(), curSeq)
					//_, err := d.Output.Write(val.Data)
					_, err := d.Output.Write(val.Data.Bytes())
					timeoutChan <- 1
					/*if curSeq - firstSeq > 7 {
					time.Sleep(20 * time.Second)
					func() {
						defer func() {
							recover()
						}()
							d.errChan <- fmt.Errorf("test")
						}()
						return
					}*/
					//bufPool.Put(val.Data)
					val.Data = nil
					if err != nil {
						d.sendErr(err)
						return
					}
					break
				}
			} else {
				isLagged := false
				d.SeqMap.Range(func(key, value interface{}) bool {
					if key.(int) > curSeq {
						isLagged = true
						return false
					} else {
						return true
					}
				})
				if loadTime > 25*time.Second {
					d.sendErr(fmt.Errorf("Failed to load segment %d within m3u8 timeout...", curSeq))
					return
				}
			}
			time.Sleep(500 * time.Millisecond)
			loadTime += 500 * time.Millisecond
			if loadTime > 3*time.Minute {
				d.sendErr(fmt.Errorf("Failed to load segment %d within timeout...", curSeq))
				return
			}
			if curSeq == d.FinishSeq { // successfully finished
				d.sendErr(nil)
				return
			}
		}
		curSeq += 1
	}
}

func (d *HLSDownloader) AltWriter() {
	if d.AltSeqMap.Len() == 0 {
		d.AltStopped = true
		return
	}
	writer := utils.GetWriter(utils.AddSuffix(d.OutPath, "tail"))
	defer writer.Close()
	d.Logger.Infof("Started to write tail video!")
	/*output, err := os.Create(utils.AddSuffix(d.OutPath, "tail"))
	if err != nil {
		d.Logger.Warnf("Failed to open tail video, err: %s", err)
	}
	defer output.Close()*/

	refreshDownload := func() {
		for _, _segNo := range d.AltSeqMap.Keys() {
			segNo := _segNo.(int)
			_segData, ok := d.AltSeqMap.Peek(segNo)
			if ok {
				go func(segNo int, segData *HLSSegment) {
					if segData.Data == nil {
						ret := d.handleSegment(segData, true)
						if !ret {
							d.AltSeqMap.Remove(segNo)
						}
						time.Sleep(100 * time.Millisecond)
					}
				}(segNo, _segData.(*HLSSegment))
			}
		}
	}
	refreshDownload()
	time.Sleep(30 * time.Second)
	d.AltStopped = true
	close(d.altforceRefreshChan)
	refreshDownload()
	time.Sleep(60 * time.Second)
	segs := []int{}
	for _, _segNo := range d.AltSeqMap.Keys() {
		segNo := _segNo.(int)
		_segData, ok := d.AltSeqMap.Peek(segNo)
		if ok {
			if _segData.(*HLSSegment).Data != nil {
				segs = append(segs, segNo)
			}
		}
	}
	d.Logger.Infof("Got tail segs: %s", segs)

	min := 10000000000
	max := -1000
	for _, v := range d.AltSeqMap.Keys() {
		if v.(int) < min {
			min = v.(int)
		}
		if v.(int) > max {
			max = v.(int)
		}
	}

	startNo := min
	lastGood := min
	for i := startNo; i <= max; i++ {
		if seg, ok := d.AltSeqMap.Peek(i); ok {
			if seg.(*HLSSegment).Data != nil {
				lastGood = startNo
				continue
			}
		}
		startNo = i
	}
	if startNo > max {
		startNo = lastGood
	}
	d.Logger.Infof("Going to write segment %d to %d", startNo, max)
	var i int
	for i = startNo + 1; i <= max; i++ {
		if _seg, ok := d.AltSeqMap.Peek(i); ok {
			seg := _seg.(*HLSSegment)
			if seg.Data != nil {
				_, err := writer.Write(seg.Data.Bytes())
				//bufPool.Put(seg.Data)
				seg.Data = nil
				if err != nil {
					d.Logger.Warnf("Failed to write to tail video, err: %s", err)
					return
				}
				continue
			}
		}
		break
	}
	d.Logger.Infof("Finished writing segment %d to %d", startNo+1, i)
}

func (d *HLSDownloader) startDownload() error {
	/*output, err := os.Create(d.OutPath)
	if err != nil {
		return err
	}
	d.Output = output
	defer output.Close()*/
	d.segRl = ratelimit.New(2)

	writer := utils.GetWriter(d.OutPath)
	d.Output = writer
	defer writer.Close()

	d.allClients = make([]*http.Client, 0)
	d.allClients = append(d.allClients, d.Clients...)
	d.allClients = append(d.allClients, d.AltClients...)

	d.AltSeqMap, _ = lru.New(32)
	d.errChan = make(chan error)
	d.alterrChan = make(chan error)
	d.firstSeqChan = make(chan int)
	d.forceRefreshChan = make(chan int)
	d.altforceRefreshChan = make(chan int)

	err, altinfoJson := updateInfo(d.Video, "", d.Cookie, true)
	if err == nil {
		alturl, altheaders, err := parseHttpJson(altinfoJson)
		if err == nil {
			d.AltHLSUrl = alturl
			d.AltHLSHeader = altheaders
		}
	}

	go d.Writer()
	go d.Downloader()
	go d.Worker(false)
	hasAlt := false
	if _, ok := d.Video.UsersConfig.ExtraConfig["AltStreamLinkArgs"]; ok {
		hasAlt = true
		d.Logger.Infof("Use alt downloader")
		go d.AltDownloader()
		go func() {
			for {
				d.AltWorker()
				if d.AltStopped {
					break
				}
			}
		}()
	} else {
		d.Logger.Infof("Disabled alt downloader")
	}

	startTime := time.Now()
	err = <-d.errChan
	usedTime := time.Now().Sub(startTime)
	if err == nil {
		d.Logger.Infof("HLS Download successfully!")
	} else {
		d.Logger.Infof("HLS Download failed: %s", err)
		if hasAlt {
			if usedTime > 1*time.Minute {
				go d.AltWriter()
			} else {
				d.AltStopped = true
			}
		}
	}
	close(d.errChan)
	close(d.forceRefreshChan)
	d.Stopped = true
	d.SeqMap = sync.Map{}
	return err
}

func (dd *DownloaderGo) doDownloadHls(entry *log.Entry, output string, video *interfaces.VideoInfo, m3u8url string, headers map[string]string, needMove bool) error {
	// Create the file
	/*out, err := os.Create(output)
	if err != nil {
		return err
	}
	if !needMove {
		defer func () {
			go out.Close()
		}()
	} else {
		defer out.Close()
	}*/

	//client := &http.Client
	clients := []*http.Client{
		{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 20 * time.Second,
				TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
				DisableCompression:    true,
			},
			Timeout: 60 * time.Second,
		},
	}

	_altproxy, ok := video.UsersConfig.ExtraConfig["AltProxy"]
	var altproxy string
	var altclients []*http.Client
	if ok {
		altproxy = _altproxy.(string)
		proxyUrl, _ := url.Parse("socks5://" + altproxy)
		altclients = []*http.Client{
			{
				Transport: &http.Transport{
					TLSNextProto:       make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
					Proxy:              http.ProxyURL(proxyUrl),
					DisableCompression: true,
				},
				Timeout: 100 * time.Second,
			},
		}
	} else {
		altclients = []*http.Client{}
	}

	d := &HLSDownloader{
		Logger:       entry,
		HLSUrl:       m3u8url,
		HLSHeader:    headers,
		AltHLSUrl:    m3u8url,
		AltHLSHeader: headers,
		Clients:      clients,
		AltClients:   altclients,
		Video:        video,
		OutPath:      output,
		Cookie:       dd.cookie,
		//Output:    out,
	}

	err := d.startDownload()
	time.Sleep(1 * time.Second)
	//d.Logger.Infof("Closing output: %s", out.Close())
	utils.ExecShell("/home/misty/rclone", "rc", "vfs/forget", "dir="+path.Dir(output))
	//time.Sleep(time.Second * 30) // wait 10 sec to for underlying fs to sync
	return err
}

var rl ratelimit.Limiter

func init() {
	rl = ratelimit.New(1)
}

func updateInfo(video *interfaces.VideoInfo, proxy string, cookie string, isAlt bool) (error, *simplejson.Json) {
	rl.Take()
	logger := log.WithField("video", video)
	var conf string
	if isAlt {
		conf = "AltStreamLinkArgs"
	} else {
		conf = "StreamLinkArgs"
	}
	_arg, ok := video.UsersConfig.ExtraConfig[conf]
	arg := []string{}
	if ok {
		for _, a := range _arg.([]interface{}) {
			arg = append(arg, a.(string))
		}
	}
	arg = append(arg, []string{"--json"}...)
	if proxy != "" {
		arg = addStreamlinkProxy(arg, proxy)
	}
	if cookie != "" {
		hasCookie := false
		for _, c := range arg {
			if c == "--http-cookies" {
				hasCookie = true
			}
		}
		if !hasCookie {
			arg = append(arg, []string{"--http-cookies", cookie}...)
		}
	}
	arg = append(arg, video.Target, config.Config.DownloadQuality)
	logger.Infof("start to query, command %s", arg)
	ret, stderr := utils.ExecShellEx(logger, false, "streamlink", arg...)
	if stderr != "" {
		log.Infof("Streamlink err output: %s", stderr)
	}
	if ret == "" {
		return fmt.Errorf("streamlink returned unexpected json"), nil
	}
	_ret := []byte(ret)
	infoJson, _ := simplejson.NewJson(_ret)
	if infoJson == nil {
		return fmt.Errorf("JSON parsed failed: %s", ret), nil
	}
	return nil, infoJson
}

func parseHttpJson(infoJson *simplejson.Json) (string, map[string]string, error) {
	jret := infoJson.Get("url")
	if jret == nil {
		return "", nil, fmt.Errorf("Not a good json ret: no url")
	}
	url := jret.MustString()
	headers := make(map[string]string)
	jret = infoJson.Get("headers")
	if jret == nil {
		return "", nil, fmt.Errorf("Not a good json ret: no headers")
	}
	for k, v := range jret.MustMap() {
		headers[k] = v.(string)
	}
	return url, headers, nil
}

func (d *DownloaderGo) StartDownload(video *interfaces.VideoInfo, proxy string, cookie string, filepath string) error {
	logger := log.WithField("video", video)
	d.cookie = cookie
	d.proxy = proxy
	err, infoJson := updateInfo(video, proxy, cookie, false)
	if err != nil {
		return err
	}
	jret := infoJson.Get("type")
	if jret == nil {
		return fmt.Errorf("Not a good json ret: no type")
	}
	streamtype := jret.MustString()
	if streamtype == "" {
		return fmt.Errorf("Not a good json ret: %s", infoJson)
	} else if streamtype == "http" || streamtype == "hls" {
		url, headers, err := parseHttpJson(infoJson)
		if err != nil {
			return err
		}
		//needMove := config.Config.UploadDir == config.Config.DownloadDir
		needMove := false
		if streamtype == "http" {
			logger.Infof("start to download httpstream %s", url)
			return doDownloadHttp(logger, filepath, url, headers, needMove)
		} else {
			logger.Infof("start to download hls stream %s", url)
			return d.doDownloadHls(logger, filepath, video, url, headers, needMove)
		}
	} else {
		return fmt.Errorf("Unknown stream type: %s", streamtype)
	}
}
