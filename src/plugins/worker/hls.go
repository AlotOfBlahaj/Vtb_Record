package worker

import (
	"Vtb_Record/src/plugins/structUtils"
	"Vtb_Record/src/utils"
	"encoding/json"
	"log"
)

type HlsDict struct {
	Title string
	Dir   string `json:"Ddir"`
	Path  string
}

func HlsVideo(video *structUtils.VideoInfo) {
	u := &HlsDict{
		Title: video.Title,
		Dir:   video.UsersConfig.DownloadDir,
		Path:  video.FilePath,
	}
	data, _ := json.Marshal(u)
	log.Println(string(data))
	utils.Publish(data, "hls")
}
