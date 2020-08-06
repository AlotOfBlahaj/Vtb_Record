package monitor

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	. "github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"regexp"
	"sync"
	"time"
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

/*func (y *Youtube) getVideoInfo() error {
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
}*/

type YoutubePoller struct {
	LivingUids map[string]LiveInfo
	lock       sync.Mutex
}

var U2bPoller YoutubePoller

func (y *YoutubePoller) parseLiveStatus(rawPage string) error {
	livingUids := make(map[string]LiveInfo)

	re, _ := regexp.Compile(`\["ytInitialData"\]\s*=\s*([^\n]+?});`)
	result := re.FindStringSubmatch(rawPage)
	if len(result) < 1 {
		y.LivingUids = livingUids
		return fmt.Errorf("youtube cannot find js_data")
	}
	jsonYtConfig := result[1]
	items := gjson.Get(jsonYtConfig, "contents.twoColumnBrowseResultsRenderer.tabs.0.tabRenderer.content.sectionListRenderer.contents.0.itemSectionRenderer.contents.0.shelfRenderer.content.gridRenderer.items")
	itemArr := items.Array()
	for _, item := range itemArr {
		style := item.Get("gridVideoRenderer.badges.0.metadataBadgeRenderer.style")

		if style.String() == "BADGE_STYLE_TYPE_LIVE_NOW" {
			channelId := item.Get("gridVideoRenderer.shortBylineText.runs.0.navigationEndpoint.browseEndpoint.browseId")
			//title := item.Get("gridVideoRenderer.shortBylineText.runs.0.text")
			videoId := item.Get("gridVideoRenderer.videoId")
			//video_thumbnail := item.Get("gridVideoRenderer.thumbnail.thumbnails.0.url")
			videoTitle := item.Get("gridVideoRenderer.title.simpleText")
			//upcomingEventData := item.Get("gridVideoRenderer.upcomingEventData")

			livingUids[channelId.String()] = LiveInfo{
				Title:         videoTitle.String(),
				StreamingLink: "https://www.youtube.com/watch?v=" + videoId.String(),
			}
		}

	}

	y.LivingUids = livingUids
	log.Debugf("Parsed uids: %s", y.LivingUids)
	return nil
}

func (y *YoutubePoller) getLiveStatus() error {
	ctx := getCtx("Youtube")
	//mod := getMod("Youtube")
	apihosts := []string{
		"https://nameless-credit-7c9e.misty.workers.dev",
		"https://delicate-cherry-9564.vtbrecorder1.workers.dev",
		"https://plain-truth-41a9.vtbrecorder2.workers.dev",
		"https://snowy-shape-95ae.vtbrecorder3.workers.dev",
	}

	rawPage, err := ctx.HttpGet(
		RandChooseStr(apihosts)+"/feed/subscriptions/",
		map[string]string{})
	if err != nil {
		return err
	}

	page := string(rawPage)
	return y.parseLiveStatus(page)
}

func (y *YoutubePoller) GetStatus() error {
	return y.getLiveStatus()
}

func (y *YoutubePoller) StartPoll() error {
	err := y.GetStatus()
	if err != nil {
		return err
	}
	mod := getMod("Youtube")
	_interval, ok := mod.ExtraConfig["PollInterval"]
	interval := time.Duration(config.Config.CriticalCheckSec) * time.Second
	if ok {
		interval = time.Duration(_interval.(float64)) * time.Second
	}
	go func() {
		for {
			time.Sleep(interval)
			err := y.GetStatus()
			if err != nil {
				log.Warnf("Error during polling GetStatus: %s", err)
			}
		}
	}()
	return nil
}

func (y *YoutubePoller) IsLiving(uid string) *LiveInfo {
	y.lock.Lock()
	if y.LivingUids == nil {
		err := y.StartPoll()
		if err != nil {
			log.Warnf("Failed to poll from youtube: %s", err)
		}
	}
	y.lock.Unlock()
	info, ok := y.LivingUids[uid]
	if ok {
		return &info
	} else {
		return nil
	}
}

func (b *Youtube) getVideoInfoByPoll() error {
	ret := U2bPoller.IsLiving(b.usersConfig.TargetId)
	b.IsLive = ret != nil
	if !b.IsLive {
		return nil
	}

	b.Target = ret.StreamingLink
	b.Title = ret.Title
	return nil
}

func (y *Youtube) CreateVideo(usersConfig config.UsersConfig) *interfaces.VideoInfo {
	if !y.yfConfig.IsLive {
		return &interfaces.VideoInfo{}
	}
	v := &interfaces.VideoInfo{
		Title:       y.Title,
		Date:        GetTimeNow(),
		Target:      y.Target,
		Provider:    "Youtube",
		UsersConfig: usersConfig,
	}
	return v
}
func (y *Youtube) CheckLive(usersConfig config.UsersConfig) bool {
	y.usersConfig = usersConfig
	err := y.getVideoInfoByPoll()
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
