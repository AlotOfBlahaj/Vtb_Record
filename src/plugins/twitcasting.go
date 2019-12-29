package plugins

import (
	. "Vtb_Record/src/utils"
	"github.com/bitly/go-simplejson"
	"strconv"
)

type Twitacasting struct {
	targetId string
	twitcastingVideoInfo
}
type twitcastingVideoInfo struct {
	IsLive        bool
	StreamingLink string
	Vid           string
}

func (t *Twitacasting) getVideoInfo() {
	rawInfoJson := HttpGet("https://twitcasting.tv/streamserver.php?target=" + t.targetId + "&mode=client")
	infoJson, _ := simplejson.NewJson(rawInfoJson)
	t.StreamingLink = "https://twitcasting.tv/" + t.targetId
	t.IsLive = infoJson.Get("movie").Get("live").MustBool()
	t.Vid = strconv.Itoa(infoJson.Get("movie").Get("id").MustInt())
}
func (t Twitacasting) createVideo(usersConfig UsersConfig) VideoInfo {
	videoTitle := t.targetId + "#" + t.Vid
	v := VideoInfo{
		Title:         videoTitle,
		Date:          GetTimeNow(),
		Target:        t.StreamingLink,
		Provider:      "Twitcasting",
		FilePath:      GenerateFilepath(usersConfig.Name, videoTitle),
		StreamingLink: t.StreamingLink,
		UsersConfig:   usersConfig,
	}
	v.CreateLiveMsg()
	return v
}
func (t Twitacasting) CheckLive(usersConfig UsersConfig) bool {
	t.getVideoInfo()
	if t.IsLive {
		ProcessVideo(t.createVideo(usersConfig))
	} else {
		NoLiving("Twitcasting", usersConfig.Name)
	}
	return t.IsLive
}
