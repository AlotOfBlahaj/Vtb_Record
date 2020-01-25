package plugins

import (
	"Vtb_Record/src/plugins/monitor"
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
	monitor       monitor.VideoMonitor
}

func (p *ProcessVideo) startDownloadVideo(ch chan string) {
	for {
		aFilePath := worker.DownloadVideo(p.liveStatus.video)
		if aFilePath == "" {
			continue
		}
		p.videoPathList = append(p.videoPathList, aFilePath)
                LiveStatus := p.liveTrace(p.monitor, p.liveStatus.video.UsersConfig)
		if LiveStatus.isLive == false ||
			(LiveStatus.video.Title != p.liveStatus.video.Title || LiveStatus.video.Target != p.liveStatus.video.Target)
                {
			videoName := p.liveStatus.video.Title + ".ts"
			if len(p.videoPathList) > 1 {
				videoName = p.videoPathList.mergeVideo(p.liveStatus.video.Title, p.liveStatus.video.UsersConfig.DownloadDir)
			}
			ch <- videoName
			break
		}
		log.Printf("%s|%s KeepAlive", p.liveStatus.video.Provider, p.liveStatus.video.UsersConfig.Name)
	}
}

func (p *ProcessVideo) isNeedDownload() bool {
	return p.liveStatus.video.UsersConfig.NeedDownload
}

func (p *ProcessVideo) StartProcessVideo() {
	log.Printf("%s|%s is living. start to process", p.liveStatus.video.Provider, p.liveStatus.video.UsersConfig.Name)
	ch := make(chan string)
	video := p.liveStatus.video
	end := make(chan int)
	go worker.CQBot(video)
	if p.isNeedDownload() {
		go p.startDownloadVideo(ch)
		go p.distributeVideo(end, <-ch)
	} else {
		go p.keepLiveAlive(end)
	}
	<-end
}

func (p *ProcessVideo) distributeVideo(end chan int, fileName string) string {
	video := p.liveStatus.video
	video.FileName = fileName
	video.FilePath = video.UsersConfig.DownloadDir + "/" + video.FileName
	worker.UploadVideo(video)
	worker.HlsVideo(video)
	end <- 1
	return video.FilePath
}

func (p *ProcessVideo) keepLiveAlive(end chan int) {
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

func (l VideoPathList) mergeVideo(Title string, downloadDir string) string {
	co := "concat:"
	for _, aPath := range l {
		co += aPath + "|"
	}
	mergedName := Title + "_merged.ts"
	mergedPath := downloadDir + mergedName
	utils.ExecShell("ffmpeg", "-i", co, "-c", "copy", mergedPath)
	return mergedName
}
