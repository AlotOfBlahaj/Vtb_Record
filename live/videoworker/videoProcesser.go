package videoworker

import (
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"github.com/fzxiao233/Vtb_Record/utils"
	"log"
	"os"
	"time"
)

type VideoPathList []string
type ProcessVideo struct {
	LiveStatus    *interfaces.LiveStatus
	videoPathList VideoPathList
	LiveTrace     monitor.LiveTrace
	Monitor       monitor.VideoMonitor
	Plugins       PluginManager
	isLive        bool
	end           chan int
}

func (p *ProcessVideo) startDownloadVideo(ch chan string) {
	p.videoPathList = VideoPathList{}
	for {
		ctx := p.Monitor.GetCtx()
		proxy, _ := ctx.GetProxy()
		aFilePath := DownloadVideo(p.LiveStatus.Video, proxy)
		if aFilePath != "" {
			p.videoPathList = append(p.videoPathList, aFilePath)
		}
		time.Sleep(time.Millisecond * 100)
		if !p.isLive {
			break
		}
	}
	var videoName string
	if utils.Config.EnableTS2MP4 {
		if len(p.videoPathList) > 1 {
			videoName = p.videoPathList.mergeVideo(p.LiveStatus.Video.Title, p.LiveStatus.Video.UsersConfig.DownloadDir)
		} else {
			videoName = ts2mp4(p.videoPathList[0], p.LiveStatus.Video.UsersConfig.DownloadDir, p.LiveStatus.Video.Title)
		}
	}
	if videoName == "" {
		p.end <- 1
		return
	}
	ch <- videoName
}

func (p *ProcessVideo) isNeedDownload() bool {
	return p.LiveStatus.Video.UsersConfig.NeedDownload
}

func (p *ProcessVideo) StartProcessVideo() {
	log.Printf("%s|%s|%s is living. start to process", p.LiveStatus.Video.Provider, p.LiveStatus.Video.UsersConfig.Name, p.LiveStatus.Video.Title)
	p.isLive = true //  默认在直播中
	ch := make(chan string)
	p.end = make(chan int)
	// plugin liveStart
	go p.Plugins.OnLiveStart(p)
	go p.keepLiveAlive()
	if p.isNeedDownload() {
		go p.Plugins.OnDownloadStart(p)
		go p.startDownloadVideo(ch)
		go p.distributeVideo(<-ch)
	}
	<-p.end
	p.Plugins.OnLiveEnd(p)
}

func (p *ProcessVideo) distributeVideo(fileName string) {
	video := p.LiveStatus.Video
	video.FileName = fileName
	video.FilePath = video.UsersConfig.DownloadDir + "/" + video.FileName
	p.Plugins.OnLiveEnd(p)
	p.end <- 1
}

func (p *ProcessVideo) keepLiveAlive() {
	ticker := time.NewTicker(time.Second * time.Duration(utils.Config.CheckSec))
	for {
		if p.isNewLive() {
			p.isLive = false
			if p.isNeedDownload() {
				return //  需要下载时不由此控制end
			}
			p.end <- 1
			return
		}
		<-ticker.C
	}
}

func (p *ProcessVideo) isNewLive() bool {
	newLiveStatus := p.LiveTrace(p.Monitor, p.LiveStatus.Video.UsersConfig)
	if newLiveStatus.IsLive == false || (p.LiveStatus.IsLive == true && p.LiveStatus.Video.Title != newLiveStatus.Video.Title || p.LiveStatus.Video.StreamingLink != newLiveStatus.Video.StreamingLink) {
		log.Printf("[isNewLive]%s|%s|%s is new live or offline", p.LiveStatus.Video.Provider, p.LiveStatus.Video.UsersConfig.Name, p.LiveStatus.Video.Title)
		return true
	} else {
		log.Printf("[isNewLive]%s|%s|%s KeepAlive", p.LiveStatus.Video.Provider, p.LiveStatus.Video.UsersConfig.Name, p.LiveStatus.Video.Title)
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
