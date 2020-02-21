package worker

import (
	"github.com/fzxiao233/Vtb_Record/plugins/structUtils"
	"github.com/fzxiao233/Vtb_Record/utils"
	"log"
)

func addStreamlinkProxy(co []string) []string {
	co = append(co, "--http-proxy", "socks5://"+utils.Config.Proxy)
	return co
}
func downloadByStreamlink(video *structUtils.VideoInfo) {
	arg := []string{"--force", "-o",
		video.FilePath}
	if utils.Config.EnableProxy {
		arg = addStreamlinkProxy(arg)
	}
	arg = append(arg, video.Target, utils.Config.DownloadQuality)
	log.Printf("start to download %s", video.FilePath)
	utils.ExecShell("streamlink", arg...)
}

func DownloadVideo(video *structUtils.VideoInfo) string {
	log.Printf("%s|%s start to download", video.Provider, video.UsersConfig.Name)
	video.Title = utils.RemoveIllegalChar(video.Title)
	video.FilePath = utils.GenerateFilepath(video.UsersConfig.Name, video.Title+".ts")
	video.UsersConfig.DownloadDir = utils.GenerateDownloadDir(video.UsersConfig.Name)
	downloadByStreamlink(video)
	if !utils.IsFileExist(video.FilePath) {
		log.Printf("downloader: %s the video file don't exist", video.Title)
		return ""
	}
	log.Printf("%s download successfully", video.FilePath)
	return video.FilePath
}
