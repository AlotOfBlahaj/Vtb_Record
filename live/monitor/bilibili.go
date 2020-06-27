package monitor

import (
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	. "github.com/fzxiao233/Vtb_Record/utils"
	"log"
)

type Bilibili struct {
	BaseMonitor
	TargetId      string
	Title         string
	isLive        bool
	streamingLink string
}

func (b *Bilibili) getVideoInfo() error {
	_url, ok := b.ctx.ExtraModConfig["ApiHostUrl"]
	var url string
	if ok {
		url = _url.(string)
	} else {
		url = "https://api.live.bilibili.com"
	}
	rawInfoJSON, err := b.ctx.HttpGet(url+"/room/v1/Room/getRoomInfoOld?mid="+b.TargetId, map[string]string{})
	if err != nil {
		return err
	}
	infoJson, _ := simplejson.NewJson(rawInfoJSON)
	b.streamingLink = infoJson.Get("data").Get("url").MustString("")
	b.Title = infoJson.Get("data").Get("title").MustString("")
	b.isLive = I2b(infoJson.Get("data").Get("liveStatus").MustInt(0))
	return nil
	//log.Printf("%+v", b)
}

func (b *Bilibili) CreateVideo(usersConfig UsersConfig) *interfaces.VideoInfo {
	v := &interfaces.VideoInfo{
		Title:         b.Title,
		Date:          GetTimeNow(),
		Target:        b.streamingLink,
		Provider:      "Bilibili",
		StreamingLink: b.streamingLink,
		UsersConfig:   usersConfig,
	}
	return v
}

func (b *Bilibili) CheckLive(usersConfig UsersConfig) bool {
	b.TargetId = usersConfig.TargetId
	err := b.getVideoInfo()
	if err != nil {
		b.isLive = false
		log.Print(err)
	}
	if !b.isLive {
		NoLiving("Bilibili", usersConfig.Name)
	}
	return b.isLive
}
