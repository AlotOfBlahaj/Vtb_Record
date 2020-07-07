package downloader

import (
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
)

type DownloadProvider interface {
	StartDownload(video *interfaces.VideoInfo, proxy string, filepath string) error
}
type Downloader struct {
	prov DownloadProvider
}

func (d *Downloader) DownloadVideo(video *interfaces.VideoInfo, proxy string, filePath string) string {
	//rl.Take()
	logger := log.WithField("video", video)
	logger.Infof("start to download")
	video.FilePath = filePath
	err := d.prov.StartDownload(video, proxy, filePath)
	logger.Infof("finished with status: %s", err)
	if !utils.IsFileExist(filePath) {
		logger.Infof("%s the video file don't exist", video.Title)
		return ""
	}
	logger.Infof("%s download successfully", filePath)
	return filePath
}

func GetDownloader(providerName string) *Downloader {
	if providerName == "" || providerName == "streamlink" {
		return &Downloader{&DownloaderStreamlink{}}
	} else if providerName == "go" {
		return &Downloader{&DownloaderGo{}}
	} else {
		log.Fatalf("Unknown download provider %s", providerName)
		return nil
	}
}
