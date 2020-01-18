package plugins

import (
	"Vtb_Record/src/cqBot"
	"Vtb_Record/src/downloader"
	"Vtb_Record/src/uploader"
	"Vtb_Record/src/utils"
	"log"
	"time"
)

type VideoPathList []string

func ProcessVideo(video *utils.VideoInfo) {
	log.Printf("%s|%s is living", video.Provider, video.UsersConfig.Name)
	var end chan int
	var liveStatus bool
	liveStatus = true
	go func() {
		var VideoPathList VideoPathList
		for {
			if !video.UsersConfig.NeedDownload {
				return
			}
			aFilePath := downloader.DownloadVideo(video)
			VideoPathList = append(VideoPathList, aFilePath)
			if !liveStatus {
				if video.UsersConfig.NeedDownload {
					videoName := VideoPathList.mergeVideo(video.Title)
					uploader.UploadVideo(video, videoName, video.UsersConfig.DownloadDir+"/"+videoName, &uploader.PubsubUploader{})
				}
				end <- 1
				break
			}
		}
	}()
	go func() {
		_ = cqBot.CQBot(video)
	}()
	go func() {
		ticker := time.NewTicker(time.Minute * 1)
		for {
			if !IsLiveAlive(video) {
				liveStatus = false
				if !video.UsersConfig.NeedDownload {
					end <- 1
				}
				break
			}
			<-ticker.C
		}
	}()
	<-end
}
func IsLiveAlive(video *utils.VideoInfo) bool {
	if CreateVideoMonitor(video.Provider).CheckLive(video.UsersConfig) {
		log.Printf("%s|%s still living", video.Provider, video.UsersConfig.Name)
		return true
	}
	return false
}
func (l VideoPathList) mergeVideo(mergedName string) string {
	co := "concat:"
	for _, aPath := range l {
		co += aPath + "|"
	}
	mergedName += ".ts"
	utils.ExecShell("ffmpeg", "-i", co, "-c", "copy", mergedName)
	return mergedName
}
