package videoworker

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"testing"
)

func TestProcessVideo_isNewLive(t *testing.T) {
	type fields struct {
		liveStatus    *interfaces.LiveStatus
		videoPathList VideoPathList
		liveTrace     live.LiveTrace
		monitor       monitor.VideoMonitor
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Now false", fields{
			liveStatus: &interfaces.LiveStatus{video: &interfaces.VideoInfo{
				Title:       "1",
				Provider:    "mock",
				Target:      "3",
				UsersConfig: config.UsersConfig{},
			}, isLive: true},
			liveTrace: live.GetLiveStatus,
			monitor: &monitor.Mock{
				Video: &interfaces.VideoInfo{
					Title:       "",
					Target:      "",
					UsersConfig: config.UsersConfig{},
				},
				IsLive: false,
			},
		}, true},
		{"Now true but same", fields{
			liveStatus: &interfaces.LiveStatus{video: &interfaces.VideoInfo{
				Title:       "1",
				Provider:    "mock",
				Target:      "3",
				UsersConfig: config.UsersConfig{},
			}, isLive: true},
			liveTrace: live.GetLiveStatus,
			monitor: &monitor.Mock{
				Video: &interfaces.VideoInfo{
					Title:       "1",
					Target:      "3",
					UsersConfig: config.UsersConfig{},
				},
				IsLive: true,
			},
		}, false},
		{"Now true and title same but new link", fields{
			liveStatus: &interfaces.LiveStatus{video: &interfaces.VideoInfo{
				Title:       "1",
				Provider:    "mock",
				Target:      "3",
				UsersConfig: config.UsersConfig{},
			}, isLive: true},
			liveTrace: live.GetLiveStatus,
			monitor: &monitor.Mock{
				Video: &interfaces.VideoInfo{
					Title:       "1",
					Target:      "4",
					UsersConfig: config.UsersConfig{},
				},
				IsLive: true,
			},
		}, true},
		{"Now true and link same but new title", fields{
			liveStatus: &interfaces.LiveStatus{video: &interfaces.VideoInfo{
				Title:       "1",
				Provider:    "mock",
				Target:      "3",
				UsersConfig: config.UsersConfig{},
			}, isLive: true},
			liveTrace: live.GetLiveStatus,
			monitor: &monitor.Mock{
				Video: &interfaces.VideoInfo{
					Title:       "2",
					Target:      "3",
					UsersConfig: config.UsersConfig{},
				},
				IsLive: true,
			},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProcessVideo{
				LiveStatus:    tt.fields.liveStatus,
				videoPathList: tt.fields.videoPathList,
				LiveTrace:     tt.fields.liveTrace,
				Monitor:       tt.fields.monitor,
			}
			if got := p.isNewLive(); got != tt.want {
				t.Errorf("isNewLive() = %v, want %v", got, tt.want)
			}
		})
	}
}
