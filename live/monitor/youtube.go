package monitor

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	. "github.com/fzxiao233/Vtb_Record/utils"
	"regexp"
)

type yfConfig struct {
	IsLive bool
	Title  string
	Target string
}
type Youtube struct {
	BaseMonitor
	yfConfig
	usersConfig config.UsersConfig
}

func (y *Youtube) getVideoInfo() error {
	url := "https://www.youtube.com/channel/" + y.usersConfig.TargetId + "/live"
	headers := y.usersConfig.UserHeaders
	if headers == nil {
		headers = map[string]string{}
	}
	htmlBody, err := y.ctx.HttpGet(url, headers)
	if err != nil {
		return err
	}
	re, _ := regexp.Compile(`ytplayer.config\s*=\s*([^\n]+?});`)
	result := re.FindSubmatch(htmlBody)
	if len(result) < 1 {
		y.IsLive = false
		return fmt.Errorf("youtube cannot find js_data")
	}
	jsonYtConfig := result[1]
	ytConfigJson, _ := simplejson.NewJson(jsonYtConfig)
	playerResponse, _ := simplejson.NewJson([]byte(ytConfigJson.Get("args").Get("player_response").MustString()))
	videoDetails := playerResponse.Get("videoDetails")
	IsLive, err := videoDetails.Get("isLive").Bool()
	if err != nil {
		IsLive = false
	}
	y.Title = videoDetails.Get("title").MustString()
	y.Target = "https://www.youtube.com/watch?v=" + videoDetails.Get("videoId").MustString()
	y.IsLive = IsLive
	return nil
	//log.Printf("%+v", y)
}
func (y *Youtube) CreateVideo(usersConfig config.UsersConfig) *interfaces.VideoInfo {
	if !y.yfConfig.IsLive {
		return &interfaces.VideoInfo{}
	}
	v := &interfaces.VideoInfo{
		Title:         y.Title,
		Date:          GetTimeNow(),
		Target:        y.Target,
		Provider:      "Youtube",
		StreamingLink: "",
		UsersConfig:   usersConfig,
	}
	return v
}
func (y *Youtube) CheckLive(usersConfig config.UsersConfig) bool {
	y.usersConfig = usersConfig
	err := y.getVideoInfo()
	if err != nil {
		y.IsLive = false
	}
	if !y.IsLive {
		NoLiving("Youtube", usersConfig.Name)
	}
	return y.yfConfig.IsLive
}

//func (y *Youtube) StartMonitor(usersConfig UsersConfig) {
//	if y.CheckLive(usersConfig) {
//		ProcessVideo(y.createVideo(usersConfig))
//	}
//}
