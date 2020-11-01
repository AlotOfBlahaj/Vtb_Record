package live

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"github.com/fzxiao233/Vtb_Record/live/monitor/base"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
)

func StartMonitor(mon base.VideoMonitor, usersConfig config.UsersConfig, pm videoworker.PluginManager) {
	//ticker := time.NewTicker(time.Second * time.Duration(utils.Config.CheckSec))
	//for {
	//pm.AddPlugin(&plugins.PluginTranslationRecorder{})
	//pm.AddPlugin(&plugins.PluginUploader{})

	var fun = func(mon base.VideoMonitor) *interfaces.LiveStatus {
		return &interfaces.LiveStatus{
			IsLive: mon.CheckLive(usersConfig),
			Video:  monitor.CleanVideoInfo(mon.CreateVideo(usersConfig)),
		}
	}

	videoworker.StartProcessVideo(fun, mon, pm)
	return
	//<-ticker.C
	//}
}
