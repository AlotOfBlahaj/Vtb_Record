package provstreamlink

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/videoworker/downloader/provbase"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
)

func addStreamlinkProxy(co []string, proxy string) []string {
	co = append(co, "--http-proxy", "socks5://"+proxy)
	return co
}

type DownloaderStreamlink struct {
	provbase.Downloader
}

func (d *DownloaderStreamlink) StartDownload(video *interfaces.VideoInfo, proxy string, cookie string, filepath string) error {
	arg := []string{}
	arg = append(arg, []string{"--force", "-o", filepath}...)
	if proxy != "" {
		arg = addStreamlinkProxy(arg, proxy)
	}
	arg = append(arg, video.Target, config.Config.DownloadQuality)
	logger := log.WithField("video", video)
	logger.Infof("start to download %s, command %s", filepath, arg)
	utils.ExecShell("streamlink", arg...)
	return nil
}
