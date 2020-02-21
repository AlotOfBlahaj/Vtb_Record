package worker

import (
	"encoding/json"
	"github.com/fzxiao233/Vtb_Record/src/plugins/structUtils"
	"github.com/fzxiao233/Vtb_Record/src/utils"
	"log"
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

func UploadVideo(video *structUtils.VideoInfo) {
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
	log.Println(string(data))
	utils.Publish(data, "upload")
}
