package downloader

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
)

type DownloaderGo struct {
	Downloader
}

func doDownloadHttp(output string, url string, headers map[string]string) error {
	// Create the file
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: 3,
		}).DialContext,
	}

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

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (d *DownloaderGo) StartDownload(video *interfaces.VideoInfo, proxy string, filepath string) error {
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
	arg = append(arg, video.Target, utils.Config.DownloadQuality)
	log.Infof("[DownloaderGo][%s]start to query, command %s", video.Title, arg)
	ret, _ := utils.ExecShellEx("streamlink", false, arg...)
	if ret == "" {
		return fmt.Errorf("streamlink returned unexpected json")
	}
	_ret := []byte(ret)
	if _ret == nil {
		return fmt.Errorf("failed to create byte array")
	}
	infoJson, _ := simplejson.NewJson(_ret)
	jret := infoJson.Get("type")
	if jret == nil {
		return fmt.Errorf("Not a good json ret: no type")
	}
	streamtype := jret.MustString()
	if streamtype == "http" {
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
		log.Infof("[DownloaderGo][%s] start to download httpstream %s", video.Title, url)
		return doDownloadHttp(filepath, url, headers)
	} else {
		return fmt.Errorf("Unknown stream type: %s", streamtype)
	}
}
