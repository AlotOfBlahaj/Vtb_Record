package main

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/interfaces"
	"github.com/fzxiao233/Vtb_Record/monitor"
	"github.com/fzxiao233/Vtb_Record/monitor/base"
	"github.com/fzxiao233/Vtb_Record/videoworker"
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
