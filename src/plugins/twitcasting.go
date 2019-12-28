package plugins

import (
	. "Vtb_Record/src/utils"
	"github.com/bitly/go-simplejson"
	"strconv"
)

type Twitacasting struct {
	targetId string
	UserName string
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
func (t Twitacasting) createVideo() VideoInfo {
	videoTitle := t.targetId + "#" + t.Vid
	return VideoInfo{
		Title:         videoTitle,
		Date:          GetTimeNow(),
		Target:        t.StreamingLink,
		Provider:      "Twitcasting",
		FilePath:      GenerateFilepath(t.UserName, videoTitle),
		StreamingLink: t.StreamingLink,
	}
}
func TwitcastingCheckLive(usersConfig UsersConfig) {
	t := new(Twitacasting)
	t.targetId = usersConfig.TargetId
	t.UserName = usersConfig.Name
	t.getVideoInfo()
	if t.IsLive {
		ProcessVideo(t.createVideo())
	} else {
		NoLiving("Twitcasting", usersConfig.Name)
	}
}
