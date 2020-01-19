package downloader

import (
	"Vtb_Record/src/utils"
	"errors"
	"log"
)

func addStreamlinkProxy(co []string) []string {
	co = append(co, "--http-proxy", "http://"+utils.Config.Proxy, "--https-proxy", "https://"+utils.Config.Proxy)
	return co
}
func downloadByStreamlink(video *utils.VideoInfo) {
	arg := []string{"--hls-live-restart", "--force", "--hls-timeout", "120", "-o",
		video.FilePath}
	if utils.Config.EnableProxy {
		arg = addStreamlinkProxy(arg)
	}
	arg = append(arg, video.Target, utils.Config.DownloadQuality)
	log.Println(arg)
	utils.ExecShell("streamlink", arg...)
}
func needDownload(video *utils.VideoInfo) error {
	if !video.UsersConfig.NeedDownload {
		return errors.New(video.UsersConfig.Name + "needn't download")
	}
	return nil
}
func DownloadVideo(video *utils.VideoInfo) string {
	log.Printf("%s|%s start to download", video.Provider, video.UsersConfig.Name)
	video.Title = utils.RemoveIllegalChar(video.Title)
	video.FilePath = utils.GenerateFilepath(video.UsersConfig.Name, video.Title)
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
		log.Fatal("downloader: the video file don't exist")
		return ""
	}
	return video.FilePath
}
