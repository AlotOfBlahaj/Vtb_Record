package monitor

import (
	"Vtb_Record/src/plugins/structUtils"
	. "Vtb_Record/src/utils"
	"github.com/bitly/go-simplejson"
	"strconv"
)

type Twitacasting struct {
	TargetId string
	twitcastingVideoInfo
}
type twitcastingVideoInfo struct {
	IsLive        bool
	StreamingLink string
	Vid           string
}

func (t *Twitacasting) getVideoInfo() {
	rawInfoJson := HttpGet("https://twitcasting.tv/streamserver.php?target=" + t.TargetId + "&mode=client")
	infoJson, _ := simplejson.NewJson(rawInfoJson)
	t.StreamingLink = "https://twitcasting.tv/" + t.TargetId
	t.IsLive = infoJson.Get("movie").Get("live").MustBool()
	t.Vid = strconv.Itoa(infoJson.Get("movie").Get("id").MustInt())
	//log.Printf("%+v", t)
}
func (t *Twitacasting) CreateVideo(usersConfig UsersConfig) *structUtils.VideoInfo {
	videoTitle := t.TargetId + "#" + t.Vid
	v := &structUtils.VideoInfo{
		Title:         videoTitle,
		Date:          GetTimeNow(),
		Target:        t.StreamingLink,
		Provider:      "Twitcasting",
		StreamingLink: t.StreamingLink,
		UsersConfig:   usersConfig,
	}
	v.CreateLiveMsg()
	return v
}
func (t *Twitacasting) CheckLive(usersConfig UsersConfig) bool {
	t.TargetId = usersConfig.TargetId
	t.getVideoInfo()
	if !t.IsLive {
		NoLiving("Twitcasting", usersConfig.Name)
	}
	return t.IsLive
}
