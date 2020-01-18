package uploader

import (
	"Vtb_Record/src/utils"
	"encoding/json"
	"github.com/go-redis/redis"
	"log"
)

type UploadDict struct {
	Title       string
	Filename    string
	Date        string
	Path        string
	User        string
	OriginTitle string `json:"Origin_Title"`
	ASS         string
	Txt         string
}
type Uploader interface {
	sendUpload(data []byte)
}
type PubsubUploader struct {
}

func UploadVideo(video *utils.VideoInfo, Filename string, Path string, uploader Uploader) {
	u := UploadDict{
		Title:       video.Title,
		Filename:    Filename,
		Date:        video.Date,
		Path:        Path,
		User:        video.UsersConfig.Name,
		OriginTitle: video.Title,
		ASS:         "",
		Txt:         "",
	}
	data, _ := json.Marshal(u)
	log.Println(string(data))
	uploader.sendUpload(data)
}
func (*PubsubUploader) sendUpload(data []byte) {
	RedisClient := redis.NewClient(
		&redis.Options{
			Addr:     utils.Config.RedisHost,
			Password: "",
			DB:       0,
		})
	_ = RedisClient.Publish("upload", data)
}
