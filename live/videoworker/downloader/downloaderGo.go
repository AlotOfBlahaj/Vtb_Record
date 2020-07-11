package downloader

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DownloaderGo struct {
	Downloader
}

func doDownloadHttp(entry *log.Entry, output string, url string, headers map[string]string) error {
	// Create the file
	out, err := os.Create(output)
	if err != nil {
		return err
	}
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
	Data          []byte
}

type HLSDownloader struct {
	Logger           *log.Entry
	HLSUrl           string
	HLSHeader        map[string]string
	UrlUpdating      sync.Mutex
	client           *http.Client
	Video            *interfaces.VideoInfo
	SeqMap           sync.Map
	Output           io.Writer
	errChan          chan error
	firstSeqChan     chan int
	forceRefreshChan chan int
	FinishSeq        int
	Stopped          bool
}

func (d *HLSDownloader) m3u8Handler() error {
	m3u8retry := 0
	for {
		if m3u8retry >= 1 {
			d.Logger.Infof("m3u8 download retry %d", m3u8retry)
			if d.Stopped {
				return nil
			}
			if m3u8retry%5 == 0 {
				d.Logger.Infof("refreshing m3u8url...", m3u8retry)
				d.forceRefreshChan <- 1
				time.Sleep(10 * time.Second)
			}
		}
		m3u8retry += 1
		if m3u8retry > 15 {
			return fmt.Errorf("Still failed to get valid m3u8 after 15 attempts!")
		}

		d.UrlUpdating.Lock()
		curUrl := d.HLSUrl
		curHeader := d.HLSHeader
		d.UrlUpdating.Unlock()
		parsedurl, err := url.Parse(curUrl)
		if err != nil {
			d.Logger.Warnf("m3u8 url parse fail: %s", err)
			d.forceRefreshChan <- 1
			time.Sleep(10 * time.Second)
			continue
		}
		baseUrl := parsedurl.Scheme + "://" + parsedurl.Host + path.Dir(parsedurl.Path)

		// Get the data
		_m3u8, err := utils.HttpGet(d.client, curUrl, curHeader)
		if err != nil {
			d.Logger.Warnf("m3u8 http get failed: %s, retrying: %s", curUrl, err)
			continue
		}
		m3u8 := string(_m3u8)
		m3u8lines := strings.Split(m3u8, "\n")
		if m3u8lines[0] != "#EXTM3U" {
			d.Logger.Warnf("Failed to parse m3u8, expected %s, got %s", "#EXTM3U", m3u8lines[0])
			continue
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
					d.Logger.Warnf("EXT-X-MEDIA-SEQUENCE malformed: %s", line)
					continue
				}
				curseq = _seq
			} else if strings.HasPrefix(line, "#EXTINF:") {
				segs = append(segs, m3u8lines[i+1])
				i += 1
			} else if strings.HasPrefix(line, "#EXT-X-ENDLIST") {
				d.Logger.Infof("Got HLS end mark!")
				finished = true
			} else if line == "" || strings.HasPrefix(line, "#EXT-X-VERSION") ||
				strings.HasPrefix(line, "#EXT-X-ALLOW-CACHE") ||
				strings.HasPrefix(line, "#EXT-X-TARGETDURATION") {

			} else {
				d.Logger.Debugf("Ignored line: %s", line)
			}
		}

		if curseq == -1 {
			// curseq parse failed
			d.Logger.Warnf("curseq parse failed!!!")
			continue
		}
		m3u8retry = 0
		if d.firstSeqChan != nil {
			d.firstSeqChan <- curseq
			d.firstSeqChan = nil
		}
		for i, seg := range segs {
			segData, loaded := d.SeqMap.LoadOrStore(curseq+i, &HLSSegment{SegNo: curseq + i, SegArriveTime: time.Now(), Url: baseUrl + "/" + seg})
			if !loaded {
				go func(segData *HLSSegment) {
					for {
						data, err := utils.HttpGet(d.client, segData.Url, d.HLSHeader)
						if err != nil {
							d.Logger.Warnf("Err when download segment %s: %s", segData.Url, err)
						} else {
							segData.Data = data
							d.Logger.Debugf("Downloaded segment %d: %s", segData.SegNo, err)
							break
						}
						time.Sleep(5 * time.Second)
						if time.Now().Sub(segData.SegArriveTime) > 100*time.Second {
							d.Logger.Warnf("Failed to download segment %d: %s")
							break
						}
					}
				}(segData.(*HLSSegment))
			}
		}
		if finished {
			d.FinishSeq = curseq + len(segs) - 1
		}
		break
	}
	return nil
}

