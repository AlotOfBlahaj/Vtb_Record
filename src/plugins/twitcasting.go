package plugins

import (
	. "Vtb_Record/src/utils"
	"github.com/bitly/go-simplejson"
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

func (t Twitacasting) getVideoInfo() twitcastingVideoInfo {
	rawInfoJson := HttpGet("https://twitcasting.tv/streamserver.php?target=" + t.targetId + "&mode=client")
	infoJson, _ := simplejson.NewJson(rawInfoJson)
	t.StreamingLink = "https://twitcasting.tv/" + t.targetId + "/metastream.m3u8"
	t.IsLive = infoJson.Get("movie").Get("live").MustBool()
	t.Vid = infoJson.Get("movie").Get("id").MustString()
	return t.twitcastingVideoInfo
}
func (t Twitacasting) createVideo() VideoInfo {
	videoTitle := t.targetId + "#" + t.Vid
	return VideoInfo{
		Title:         videoTitle,
		Date:          GetTimeNow(),
		Target:        t.StreamingLink,
		Provider:      "Twitcasting",
		Filename:      GenerateFilepath(t.UserName, videoTitle),
		StreamingLink: t.StreamingLink,
	}
}
func TwitcastingCheckLive(userConfig UsersConfig) {
	t := new(Twitacasting)
	t.targetId = userConfig.TargetId
	t.UserName = userConfig.Name
	twitcastingVideoInfo := t.getVideoInfo()
	if twitcastingVideoInfo.IsLive {
		ProcessVideo(t.createVideo())
	}
}
