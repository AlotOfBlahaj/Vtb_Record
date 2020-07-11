package monitor

import (
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	. "github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type Twitcasting struct {
	BaseMonitor
	TargetId string
	twitcastingVideoInfo
}
type twitcastingVideoInfo struct {
	IsLive        bool
	StreamingLink string
	Vid           string
}

func (t *Twitcasting) getVideoInfo() error {
	rawInfoJSON, err := t.ctx.HttpGet("https://twitcasting.tv/streamserver.php?target="+t.TargetId+"&mode=client", map[string]string{})
	if err != nil {
		return err
	}
	infoJson, _ := simplejson.NewJson(rawInfoJSON)
	t.StreamingLink = "https://twitcasting.tv/" + t.TargetId
	t.IsLive = infoJson.Get("movie").Get("live").MustBool()
	t.Vid = strconv.Itoa(infoJson.Get("movie").Get("id").MustInt())
	ret, err := t.ctx.HttpGet("https://twitcasting.tv/"+t.TargetId, map[string]string{})
	if err != nil {
		return err
	}
	if strings.Contains(string(ret), "password") {
		log.Warn("TwitCasting has password! ignoring...")
		t.IsLive = false
	}
	return nil
	//log.Printf("%+v", t)
}
func (t *Twitcasting) CreateVideo(usersConfig config.UsersConfig) *interfaces.VideoInfo {
	videoTitle := t.TargetId + "#" + t.Vid
	v := &interfaces.VideoInfo{
		Title:         videoTitle,
		Date:          GetTimeNow(),
		Target:        t.StreamingLink,
		Provider:      "Twitcasting",
		StreamingLink: t.StreamingLink,
		UsersConfig:   usersConfig,
	}
	return v
}
func (t *Twitcasting) CheckLive(usersConfig config.UsersConfig) bool {
	t.TargetId = usersConfig.TargetId
	err := t.getVideoInfo()
	if err != nil {
		t.IsLive = false
	}
	if !t.IsLive {
		NoLiving("Twitcasting", usersConfig.Name)
	}
	return t.IsLive
}
