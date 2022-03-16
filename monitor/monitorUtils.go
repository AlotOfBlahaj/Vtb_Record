package monitor

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/interfaces"
	"github.com/fzxiao233/Vtb_Record/monitor/base"
	"github.com/fzxiao233/Vtb_Record/monitor/bilibili"
	"github.com/fzxiao233/Vtb_Record/monitor/twitcasting"
	"github.com/fzxiao233/Vtb_Record/monitor/youtube"
	"github.com/fzxiao233/Vtb_Record/utils"
)

type VideoMonitor = base.VideoMonitor
type LiveTrace func() *interfaces.LiveStatus

// Monitor is responsible for checking if live starts & live's title/link changed
func CreateVideoMonitor(module config.ModuleConfig) VideoMonitor {
	var monitor VideoMonitor
	//var monitor *BaseMonitor
	ctx := base.CreateMonitorCtx(module)
	base := base.BaseMonitor{Ctx: ctx, Provider: module.DownloadProvider}
	switch module.Name {
	case "Youtube":
		monitor = &youtube.Youtube{BaseMonitor: base}
	case "Twitcasting":
		monitor = &twitcasting.Twitcasting{BaseMonitor: base}
	case "Bilibili":
		monitor = &bilibili.Bilibili{BaseMonitor: base}
	default:
		return nil
	}
	return monitor
}

// sanitize everything in the videoinfo for downloader & plugins
func GetCleanVideoInfo(info *interfaces.VideoInfo) *interfaces.VideoInfo {
	info.Title = utils.RemoveIllegalChar(info.Title)
	return info
}
