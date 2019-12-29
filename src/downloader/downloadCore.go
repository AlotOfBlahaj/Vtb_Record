package downloader

import (
	"Vtb_Record/src/utils"
	"errors"
	"log"
)

// TODO proxy support
func downloadByFFMPEG(video utils.VideoInfo) {
	ExecShell("ffmpeg", "-i", video.Target, "-f",
		"hls", "-hls_time", "3600", "-hls_list_size", "0", video.FilePath)
}
func downloadByStreamlink(video utils.VideoInfo) {
	ExecShell("streamlink", "--hls-live-restart", "--force", "--hls-timeout", "120", "-o",
		video.FilePath, video.StreamingLink, utils.Config.DownloadQuality)
}
func getStreamingLink(video utils.VideoInfo) utils.VideoInfo {
	result := ExecShell("streamlink", "--stream-url", video.Target, utils.Config.DownloadQuality)
	video.StreamingLink = result
	return video
}
func needDownload(video utils.VideoInfo) error {
	if !video.UsersConfig.NeedDownload {
		return errors.New(video.UsersConfig.Name + "needn't download")
	}
	return nil
}
func DownloadVideo(video utils.VideoInfo) error {
	log.Printf("%s|%s start to download", video.Provider, video.UsersConfig.Name)
	if err := needDownload(video); err != nil {
		return err
	}
	switch video.Provider {
	case "Youtube":
		video = getStreamingLink(video)
		downloadByFFMPEG(video)
	case "Twitcasting":
		downloadByStreamlink(video)
	}
	if !utils.IsFileExist(video.FilePath) {
		return errors.New("downloader: the video file don't exist")
	}
	return nil
}
