package plugins

import (
	"Vtb_Record/src/plugins/monitor"
	"Vtb_Record/src/plugins/structUtils"
	"Vtb_Record/src/utils"
	"time"
)

type LiveStatus struct {
	isLive bool
	video  *structUtils.VideoInfo
}
type LiveTrace func(monitor monitor.VideoMonitor, usersConfig utils.UsersConfig) *LiveStatus

func GetLiveStatus(monitor monitor.VideoMonitor, usersConfig utils.UsersConfig) *LiveStatus {
	return &LiveStatus{
		isLive: monitor.CheckLive(usersConfig),
		video:  monitor.CreateVideo(usersConfig),
	}
}

func StartMonitor(monitor monitor.VideoMonitor, usersConfig utils.UsersConfig) {
	ticker := time.NewTicker(time.Second * time.Duration(utils.Config.CheckSec))
	for {
		p := &ProcessVideo{liveTrace: GetLiveStatus, monitor: monitor}
		liveStatus := GetLiveStatus(monitor, usersConfig)
		if liveStatus.isLive == true {
			p.liveStatus = liveStatus
			p.StartProcessVideo()
		}
		<-ticker.C
	}
}
