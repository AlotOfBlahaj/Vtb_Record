package downloader

import (
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

var rl ratelimit.Limiter

func init() {
	rl = ratelimit.New(1)
}

type DownloadProvider interface {
	StartDownload(video *interfaces.VideoInfo, proxy string, filepath string) error
}
type Downloader struct {
	prov DownloadProvider
}

func (d *Downloader) DownloadVideo(video *interfaces.VideoInfo, proxy string) string {
	rl.Take()
	log.Infof("[Downloader]%s|%s start to download", video.Provider, video.UsersConfig.Name)
	filePath := utils.GenerateFilepath(video.UsersConfig.DownloadDir, video.Title+".ts")
	video.FilePath = filePath
	err := d.prov.StartDownload(video, proxy, filePath)
	log.Infof("[Downloader] finished with status: %s", err)
	if !utils.IsFileExist(filePath) {
		log.Infof("[Downloader] %s the video file don't exist", video.Title)
		return ""
	}
	log.Infof("[Downloader]%s download successfully", filePath)
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
