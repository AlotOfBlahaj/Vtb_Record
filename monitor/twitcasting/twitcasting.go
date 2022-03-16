package twitcasting

import (
	"context"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/interfaces"
	"github.com/fzxiao233/Vtb_Record/monitor/base"
	. "github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"golang.org/x/sync/semaphore"
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
	t.StreamingLink = "https://twitcasting.tv/" + t.TargetId
	t.IsLive = gjson.GetBytes(rawInfoJSON, "movie.live").Bool()
	t.Vid = gjson.GetBytes(rawInfoJSON, "movie.id").String()
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
