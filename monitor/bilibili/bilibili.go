package bilibili

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/interfaces"
	"github.com/fzxiao233/Vtb_Record/monitor/base"
	. "github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

const (
	ApiHost        = "https://api.live.bilibili.com"
	GetRoom        = ApiHost + "/room/v1/Room/getRoomInfoOld?mid="
	LivingPage     = "https://live.bilibili.com/%d"
	LiveApiUrl     = ApiHost + "/room/v1/Room/playUrl?cid=%s&quality=4&platform=web"
	GetRoomBySpace = "https://api.bilibili.com/x/space/acc/info?mid=%s&jsonp=jsonp"
)

type Bilibili struct {
	base.BaseMonitor
	TargetId      string
	Title         string
	isLive        bool
	streamingLink string
	sourceUrl     string
}

func (b *Bilibili) getVideoInfoByRoom() error {
	rawInfoJSON, err := b.Ctx.HttpGet(fmt.Sprintf(GetRoomBySpace, b.TargetId), map[string]string{})
	if err != nil {
		return err
	}
	livestatus := gjson.GetBytes(rawInfoJSON, "data.live_room.liveStatus").Int()
	b.isLive = livestatus == 1
	b.streamingLink = gjson.GetBytes(rawInfoJSON, "data.live_room.url").String()
	b.Title = gjson.GetBytes(rawInfoJSON, "data.live_room.title").String()
	return nil
}

func (b *Bilibili) getSourceUrl() error {
	url := fmt.Sprintf(LiveApiUrl, b.TargetId)
	res, err := b.Ctx.HttpGet(url, map[string]string{})
	if err != nil {
		return err
	}
	urls := gjson.GetBytes(res, "data.durl.#.url").Array()
	if len(urls) < 1 {
		return fmt.Errorf("cannot get download url")
	}
	b.sourceUrl = urls[0].String()
	return nil
}

func (b *Bilibili) CreateVideo(usersConfig config.UsersConfig) *interfaces.VideoInfo {
	v := &interfaces.VideoInfo{
		Title:       b.Title,
		Date:        GetTimeNow(),
		Target:      b.streamingLink,
		Provider:    "Bilibili",
		UsersConfig: usersConfig,
		SourceUrl:   b.sourceUrl,
	}
	return v
}

func (b *Bilibili) CheckLive(usersConfig config.UsersConfig) bool {
	b.TargetId = usersConfig.TargetId
	err := b.getVideoInfoByRoom()

	if err != nil {
		b.isLive = false
		log.WithField("user", fmt.Sprintf("%s|%s", "Bilibili", usersConfig.Name)).WithError(err).Errorf("GetVideoInfo error")
	}
	if !b.isLive {
		base.NoLiving("Bilibili", usersConfig.Name)
	}
	return b.isLive
}
