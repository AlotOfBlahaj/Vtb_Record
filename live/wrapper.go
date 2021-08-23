package live

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"github.com/fzxiao233/Vtb_Record/live/monitor/base"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
)

func StartMonitor(mon base.VideoMonitor, usersConfig config.UsersConfig, pm videoworker.PluginManager) {
	var liveTrace = func() *interfaces.LiveStatus {
		return &interfaces.LiveStatus{
			IsLive: mon.CheckLive(usersConfig),
			Video:  monitor.GetCleanVideoInfo(mon.CreateVideo(usersConfig)),
		}
	}

	videoworker.StartProcessVideo(liveTrace, mon, pm)
	return
}
