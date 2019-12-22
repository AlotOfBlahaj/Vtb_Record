package plugins

import (
	"Vtb_Record/src/downloader"
	"Vtb_Record/src/utils"
)

func ProcessVideo(video utils.VideoInfo) {
	var ch chan []int
	go func() {
		downloader.DownloadVideo(video)
		status := <-ch
		status[0] = 1
		ch <- status
	}()
	for true {
	Blocking:
		status := <-ch
		for i := range status {
			if i == 0 {
				goto Blocking
			}
		}
		break
	}
}
