package downloader

import (
	"Vtb_Record/src/utils"
	"errors"
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
func DownloadVideo(video utils.VideoInfo) error {
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
