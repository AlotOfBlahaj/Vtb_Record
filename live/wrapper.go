package live

import (
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"github.com/fzxiao233/Vtb_Record/live/plugins"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
	"github.com/fzxiao233/Vtb_Record/utils"
	"time"
)

func GetLiveStatus(monitor monitor.VideoMonitor, usersConfig utils.UsersConfig) *interfaces.LiveStatus {
	return &interfaces.LiveStatus{
		IsLive: monitor.CheckLive(usersConfig),
		Video:  monitor.CreateVideo(usersConfig),
	}
}

func StartMonitor(monitor monitor.VideoMonitor, usersConfig utils.UsersConfig) {
	ticker := time.NewTicker(time.Second * time.Duration(utils.Config.CheckSec))
	for {
		pm := videoworker.PluginManager{}
		pm.AddPlugin(&plugins.PluginCQBot{})
		pm.AddPlugin(&plugins.PluginTranslationRecorder{})
		pm.AddPlugin(&plugins.PluginUploader{})

		p := &videoworker.ProcessVideo{LiveTrace: GetLiveStatus, Monitor: monitor, Plugins: pm}
		liveStatus := GetLiveStatus(monitor, usersConfig)
		if liveStatus.IsLive {
			p.LiveStatus = liveStatus
			p.StartProcessVideo()
		}
		<-ticker.C
	}
}
