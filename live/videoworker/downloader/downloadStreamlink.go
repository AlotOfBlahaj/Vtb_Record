package downloader

import (
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
)

func addStreamlinkProxy(co []string, proxy string) []string {
	co = append(co, "--http-proxy", "socks5://"+proxy)
	return co
}

type DownloaderStreamlink struct {
	Downloader
}

func (d *DownloaderStreamlink) StartDownload(video *interfaces.VideoInfo, proxy string, filepath string) error {
	_arg, ok := video.UsersConfig.ExtraConfig["StreamLinkArgs"]
	arg := []string{}
	if ok {
		for _, a := range _arg.([]interface{}) {
			arg = append(arg, a.(string))
		}
	}
	arg = append(arg, []string{"--force", "-o", filepath}...)
	if proxy != "" {
		arg = addStreamlinkProxy(arg, proxy)
	}
	arg = append(arg, video.Target, utils.Config.DownloadQuality)
	log.Infof("[DownloaderStreamlink]start to download %s, command %s", filepath, arg)
	utils.ExecShell("streamlink", arg...)
	return nil
}
