package videoworker

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"github.com/fzxiao233/Vtb_Record/live/videoworker/downloader"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

type VideoPathList []string
type LiveTitleHistoryEntry struct {
	Title     string
	StartTime time.Time
}
type ProcessVideo struct {
	LiveStatus    *interfaces.LiveStatus
	TitleHistory  []LiveTitleHistoryEntry
	liveStartTime time.Time
	videoPathList VideoPathList
	LiveTrace     monitor.LiveTrace
	Monitor       monitor.VideoMonitor
	Plugins       PluginManager
	needStop      bool
	triggerChan   chan int
	finish        chan int
}

func StartProcessVideo(LiveTrace monitor.LiveTrace, Monitor monitor.VideoMonitor, Plugins PluginManager) *ProcessVideo {
	p := &ProcessVideo{LiveTrace: LiveTrace, Monitor: Monitor, Plugins: Plugins}
	liveStatus := LiveTrace(Monitor)
	if liveStatus.IsLive {
		p.LiveStatus = liveStatus
		p.appendTitleHistory(p.LiveStatus.Video.Title)
		p.StartProcessVideo()
	}
	return p
}

func (p *ProcessVideo) StartProcessVideo() {
	log.Infof("%s|%s|%s is living. start to process", p.LiveStatus.Video.Provider, p.LiveStatus.Video.UsersConfig.Name, p.LiveStatus.Video.Title)
	p.needStop = false //  默认在直播中
	p.liveStartTime = time.Now()
	p.finish = make(chan int)
	p.triggerChan = make(chan int)
	// plugin liveStart
	go p.Plugins.OnLiveStart(p)
	go p.keepLiveAlive()
	if p.isNeedDownload() {
		go p.Plugins.OnDownloadStart(p)
		go p.startDownloadVideo()
	}
	<-p.finish
	p.Plugins.OnLiveEnd(p)
}

func (p *ProcessVideo) startDownloadVideo() {
	var pathSlice []string
	if !utils.Config.EnableTS2MP4 {
		pathSlice = []string{utils.Config.DownloadDir, p.LiveStatus.Video.UsersConfig.Name,
			p.liveStartTime.Format("20060102 1504")}
	} else {
		pathSlice = []string{utils.Config.DownloadDir, p.LiveStatus.Video.UsersConfig.Name}
	}
	dirpath := strings.Join(pathSlice, "/")
	p.LiveStatus.Video.UsersConfig.DownloadDir = utils.MakeDir(dirpath)
	p.videoPathList = VideoPathList{}
	var failRecord []time.Time
	for {
		ctx := p.Monitor.GetCtx()
		proxy, _ := ctx.GetProxy()
		down := downloader.GetDownloader(p.Monitor.DownloadProvider())
		aFilePath := down.DownloadVideo(p.LiveStatus.Video, proxy)
		if aFilePath != "" {
			p.videoPathList = append(p.videoPathList, aFilePath)
		} else {
			failRecord = append(failRecord, time.Now())
			log.Info("Failed to record, trying to refresh live state!")
			p.triggerChan <- 1 // refresh at once
			if len(failRecord) >= 3 {
				if time.Now().Unix()-failRecord[0].Unix() < 30 {
					log.Info("Waiting for next refresh before we retry")
					failRecord = make([]time.Time, 0)
					time.Sleep(time.Duration(utils.Config.CriticalCheckSec) * time.Second)
				} else {
					failRecord = failRecord[1:]
				}
			}
		}
		time.Sleep(time.Millisecond * 100)
		if p.needStop {
			break
		}
	}

	videoName := p.postProcessing()
	if videoName == "" {
		p.finish <- 1
		return
	} else {
		//video := p.LiveStatus.Video
		//video.FileName = videoName
		//video.FilePath = video.UsersConfig.DownloadDir + "/" + video.FileName
	}
	p.finish <- 1
}

func (p *ProcessVideo) isNeedDownload() bool {
	return p.LiveStatus.Video.UsersConfig.NeedDownload
}

func (p *ProcessVideo) keepLiveAlive() {
	ticker := time.NewTicker(time.Second * time.Duration(utils.Config.NormalCheckSec*3))
	for {
		select {
		case _ = <-ticker.C:
			log.Info("Refreshing live status...")
		case _ = <-p.triggerChan:
			log.Info("Got emergency triggerChan signal, refresh at once!")
		}
		if p.isNewLive() {
			p.needStop = true
			if p.isNeedDownload() {
				return //  需要下载时不由此控制end
			}
			p.finish <- 1
			return
		}
	}
}

