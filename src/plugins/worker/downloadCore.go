package worker

import (
	"Vtb_Record/src/plugins/structUtils"
	"Vtb_Record/src/utils"
	"errors"
	"log"
)

func addStreamlinkProxy(co []string) []string {
	co = append(co, "--http-proxy", "socks5://"+utils.Config.Proxy)
	return co
}
func downloadByStreamlink(video *structUtils.VideoInfo) {
	arg := []string{"--hls-live-restart", "--force", "--hls-timeout", "120", "-o",
		video.FilePath}
	if utils.Config.EnableProxy {
		arg = addStreamlinkProxy(arg)
	}
	arg = append(arg, video.Target, utils.Config.DownloadQuality)
	log.Println(arg)
	utils.ExecShell("streamlink", arg...)
}
func needDownload(video *structUtils.VideoInfo) error {
	if !video.UsersConfig.NeedDownload {
		return errors.New(video.UsersConfig.Name + "needn't download")
	}
	return nil
}
func DownloadVideo(video *structUtils.VideoInfo) string {
	log.Printf("%s|%s start to download", video.Provider, video.UsersConfig.Name)
	video.Title = utils.RemoveIllegalChar(video.Title)
	video.FilePath = utils.GenerateFilepath(video.UsersConfig.Name, video.Title+".ts")
	video.UsersConfig.DownloadDir = utils.GenerateDownloadDir(video.UsersConfig.Name)
	if err := needDownload(video); err != nil {
		return ""
	}
	switch video.Provider {
	case "Youtube":
		//video = getStreamingLink(video)
		//downloadByFFMPEG(video)
		downloadByStreamlink(video)
	case "Twitcasting":
		downloadByStreamlink(video)
	}
	if !utils.IsFileExist(video.FilePath) {
		log.Printf("downloader: %s the video file don't exist", video.Title)
		return ""
	}
	return video.FilePath
}
