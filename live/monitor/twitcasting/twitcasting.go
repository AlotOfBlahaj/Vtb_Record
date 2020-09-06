package twitcasting

import (
	"context"
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/monitor/base"
	. "github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"strconv"
	"strings"
)

type Twitcasting struct {
	base.BaseMonitor
	TargetId string
	twitcastingVideoInfo
}
type twitcastingVideoInfo struct {
	IsLive        bool
	StreamingLink string
	Vid           string
}

func (t *Twitcasting) getVideoInfo() error {
	rawInfoJSON, err := t.Ctx.HttpGet("https://twitcasting.tv/streamserver.php?target="+t.TargetId+"&mode=client", map[string]string{})
	if err != nil {
		return err
	}
	infoJson, _ := simplejson.NewJson(rawInfoJSON)
	t.StreamingLink = "https://twitcasting.tv/" + t.TargetId
	t.IsLive = infoJson.Get("movie").Get("live").MustBool()
	t.Vid = strconv.Itoa(infoJson.Get("movie").Get("id").MustInt())
	ret, err := t.Ctx.HttpGet("https://twitcasting.tv/"+t.TargetId, map[string]string{})
	if err != nil {
		return err
	}
	if strings.Contains(string(ret), "password") {
		log.WithField("TargetId", t.TargetId).Warn("TwitCasting has password! ignoring...")
		t.IsLive = false
	}
	return nil
	//log.Printf("%+v", t)
}
func (t *Twitcasting) CreateVideo(usersConfig config.UsersConfig) *interfaces.VideoInfo {
	videoTitle := t.TargetId + "#" + t.Vid
	v := &interfaces.VideoInfo{
		Title:       videoTitle,
		Date:        GetTimeNow(),
		Target:      t.StreamingLink,
		Provider:    "Twitcasting",
		UsersConfig: usersConfig,
	}
	return v
}

var TwitSemaphore = semaphore.NewWeighted(3)

func (t *Twitcasting) CheckLive(usersConfig config.UsersConfig) bool {
	TwitSemaphore.Acquire(context.Background(), 1)
	defer TwitSemaphore.Release(1)

	t.TargetId = usersConfig.TargetId
	err := t.getVideoInfo()
	if err != nil {
		t.IsLive = false
	}
	if !t.IsLive {
		base.NoLiving("Twitcasting", usersConfig.Name)
	}
	return t.IsLive
}
