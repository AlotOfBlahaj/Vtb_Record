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
	LiveStatus := &LiveStatus{}
	ticker := time.NewTicker(time.Second * time.Duration(utils.Config.CheckSec))
	p := &ProcessVideo{liveTrace: GetLiveStatus, monitor: monitor}
	for {
		liveStatus := GetLiveStatus(monitor, usersConfig)
		if liveStatus.isLive == true && liveStatus.video != LiveStatus.video {
			p.liveStatus = liveStatus
			go p.StartProcessVideo()
		}
		<-ticker.C
	}
}
