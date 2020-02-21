package monitor

import (
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/src/plugins/structUtils"
	. "github.com/fzxiao233/Vtb_Record/src/utils"
	"regexp"
)

type yfConfig struct {
	IsLive bool
	Title  string
	Target string
}
type Youtube struct {
	yfConfig
	Url string
}

func (y *Youtube) getVideoInfo() {
	htmlBody := HttpGet(y.Url)
	re, _ := regexp.Compile(`ytplayer.config\s*=\s*([^\n]+?});`)
	result := re.FindSubmatch(htmlBody)
	if len(result) < 1 {
		y.IsLive = false
		return
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
	//log.Printf("%+v", y)
}
func (y *Youtube) CreateVideo(usersConfig UsersConfig) *structUtils.VideoInfo {
	if !y.yfConfig.IsLive {
		return &structUtils.VideoInfo{}
	}
	v := &structUtils.VideoInfo{
		Title:         y.Title,
		Date:          GetTimeNow(),
		Target:        y.Target,
		Provider:      "Youtube",
		StreamingLink: "",
		UsersConfig:   usersConfig,
	}
	v.CreateLiveMsg()
	return v
}
func (y *Youtube) CheckLive(usersConfig UsersConfig) bool {
	y.Url = "https://www.youtube.com/channel/" + usersConfig.TargetId + "/live"
	y.getVideoInfo()
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
