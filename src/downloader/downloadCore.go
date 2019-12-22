package downloader

import (
	"Vtb_Record/src/utils"
)

// TODO proxy support
func downloadByFFMPEG(video utils.VideoInfo) {
	ExecShell("ffmpeg", "-i", video.Target, "-f",
		"hls", "-hls_time", "3600", "-hls_list_size", "0", video.Filename)
}
func getStreamingLink(video utils.VideoInfo) utils.VideoInfo {
	result := ExecShell("streamlink", "--stream-url", video.Target, utils.Config.DownloadQuality)
	video.StreamingLink = result
	return video
}
func DownloadVideo(video utils.VideoInfo) {
	video = getStreamingLink(video)
	downloadByFFMPEG(video)
}
