package provbase

/*
	Contains common functions & base types for downloaders
*/

import (
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
)

type DownloadProvider interface {
	StartDownload(video *interfaces.VideoInfo, proxy string, cookie string, filepath string) error
}
type Downloader struct {
	Prov DownloadProvider
}

func (d *Downloader) DownloadVideo(video *interfaces.VideoInfo, proxy string, cookie string, filePath string) string {
	logger := log.WithField("video", video)
	logger.Infof("start to download")
	video.FilePath = filePath
	err := d.Prov.StartDownload(video, proxy, cookie, filePath)
	logger.Infof("finished with status: %s", err)
	if !utils.IsFileExist(filePath) {
		logger.Infof("download failed: %s", err)
		return ""
	}
	logger.Infof("%s download successfully", filePath)
	return filePath
}
