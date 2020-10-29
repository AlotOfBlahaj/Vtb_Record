package provgo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fzxiao233/Vtb_Record/utils"
	"golang.org/x/sync/semaphore"
	"net/http"
	"strings"
	"sync"
	"time"
)

func (d *HLSDownloader) handleAltSegment(segData *HLSSegment) (bool, []error) {
	d.segRl.Take()
	if IsStub {
		return true, nil
	}

	logger := d.Logger.WithField("alt", true)
	downChan := make(chan *bytes.Buffer)
	defer func() {
		defer func() {
			recover()
		}()
		close(downChan)
	}()
	// alt seg download is much slower (because we use mainland node), so longer timeout
	ALT_TIMEOUT := 35 * time.Second

	errs := []error{}
	errMutex := sync.Mutex{}
	doDownload := func(client *http.Client) {
		s := time.Now()
		newbuf, err := utils.HttpGetBuffer(client, segData.Url, d.HLSHeader, nil)
		if err != nil {
			errMutex.Lock()
			errs = append(errs, err)
			errMutex.Unlock()
			if strings.HasSuffix(err.Error(), "404") {
				func() {
					defer func() {
						recover()
					}()
					ch := downChan
					if ch == nil {
						return
					}
					ch <- nil
				}()
			}
		} else {
			usedTime := time.Now().Sub(s)
			if usedTime > ALT_TIMEOUT {
				logger.Infof("Download %d used %s", segData.SegNo, usedTime)
			}
			func() {
				defer func() {
					recover()
				}()
				ch := downChan
				if ch == nil {
					return
				}
				ch <- newbuf
			}()
		}
	}
	onlyAlt := false
	// gotcha104 is tencent yun, only m3u8 blocked the foreign ip, so after that we simply ignore it
	/*if strings.Contains(segData.Url, "gotcha104") {
		onlyAlt = true
	}*/
	i := 0
	clients := d.Clients
	if onlyAlt {
		clients = d.AltClients
		if len(clients) == 0 {
			clients = d.allClients
		}
	} else {
		// TODO: Refactor this
		if strings.Contains(segData.Url, "gotcha105") {
			clients = make([]*http.Client, 0)
			clients = append(clients, d.Clients...)
			clients = append(clients, d.Clients...) // double same client
		} else if strings.Contains(segData.Url, "gotcha104") {
			clients = []*http.Client{}
			clients = append(clients, d.Clients...)
			clients = append(clients, d.AltClients...)
		} else if strings.Contains(segData.Url, "googlevideo.com") {
			clients = []*http.Client{}
			clients = append(clients, d.Clients...)
		}
	}
	round := 0
breakout:
	for {
		i %= len(clients)
		go doDownload(clients[i])
		i += 1
		select {
		case ret := <-downChan:
			close(downChan)
			if ret == nil { // unrecoverable error, so return at once
				errMutex.Lock()
				reterr := make([]error, len(errs))
				copy(reterr, errs)
				errMutex.Unlock()
				return false, reterr
			}
			segData.Data = ret
			break breakout
		case <-time.After(ALT_TIMEOUT):
			// wait 10 second for each download try
		}
		if i == len(clients) {
			//logger.Warnf("Failed all-clients to download segment %d", segData.SegNo)
			round++
		}
		if round == 2 {
			logger.WithField("errors", errs).Warnf("Failed to download alt segment %d after 2 round, giving up", segData.SegNo)
			errMutex.Lock()
			reterr := make([]error, len(errs))
			copy(reterr, errs)
			errMutex.Unlock()
			return true, reterr // true but not setting segment, so not got removed
		}
	}
	return true, nil
}

// download alt m3u8 every 3 seconds
func (d *HLSDownloader) AltDownloader() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		err := d.m3u8Handler(true, d)
		if err != nil {
			if strings.Contains(err.Error(), "aborting") { // for aborting errors, we sleep for a while to avoid too much error
				time.Sleep(10 * time.Second)
			} else {
				d.Logger.WithError(err).Infof("Alt m3u8 download failed")
			}
		}
		if d.AltStopped {
			break
		}
		<-ticker.C
	}
}

