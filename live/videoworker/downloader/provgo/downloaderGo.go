package provgo

import (
	"context"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/videoworker/downloader/provbase"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
	"golang.org/x/sync/semaphore"
	"io"
	"os"

	"strings"
)

type DownloaderGo struct {
	provbase.Downloader
	cookie string
	proxy  string
	useAlt bool
}

func addStreamlinkProxy(co []string, proxy string) []string {
	co = append(co, "--http-proxy", "socks5://"+proxy)
	return co
}

var rl ratelimit.Limiter
var randData []byte

func init() {
	rl = ratelimit.New(1)
	randFile, err := os.Open("randData")
	if err == nil {
		randData = make([]byte, 6*1024*1024)
		io.ReadFull(randFile, randData)
	}
}

var StreamlinkSemaphore = semaphore.NewWeighted(3)

func updateInfo(video *interfaces.VideoInfo, proxy string, cookie string, isAlt bool) (needAbort bool, err error, infoJson *simplejson.Json) {
	needAbort = false
	rl.Take()
	logger := log.WithField("video", video).WithField("alt", isAlt)
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
	StreamlinkSemaphore.Acquire(context.Background(), 1)
	ret, stderr := utils.ExecShellEx(logger, false, "streamlink", arg...)
	StreamlinkSemaphore.Release(1)
	if stderr != "" {
		logger.Infof("Streamlink err output: %s", stderr)
		if strings.Contains(stderr, "(abort)") {
			err = fmt.Errorf("streamlink requested abort")
			needAbort = true
			return
		}
	}
	if ret == "" {
		err = fmt.Errorf("streamlink returned unexpected json")
		return
	}
	_ret := []byte(ret)
	infoJson, _ = simplejson.NewJson(_ret)
	if infoJson == nil {
		err = fmt.Errorf("JSON parsed failed: %s", ret)
		return
	}
	slErr := infoJson.Get("error").MustString()
	if slErr != "" {
		err = fmt.Errorf("Streamlink error: " + slErr)
		if strings.Contains(stderr, "(abort)") {
			log.WithField("video", video).WithError(err).Warnf("streamlink requested abort")
			needAbort = true
		}
		return
	}
	err = nil
	return
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
	d.useAlt = false

	var err error
	var infoJson *simplejson.Json
	var streamtype string
	var needAbort bool
	for i := 0; i < 6; i++ {
		if i < 3 {
			needAbort, err, infoJson = updateInfo(video, proxy, cookie, false)
		} else {
			d.useAlt = true
			needAbort, err, infoJson = updateInfo(video, proxy, cookie, true)
		}
		if needAbort {
			// if we didn't entered live
			logger.Warnf("Streamlink requested to abort because: %s", err)
			panic("forceabort")
		}
		if err == nil {
			err = func() error {
				jret := infoJson.Get("type")
				if jret == nil {
					return fmt.Errorf("Not a good json ret: no type")
				}
				streamtype = jret.MustString()
				if streamtype == "" {
					return fmt.Errorf("Not a good json ret: %s", infoJson)
				}
				return nil
			}()
			if err != nil {
				continue
			}
			if streamtype == "http" || streamtype == "hls" {
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
					if strings.Contains(url, "gotcha103") {
						//fuck qiniu
						//entry.Errorf("Not supporting qiniu cdn... %s", m3u8url)
						logger.Warnf("We're getting qiniu cdn... %s, having shitty downloading experiences", url)
						//continue
					}
					logger.Infof("start to download hls stream %s", url)
					return d.doDownloadHls(logger, filepath, video, url, headers, needMove)
				}
			} else {
				return fmt.Errorf("Unknown stream type: %s", streamtype)
			}
		} else {
			logger.WithField("alt", d.useAlt).Infof("Failed to query m3u8 url, err: %s", err)
			if needAbort {
				return fmt.Errorf("abort")
			}
		}
	}
	return err
}