func (d *HLSDownloader) Downloader() {
	ticker := time.NewTicker(time.Second * 3)
	for {
		err := d.m3u8Handler()
		if err != nil {
			d.errChan <- err
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

func (d *HLSDownloader) Worker() {
	ticker := time.NewTicker(time.Minute * 40)
	for {
		select {
		case _ = <-ticker.C:

		case _ = <-d.forceRefreshChan:
			d.Logger.Info("Got forceRefresh signal, refresh at once!")
		}
		retry := 0
		for {
			retry += 1
			if retry > 1 {
				time.Sleep(30 * time.Second)
				if retry > 20 {
					d.errChan <- fmt.Errorf("failed to update playlist in 20 attempts")
					return
				}
				if d.Stopped {
					return
				}
			}
			err, infoJson := updateInfo(d.Video, "")
			if err != nil {
				d.Logger.Warnf("Failed to update playlist: %s", err)
				continue
			}

			var m3u8url string
			if jret, ok := infoJson.CheckGet("url"); !ok {
				d.Logger.Warnf("Not a good json ret: no url")
				continue
			} else {
				m3u8url = jret.MustString()
			}

			headers := make(map[string]string)
			if jret, ok := infoJson.CheckGet("headers"); !ok {
				d.Logger.Warnf("Not a good json ret: no headers")
				continue
			} else {
				for k, v := range jret.MustMap() {
					headers[k] = v.(string)
				}
			}

			d.Logger.Infof("Got new m3u8url: %s", m3u8url)
			if m3u8url == "" {
				d.Logger.Warnf("Got empty m3u8 url...: %s", infoJson)
				continue
			}
			d.UrlUpdating.Lock()
			d.HLSUrl = m3u8url
			d.HLSHeader = headers
			d.UrlUpdating.Unlock()
			break
		}
		if d.Stopped {
			return
		}
	}
}

func (d *HLSDownloader) Writer() {
	curSeq := <-d.firstSeqChan
	for {
		loadTime := time.Second * 0
		for {
			_val, ok := d.SeqMap.Load(curSeq)
			if ok {
				val := _val.(*HLSSegment)
				if val.Data != nil {
					d.Logger.Debugf("Writing segment %d", curSeq)
					_, err := d.Output.Write(val.Data)
					if err != nil {
						d.errChan <- err
						return
					}
					val.Data = nil
					break
				}
				if curSeq >= 15 {
					d.SeqMap.Delete(curSeq - 15)
				}
			}
			time.Sleep(500 * time.Millisecond)
			loadTime += 500 * time.Millisecond
			if loadTime > 2*time.Minute {
				d.errChan <- fmt.Errorf("Failed to load segment %d within timeout...", curSeq)
				return
			}
			if curSeq == d.FinishSeq { // successfully finished
				d.errChan <- nil
				return
			}
		}
		curSeq += 1
	}

}

func (d *HLSDownloader) startDownload() error {
	d.errChan = make(chan error)
	d.firstSeqChan = make(chan int)
	d.forceRefreshChan = make(chan int)
	go d.Writer()
	go d.Downloader()
	go d.Worker()
	err := <-d.errChan
	if err == nil {
		d.Logger.Infof("HLS Download successfully!")
	} else {
		d.Logger.Infof("HLS Download failed: %s", err)
	}
	d.Stopped = true
	return err
}

func doDownloadHls(entry *log.Entry, output string, video *interfaces.VideoInfo, m3u8url string, headers map[string]string) error {
	// Create the file
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	transport := &http.Transport{
		ResponseHeaderTimeout: 20 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   80 * time.Second,
	}

	d := &HLSDownloader{
		Logger:    entry,
		HLSUrl:    m3u8url,
		HLSHeader: headers,
		client:    client,
		Video:     video,
		Output:    out,
	}

	return d.startDownload()
}

var rl ratelimit.Limiter

func init() {
	rl = ratelimit.New(1)
}

func updateInfo(video *interfaces.VideoInfo, proxy string) (error, *simplejson.Json) {
	rl.Take()
	logger := log.WithField("video", video)
	_arg, ok := video.UsersConfig.ExtraConfig["StreamLinkArgs"]
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
	arg = append(arg, video.Target, config.Config.DownloadQuality)
	logger.Infof("start to query, command %s", arg)
	ret, _ := utils.ExecShellEx(logger, false, "streamlink", arg...)
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

func (d *DownloaderGo) StartDownload(video *interfaces.VideoInfo, proxy string, filepath string) error {
	logger := log.WithField("video", video)
	err, infoJson := updateInfo(video, proxy)
	if err != nil {
		return err
	}
	jret := infoJson.Get("type")
	if jret == nil {
		return fmt.Errorf("Not a good json ret: no type")
	}
	streamtype := jret.MustString()
	if streamtype == "http" || streamtype == "hls" {
		jret := infoJson.Get("url")
		if jret == nil {
			return fmt.Errorf("Not a good json ret: no url")
		}
		url := jret.MustString()
		headers := make(map[string]string)
		jret = infoJson.Get("headers")
		if jret == nil {
			return fmt.Errorf("Not a good json ret: no headers")
		}
		for k, v := range jret.MustMap() {
			headers[k] = v.(string)
		}
		if streamtype == "http" {
			logger.Infof("start to download httpstream %s", url)
			return doDownloadHttp(logger, filepath, url, headers)
		} else {
			logger.Infof("start to download hls stream %s", url)
			return doDownloadHls(logger, filepath, video, url, headers)
		}
	} else {
		return fmt.Errorf("Unknown stream type: %s", streamtype)
	}
}