// update the alt hls stream's link
func (d *HLSDownloader) AltWorker() {
	logger := d.Logger.WithField("alt", true)
	ticker := time.NewTicker(time.Minute * 40)
	defer ticker.Stop()

	if d.AltHLSUrl == "" {
		d.AltUrlUpdating.Lock()
		d.AltHLSUrl = d.HLSUrl
		d.AltHLSHeader = d.HLSHeader
		d.AltUrlUpdating.Unlock()
	}

	for {
		if d.altforceRefreshChan == nil {
			time.Sleep(120 * time.Second)
			d.altforceRefreshChan = make(chan int)
		}
		select {
		case _ = <-ticker.C:

		case _ = <-d.altforceRefreshChan:
			logger.Info("Got altforceRefresh signal, refresh at once!")
			isClose := false
			func() {
				defer func() {
					panicMsg := recover()
					if panicMsg != nil {
						isClose = true
					}
				}()
				ch := d.altforceRefreshChan
				d.altforceRefreshChan = nil // avoid multiple refresh
				close(ch)
			}()
			if isClose {
				return
			}
		}
		retry := 0
		for {
			retry += 1
			if retry > 1 {
				time.Sleep(30 * time.Second)
				if retry > 5 {
					logger.Warnf("failed to update playlist in 5 attempts, fallback to main hls")
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
			needAbort, err, infoJson := updateInfo(d.Video, "", d.Cookie, true)
			if needAbort {
				logger.WithError(err).Warnf("Alt streamlink requested to abort")
				for {
					if d.AltStopped {
						return
					}
					time.Sleep(10 * time.Second)
				}
			}
			if err != nil {
				logger.Warnf("Failed to update playlist: %s", err)
				continue
			}
			m3u8url, headers, err := parseHttpJson(infoJson)
			if err != nil {
				logger.WithError(err).Warnf("Failed to parse json, rawData: %s", infoJson)
				continue
			}

			logger.Infof("Got new m3u8url: %s", m3u8url)
			if m3u8url == "" {
				logger.Warnf("Got empty m3u8 url...: %s", infoJson)
				continue
			}
			// if we only have qiniu
			if strings.Contains(m3u8url, "gotcha103") {
				// fuck qiniu, we have to specially handle gotcha103...
				logger.Infof("We got qiniu cdn... %s", m3u8url)
				// if we have different althlsurl, then we've got other cdn other than qiniu cdn, so we retry!
				// todo: still fallback if we failed too much
				url1 := d.HLSUrl[strings.Index(d.HLSUrl, "://")+3:]
				url2 := d.AltHLSUrl[strings.Index(d.AltHLSUrl, "://")+3:]
				urlhost1 := url1[:strings.Index(url1, "/")]
				urlhost2 := url2[:strings.Index(url2, "/")]
				if urlhost1 == urlhost2 {
					m3u8url = d.HLSUrl
					headers = d.HLSHeader
				} else {
					logger.Infof("We got a good alt m3u8 before: %s, not replacing it", d.AltHLSUrl)
					m3u8url = ""
					time.Sleep(270 * time.Second) // additional sleep time for this reason
					continue                      // use the retry logic
				}
			}

			if m3u8url != "" {
				logger.Infof("Updated AltHLSUrl: %s", m3u8url)
				d.AltUrlUpdating.Lock()
				d.AltHLSUrl = m3u8url
				d.AltHLSHeader = headers
				d.AltUrlUpdating.Unlock()
			}
			break
		}
		if d.AltStopped {
			return
		}
	}
}

var AltDownSem = semaphore.NewWeighted(8)

// Download the segments located in the alt cache
func (d *HLSDownloader) AltSegDownloader() {
	AltSemaphore.Acquire(context.Background(), 1)
	defer AltSemaphore.Release(1)
	for _, _segNo := range d.AltSeqMap.Keys() {
		segNo := _segNo.(int)
		_segData, ok := d.AltSeqMap.Peek(segNo)
		if ok {
			segData := _segData.(*HLSSegment)
			if segData.Data == nil {
				go func(segNo int, segData *HLSSegment) {
					if segData.Data == nil {
						ret, errs := d.handleAltSegment(segData)
						if !ret {
							d.AltSeqMap.Remove(segNo)
						}
						if errs != nil {
							_ori_errs, loaded := d.altSegErr.LoadOrStore(fmt.Sprintf("%d|%s", segNo, segData.Url), errs)
							ori_errs := _ori_errs.([]error)
							if !loaded {
								for _, c := range errs {
									ori_errs = append(ori_errs, c)
								}
							}
						}
					}
				}(segNo, segData)
				time.Sleep(1 * time.Second)
			}
		}
	}
}

var AltSemaphore = semaphore.NewWeighted(30)

// AltWriter writes the alt hls stream's segments into _tail.ts files
func (d *HLSDownloader) AltWriter() {
	AltSemaphore.Acquire(context.Background(), 1)
	defer AltSemaphore.Release(1)
	defer d.AltSeqMap.Purge()

	if d.AltSeqMap.Len() == 0 {
		d.AltStopped = true
		return
	}
	writer := utils.GetWriter(utils.AddSuffix(d.OutPath, "tail"))
	defer writer.Close()
	d.Logger.Infof("Started to write tail video!")

	// download seg 2 times, in 35 seconds totally
	d.AltSegDownloader()
	time.Sleep(15 * time.Second)
	d.AltStopped = true
	func() {
		defer func() {
			recover()
		}()
		close(d.altforceRefreshChan)
	}()
	d.AltSegDownloader()
	time.Sleep(20 * time.Second)
	errMapCpy := map[string][]error{}
	d.altSegErr.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.([]error)
		errMapCpy[k] = v
		return true
	})
	d.Logger.Infof("Errors during alt download: %v", errMapCpy)
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

	// check the tail video parts
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
	d.Logger.Infof("Got tail segs: %v, key max: %d, min %d", segs, max, min)

	// sometimes the cdn will reset everything back to 1 and then restart, so after wrote the
	// last segments, we try to write the first parts
	resetNo := 0
	if min < 25 {
		for i := min; i < 25; i++ {
			if seg, ok := d.AltSeqMap.Peek(i); ok {
				if seg.(*HLSSegment).Data != nil {
					resetNo = i + 1
					continue
				}
			}
			break
		}
	}

	// select the last part of tail video
	startNo := min
	lastGood := max
	for i := startNo; i <= max; i++ {
		if seg, ok := d.AltSeqMap.Peek(i); ok {
			if seg.(*HLSSegment).Data != nil {
				lastGood = startNo
				continue
			}
		}
		if i > max-3 {
			continue
		}
		startNo = i
	}
	if startNo > max {
		startNo = lastGood
	}

	// write tail videos
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

	if resetNo != 0 {
		for i := min; i < resetNo; i++ {
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
		d.Logger.Infof("Finished writing reset segment %d to %d", 1, resetNo-1)
	}
}
