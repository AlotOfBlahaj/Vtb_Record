package monitor

import (
	"Vtb_Record/src/plugins/structUtils"
	. "Vtb_Record/src/utils"
	"github.com/bitly/go-simplejson"
	"log"
)

type Bilibili struct {
	TargetId      string
	Title         string
	isLive        bool
	streamingLink string
}

func (b *Bilibili) getVideoInfo() {
	rawInfoJson := HttpGet("https://api.live.bilibili.com/room/v1/Room/getRoomInfoOld?mid=" + b.TargetId)
	infoJson, _ := simplejson.NewJson(rawInfoJson)
	data := infoJson.Get("data")
	b.streamingLink = data.Get("url").MustString()
	b.Title = data.Get("title").MustString()
	b.isLive = I2b(data.Get("liveStatus").MustInt())
	log.Printf("%+v", b)
}

func (b *Bilibili) CreateVideo(usersConfig UsersConfig) *structUtils.VideoInfo {
	v := &structUtils.VideoInfo{
		Title:         b.Title,
		Date:          GetTimeNow(),
		Target:        b.streamingLink,
		Provider:      "Bilibili",
		StreamingLink: b.streamingLink,
		UsersConfig:   usersConfig,
	}
	v.CreateNoticeMsg()
	return v
}

func (b *Bilibili) CheckLive(usersConfig UsersConfig) bool {
	b.TargetId = usersConfig.TargetId
	b.getVideoInfo()
	if !b.isLive {
		NoLiving("Bilibili", usersConfig.Name)
	}
	return b.isLive
}