func (p *ProcessVideo) appendTitleHistory(title string) {
	p.TitleHistory = append(p.TitleHistory, LiveTitleHistoryEntry{
		Title:     title,
		StartTime: time.Now(),
	})
}

func (p *ProcessVideo) isNewLive() bool {
	newLiveStatus := p.LiveTrace(p.Monitor)
	if newLiveStatus.IsLive == false || p.LiveStatus.IsLive == false || (p.LiveStatus.IsLive == true && p.LiveStatus.Video.StreamingLink != newLiveStatus.Video.StreamingLink) {
		log.Infof("[isNewLive]%s|%s|%s is new live or offline", p.LiveStatus.Video.Provider, p.LiveStatus.Video.UsersConfig.Name, p.LiveStatus.Video.Title)
		return true
	} else {
		if len(p.TitleHistory) == 0 || p.LiveStatus.Video.Title != newLiveStatus.Video.Title {
			log.Infof("Room title changed from %s to %s", p.LiveStatus.Video.Title, newLiveStatus.Video.Title)
			p.appendTitleHistory(newLiveStatus.Video.Title)
		}
		log.Infof("[isNewLive]%s|%s|%s KeepAlive", p.LiveStatus.Video.Provider, p.LiveStatus.Video.UsersConfig.Name, p.LiveStatus.Video.Title)
		return false
	}
}

func (p *ProcessVideo) getFullTitle() string {
	title := fmt.Sprintf("【%s】", p.liveStartTime.Format("2006-01-02"))
	if len(p.TitleHistory) == 0 {
		p.TitleHistory = append(p.TitleHistory, LiveTitleHistoryEntry{
			Title:     p.LiveStatus.Video.Title,
			StartTime: p.liveStartTime,
		})
		log.Warnf("no TitleHistory!")
	}

	for _, titleHistory := range p.TitleHistory {
		title += fmt.Sprintf("【%s】%s", titleHistory.StartTime.Format("15:04:05"), titleHistory.Title)
	}

	return title
}

func (p *ProcessVideo) postProcessing() string {
	if utils.Config.EnableTS2MP4 {
		return p.convertToMp4()
	} else {
		pathSlice := []string{utils.Config.DownloadDir, p.getFullTitle()}
		dirpath := strings.Join(pathSlice, "/")
		os.Rename(p.LiveStatus.Video.UsersConfig.DownloadDir, dirpath)
		return dirpath
	}
}

func (p *ProcessVideo) convertToMp4() string {
	//livetime := p.liveStartTime.Format("2006-01-02 15:04:05")
	//title := fmt.Sprintf("【%s】", livetime) + p.LiveStatus.Video.Title
	title := p.getFullTitle()
	downloadDir := p.LiveStatus.Video.UsersConfig.DownloadDir
	var videoName string
	if len(p.videoPathList) == 0 {
		log.Warnf("videoPathList is empty!!!!")
		log.Warn(p)
	} else if len(p.videoPathList) > 1 {
		mergedName := utils.ChangeName(title + "_merged.mp4")
		mergedPath := downloadDir + "/" + mergedName
		videoName = p.videoPathList.mergeVideo(mergedPath)
	} else {
		mp4Name := utils.ChangeName(title + ".mp4")
		mp4Path := downloadDir + "/" + mp4Name
		videoName = ts2mp4(p.videoPathList[0], mp4Path)
	}
	return videoName
}

func (l VideoPathList) mergeVideo(outpath string) string {
	co := "concat:"
	for _, aPath := range l {
		co += aPath + "|"
	}
	utils.ExecShell("ffmpeg", "-i", co, "-c", "copy", "-f", "mp4", outpath)
	if !utils.IsFileExist(outpath) {
		log.Warnf("mergeVideo: %s the video file don't exist", outpath)
		return ""
	}
	for _, aPath := range l {
		_ = os.Remove(aPath)
	}
	return outpath
}

func ts2mp4(tsPath string, outpath string) string {
	utils.ExecShell("ffmpeg", "-i", tsPath, "-c", "copy", "-f", "mp4", outpath)
	if !utils.IsFileExist(outpath) {
		log.Warnf("ts2mp4: %s the video file don't exist", outpath)
		return ""
	}
	_ = os.Remove(tsPath)
	return outpath
}
