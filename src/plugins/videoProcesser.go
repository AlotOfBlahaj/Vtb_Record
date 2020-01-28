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
		time.Sleep(time.Millisecond * 100)
		if aFilePath == "" {
			continue
		}
		p.videoPathList = append(p.videoPathList, aFilePath)
		LiveStatus := p.liveTrace(p.monitor, p.liveStatus.video.UsersConfig)
		if LiveStatus.isLive == false ||
			(LiveStatus.video.Title != p.liveStatus.video.Title || LiveStatus.video.Target != p.liveStatus.video.Target) {
			videoName := p.liveStatus.video.Title + ".ts"
			if len(p.videoPathList) > 1 {
				videoName = p.videoPathList.mergeVideo(p.liveStatus.video.Title, p.liveStatus.video.UsersConfig.DownloadDir)
			} else {
				videoName = ts2mp4(aFilePath, p.liveStatus.video.UsersConfig.DownloadDir, p.liveStatus.video.Title)
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
	mergedName := utils.ChangeName(Title + "_merged.mp4")
	mergedPath := downloadDir + "/" + mergedName
	utils.ExecShell("ffmpeg", "-i", co, "-c", "copy", "-f", "mp4", mergedPath)
	return mergedName
}

func ts2mp4(tsPath string, downloadDir string, title string) string {
	mp4Name := utils.ChangeName(title + ".mp4")
	mp4Path := downloadDir + "/" + mp4Name
	utils.ExecShell("ffmpeg", "-i", tsPath, "-c", "copy", "-f", "mp4", mp4Path)
	return mp4Name
}
