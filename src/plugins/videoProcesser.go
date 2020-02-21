package plugins

import (
	"github.com/fzxiao233/Vtb_Record/src/plugins/monitor"
	"github.com/fzxiao233/Vtb_Record/src/plugins/worker"
	"github.com/fzxiao233/Vtb_Record/src/utils"
	"log"
	"os"
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
	p.videoPathList = VideoPathList{}
	for {
		aFilePath := worker.DownloadVideo(p.liveStatus.video)
		time.Sleep(time.Millisecond * 100)
		if aFilePath == "" {
			continue
		}
		p.videoPathList = append(p.videoPathList, aFilePath)
		if p.isNewLive() {
			videoName := p.liveStatus.video.Title + ".ts"
			if len(p.videoPathList) > 1 {
				videoName = p.videoPathList.mergeVideo(p.liveStatus.video.Title, p.liveStatus.video.UsersConfig.DownloadDir)
			} else {
				videoName = ts2mp4(aFilePath, p.liveStatus.video.UsersConfig.DownloadDir, p.liveStatus.video.Title)
			}
			if videoName == "" {
				return
			}
			ch <- videoName
			break
		}
	}
}

func (p *ProcessVideo) isNeedDownload() bool {
	return p.liveStatus.video.UsersConfig.NeedDownload
}

func (p *ProcessVideo) StartProcessVideo() {
	log.Printf("%s|%s|%s is living. start to process", p.liveStatus.video.Provider, p.liveStatus.video.UsersConfig.Name, p.liveStatus.video.Title)
	ch := make(chan string)
	video := p.liveStatus.video
	end := make(chan int)
	go worker.CQBot(video)
	if p.isNeedDownload() {
		p.liveStatus.video.TransRecordPath = worker.StartRecord(video)
		go p.startDownloadVideo(ch)
		go p.distributeVideo(end, <-ch)
	} else {
		go p.keepLiveAlive(end)
	}
	<-end
	worker.CloseRecord(video)
}

func (p *ProcessVideo) distributeVideo(end chan int, fileName string) {
	video := p.liveStatus.video
	video.FileName = fileName
	video.FilePath = video.UsersConfig.DownloadDir + "/" + video.FileName
	worker.UploadVideo(video)
	end <- 1
}

func (p *ProcessVideo) keepLiveAlive(end chan int) {
	ticker := time.NewTicker(time.Second * 30)
	for {
		if p.isNewLive() {
			end <- 1
			return
		}
		<-ticker.C
	}
}

func (p *ProcessVideo) isNewLive() bool {
	newLiveStatus := p.liveTrace(p.monitor, p.liveStatus.video.UsersConfig)
	if newLiveStatus.isLive == false || (p.liveStatus.isLive == true && p.liveStatus.video.Title != newLiveStatus.video.Title || p.liveStatus.video.StreamingLink != newLiveStatus.video.StreamingLink) {
		log.Printf("%s|%s|%s is new live or offline", p.liveStatus.video.Provider, p.liveStatus.video.UsersConfig.Name, p.liveStatus.video.Title)
		return true
	} else {
		log.Printf("%s|%s|%s KeepAlive", p.liveStatus.video.Provider, p.liveStatus.video.UsersConfig.Name, p.liveStatus.video.Title)
		return false
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
	if !utils.IsFileExist(mergedPath) {
		log.Printf("mergeVideo: %s the video file don't exist", mergedPath)
		return ""
	}
	for _, aPath := range l {
		_ = os.Remove(aPath)
	}
	return mergedName
}

func ts2mp4(tsPath string, downloadDir string, title string) string {
	mp4Name := utils.ChangeName(title + ".mp4")
	mp4Path := downloadDir + "/" + mp4Name
	utils.ExecShell("ffmpeg", "-i", tsPath, "-c", "copy", "-f", "mp4", mp4Path)
	if !utils.IsFileExist(mp4Path) {
		log.Printf("ts2mp4: %s the video file don't exist", mp4Path)
		return ""
	}
	_ = os.Remove(tsPath)
	return mp4Name
}
