package plugins

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
	"testing"
)

func TestPluginCQBot_LiveStart(t *testing.T) {
	p := &videoworker.ProcessVideo{
		LiveStatus: &interfaces.LiveStatus{
			IsLive: true,
			Video: &interfaces.VideoInfo{
				Title:           "",
				Date:            "",
				Target:          "",
				Provider:        "",
				FileName:        "",
				FilePath:        "",
				UsersConfig:     config.Config.Module[0].Users[0],
				TransRecordPath: "",
			},
		},
	}
	cq := PluginCQBot{}
	cq.LiveStart(p)
}
