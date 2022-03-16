package videoworker

import (
	"context"
	"fmt"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/downloader"
	"github.com/fzxiao233/Vtb_Record/interfaces"
	"github.com/fzxiao233/Vtb_Record/monitor"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"path/filepath"
	"runtime/debug"
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

var limit *rate.Limiter

func init() {
	limit = rate.NewLimiter(rate.Every(time.Second*5), 1)
}

func StartProcessVideo(LiveTrace monitor.LiveTrace, Monitor monitor.VideoMonitor, Plugins PluginManager) {
	p := &ProcessVideo{LiveTrace: LiveTrace, Monitor: Monitor, Plugins: Plugins}
	liveStatus := LiveTrace()
	if liveStatus.IsLive {
		p.LiveStatus = liveStatus
		p.appendTitleHistory(p.LiveStatus.Video.Title)
		limit.Wait(context.Background())
		p.StartProcessVideo()
	}
}

func (p *ProcessVideo) getLogger() *log.Entry {
	return log.WithField("video", p.LiveStatus.Video)
}

func (p *ProcessVideo) StartProcessVideo() {
	p.getLogger().Infof("is living. start to process")
	p.needStop = false //  默认在直播中
	p.liveStartTime = time.Now()
	p.finish = make(chan int)
	p.triggerChan = make(chan int)
	// plugin liveStart
	go p.Plugins.OnLiveStart(p)
	go p.keepLiveAlive()
	if p.isNeedDownload() {
		if err := p.prepareDownload(); err != nil {
			p.getLogger().WithError(err).Warnf("Failed to prepare download")
			p.finish <- -1
		} else {
			go p.Plugins.OnDownloadStart(p)
			go p.startDownloadVideo()
		}
	}
	<-p.finish
	p.Plugins.OnLiveEnd(p)
}

func (p *ProcessVideo) prepareDownload() error {
	var pathSlice []string
	if !config.Config.EnableTS2MP4 {
		pathSlice = []string{config.Config.DownloadDir, p.LiveStatus.Video.UsersConfig.Name,
			p.liveStartTime.Format("20060102 150405")}
	} else {
		pathSlice = []string{config.Config.DownloadDir, p.LiveStatus.Video.UsersConfig.Name}
	}
	dirpath := strings.Join(pathSlice, "/")
	ret, err := utils.MakeDir(dirpath)
	p.getLogger().Debugf("Made directory: %s, ret: %s, err: %s", dirpath, ret, err)
	if err != nil {
		return err
	}
	p.LiveStatus.Video.UsersConfig.DownloadDir = ret
	p.videoPathList = VideoPathList{}
	return nil
}

func (p *ProcessVideo) startDownloadVideo() {
	logger := p.getLogger()
	dirpath := p.LiveStatus.Video.UsersConfig.DownloadDir

	func() {
		defer func() {
			panicMsg := recover()
			if panicMsg == nil {
				return
			}
			if panicMsg == "forceabort" {
				logger.Warn("Downloader requested to abort! Waiting live to finish")
				time.AfterFunc(2*time.Second, func() { // trigger a refresh now
					defer func() {
						recover()
					}()
					p.triggerChan <- 1 // refresh at once
				})
				for {
					time.Sleep(time.Millisecond * 3000)
					if p.needStop {
						break
					}
				}
			} else {
				fmt.Println("panic: ", panicMsg)
				debug.PrintStack()
				panic(panicMsg) // rethrow
			}
		}()

		var failRecord []time.Time
		for {
			ctx := p.Monitor.GetCtx()
			proxy, _ := ctx.GetProxy()
			down := downloader.GetDownloader(p.Monitor.DownloadProvider())
			filePath := utils.GenerateFilepath(dirpath, p.LiveStatus.Video.Title+".ts")
			cookie := ""
			header := p.Monitor.GetCtx().GetHeaders()
			if header != nil {
				cookie = header["Cookie"]
			}
			aFilePath := down.DownloadVideo(p.LiveStatus.Video, proxy, cookie, filePath)

			func() {
				defer func() {
					recover()
				}()
				p.triggerChan <- 1 // refresh at once
			}()

			if aFilePath != "" {
				p.videoPathList = append(p.videoPathList, aFilePath)
			} else {
				failRecord = append(failRecord, time.Now())
				logger.Info("Failed to record, trying to refresh live state!")

				if len(failRecord) >= 3 {
					if time.Now().Unix()-failRecord[0].Unix() < 30 {
						logger.Info("Waiting for next refresh before we retry")
						//failRecord = make([]time.Time, 0)
						//time.Sleep(time.Duration(config.Config.CriticalCheckSec) * time.Second)
						failRecord = failRecord[1:]
						time.Sleep(5 * time.Second) // be patient
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
	}()

	logger.Info("Do postprocessing...")
	videoName := p.postProcessing()
	if videoName != "" {
		video := p.LiveStatus.Video
		video.FileName = videoName
		video.FilePath = video.UsersConfig.DownloadDir + "/" + video.FileName
	}
	p.finish <- 1
}

func (p *ProcessVideo) isNeedDownload() bool {
	return p.LiveStatus.Video.UsersConfig.NeedDownload
}

func (p *ProcessVideo) keepLiveAlive() {
	logger := p.getLogger()
	ticker := time.NewTicker(time.Second * time.Duration(config.Config.NormalCheckSec*3))
	defer ticker.Stop()
	for {
		select {
		case _ = <-ticker.C:
			//logger.Info("Refreshing live status...")
		case _ = <-p.triggerChan:
			logger.Info("Got emergency triggerChan signal, refresh at once!")
		}
		if p.isNewLive() {
			p.needStop = true
			if p.isNeedDownload() {
				close(p.triggerChan)
				return
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
	newLiveStatus := p.LiveTrace()
	logger := p.getLogger()
	if newLiveStatus.IsLive == false || p.LiveStatus.IsLive == false {
		logger.Infof("[isNewLive] live offline")
		return true
	} else if p.LiveStatus.IsLive == true && p.LiveStatus.Video.Target != newLiveStatus.Video.Target {
		logger.Infof("[isNewLive] is new live")
		return true
	} else {
		if len(p.TitleHistory) == 0 || p.LiveStatus.Video.Title != newLiveStatus.Video.Title {
			logger.Infof("Room title changed from %s to %s", p.LiveStatus.Video.Title, newLiveStatus.Video.Title)
			p.LiveStatus.Video.Title = newLiveStatus.Video.Title
			p.appendTitleHistory(newLiveStatus.Video.Title)
		}
		logger.Trace("[isNewLive] KeepAlive")
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
		p.getLogger().Warnf("no TitleHistory!")
	}

	for _, titleHistory := range p.TitleHistory {
		title += fmt.Sprintf("【%s】%s", titleHistory.StartTime.Format("15:04:05"), titleHistory.Title)
	}

	return title
}

func (p *ProcessVideo) postProcessing() string {
	pathSlice := []string{config.Config.DownloadDir, p.LiveStatus.Video.UsersConfig.Name} // , p.getFullTitle()
	dirpath := strings.Join(pathSlice, "/")
	_, err := utils.MakeDir(filepath.Dir(dirpath))
	if err != nil {
		return ""
	}
	if config.Config.EnableTS2MP4 {
		return p.convertToMp4(dirpath)
	}
	return dirpath
}
