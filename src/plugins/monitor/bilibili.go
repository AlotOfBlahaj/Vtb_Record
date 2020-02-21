package monitor

import (
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/src/plugins/structUtils"
	. "github.com/fzxiao233/Vtb_Record/src/utils"
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
	b.streamingLink = infoJson.Get("data").Get("url").MustString("")
	b.Title = infoJson.Get("data").Get("title").MustString("")
	b.isLive = I2b(infoJson.Get("data").Get("liveStatus").MustInt(0))
	//log.Printf("%+v", b)
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
