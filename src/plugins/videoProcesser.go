package plugins

import (
	monitor2 "Vtb_Record/src/plugins/monitor"
	"Vtb_Record/src/plugins/worker"
	"Vtb_Record/src/utils"
	"log"
	"time"
)

type VideoPathList []string
type ProcessVideo struct {
	liveStatus    *LiveStatus
	videoPathList VideoPathList
	liveTrace     LiveTrace
	monitor       monitor2.VideoMonitor
}

func (p *ProcessVideo) startDownloadVideo(ch chan string) {
	if !p.liveStatus.video.UsersConfig.NeedDownload {
		return
	}
	for {
		aFilePath := worker.DownloadVideo(p.liveStatus.video)
		p.videoPathList = append(p.videoPathList, aFilePath)
		if p.liveStatus != p.liveTrace(p.monitor, p.liveStatus.video.UsersConfig) {
			videoName := p.videoPathList.mergeVideo(p.liveStatus.video.Title)
			ch <- videoName
			break
		}
		log.Printf("%s|%s KeepAlive", p.liveStatus.video.Provider, p.liveStatus.video.UsersConfig.Name)
	}
}

func (p *ProcessVideo) StartProcessVideo() {
	log.Printf("%s|%s is living. start to process", p.liveStatus.video.Provider, p.liveStatus.video.UsersConfig.Name)
	ch := make(chan string)
	end := make(chan int)
	go p.startDownloadVideo(ch)
	video := p.liveStatus.video
	go worker.CQBot(video)
	go func(ch chan string) {
		if p.liveStatus.video.UsersConfig.NeedDownload {
			video.FileName = <-ch
			video.FilePath = utils.GenerateFilepath(video.UsersConfig.Name, video.FileName)
			worker.UploadVideo(video)
			worker.HlsVideo(video)
			end <- 1
		} else {
			ticker := time.NewTicker(time.Second * 60)
			for {
				if p.liveStatus != p.liveTrace(p.monitor, p.liveStatus.video.UsersConfig) {
					end <- 1
				} else {
					log.Printf("%s|%s KeepAlive", p.liveStatus.video.Provider, p.liveStatus.video.UsersConfig.Name)
				}
				<-ticker.C
			}
		}
	}(ch)
	<-end
}
func (l VideoPathList) mergeVideo(mergedName string) string {
	co := "concat:"
	for _, aPath := range l {
		co += aPath + "|"
	}
	utils.ExecShell("ffmpeg", "-i", co, "-c", "copy", mergedName+".ts")
	return mergedName
}
