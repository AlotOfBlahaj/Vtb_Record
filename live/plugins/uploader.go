package plugins

import (
	"encoding/json"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
	log "github.com/sirupsen/logrus"
)

type UploadDict struct {
	Title        string
	Filename     string
	Date         string
	Path         string
	User         string
	OriginTitle  string `json:"Origin_Title"`
	ASS          string
	Txt          string
	OriginTarget string `json:"originTarget"`
}

type PluginUploader struct {
}

func (p *PluginUploader) LiveStart(process *videoworker.ProcessVideo) error {
	return nil
}

func (p *PluginUploader) DownloadStart(process *videoworker.ProcessVideo) error {
	video := process.LiveStatus.Video
	u := UploadDict{
		Title:        video.Title,
		Filename:     video.FileName,
		Date:         video.Date,
		Path:         video.FilePath,
		User:         video.UsersConfig.Name,
		OriginTitle:  video.Title,
		ASS:          "",
		Txt:          video.TransRecordPath,
		OriginTarget: video.Target,
	}
	data, _ := json.Marshal(u)
	log.Debug(string(data))
	Publish(data, "upload")
	return nil
}

func (p *PluginUploader) LiveEnd(process *videoworker.ProcessVideo) error {
	return nil
}
