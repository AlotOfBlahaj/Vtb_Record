package videoworker

import (
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	"log"
)

func addStreamlinkProxy(co []string, proxy string) []string {
	co = append(co, "--http-proxy", "socks5://"+proxy)
	return co
}
func downloadByStreamlink(video *interfaces.VideoInfo, proxy string) {
	_arg, ok := video.UsersConfig.ExtraConfig["StreamLinkArgs"]
	arg := []string{}
	if ok {
		for _, a := range _arg.([]interface{}) {
			arg = append(arg, a.(string))
		}
	}
	arg = append(arg, []string{"--force", "-o", video.FilePath}...)
	if proxy != "" {
		arg = addStreamlinkProxy(arg, proxy)
	}
	arg = append(arg, video.Target, utils.Config.DownloadQuality)
	log.Printf("[Downloader]start to download %s, command %s", video.FilePath, arg)
	utils.ExecShell("streamlink", arg...)
}

func DownloadVideo(video *interfaces.VideoInfo, proxy string) string {
	log.Printf("[Downloader]%s|%s start to download", video.Provider, video.UsersConfig.Name)
	video.Title = utils.RemoveIllegalChar(video.Title)
	video.FilePath = utils.GenerateFilepath(video.UsersConfig.Name, video.Title+".ts")
	video.UsersConfig.DownloadDir = utils.GenerateDownloadDir(video.UsersConfig.Name)
	downloadByStreamlink(video, proxy)
	if !utils.IsFileExist(video.FilePath) {
		log.Printf("[Downloader] %s the video file don't exist", video.Title)
		return ""
	}
	log.Printf("[Downloader]%s download successfully", video.FilePath)
	return video.FilePath
}
