package monitor

import (
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/plugins/structUtils"
	. "github.com/fzxiao233/Vtb_Record/utils"
	"strconv"
)

type Twitcasting struct {
	TargetId string
	twitcastingVideoInfo
}
type twitcastingVideoInfo struct {
	IsLive        bool
	StreamingLink string
	Vid           string
}

func (t *Twitcasting) getVideoInfo() error {
	rawInfoJson, err := HttpGet("https://twitcasting.tv/streamserver.php?target="+t.TargetId+"&mode=client", map[string]string{})
	if err != nil {
		return err
	}
	infoJson, _ := simplejson.NewJson(rawInfoJson)
	t.StreamingLink = "https://twitcasting.tv/" + t.TargetId
	t.IsLive = infoJson.Get("movie").Get("live").MustBool()
	t.Vid = strconv.Itoa(infoJson.Get("movie").Get("id").MustInt())
	return nil
	//log.Printf("%+v", t)
}
func (t *Twitcasting) CreateVideo(usersConfig UsersConfig) *structUtils.VideoInfo {
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
func (t *Twitcasting) CheckLive(usersConfig UsersConfig) bool {
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
