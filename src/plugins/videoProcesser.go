package plugins

import (
	"Vtb_Record/src/cqBot"
	"Vtb_Record/src/downloader"
	"Vtb_Record/src/utils"
	"log"
	"time"
)

func ProcessVideo(video utils.VideoInfo) {
	log.Printf("%s|%s is living", video.Provider, video.UsersConfig.Name)
	var ch chan int
	switch video.UsersConfig.NeedDownload {
	case true:
		go func() {
			_ = downloader.DownloadVideo(video)
			ch <- 1
		}()
	case false:
		go func() {
			ticker := time.NewTicker(time.Minute * 1)
			for {
				if CreateVideoMonitor(video.Provider).CheckLive(video.UsersConfig) {
					break
				}
				log.Printf("%s|%s still living", video.Provider, video.UsersConfig.Name)
				<-ticker.C
			}
			ch <- 1
		}()
	}
	go func() {
		_ = cqBot.CQBot(video)
	}()
	<-ch
}
