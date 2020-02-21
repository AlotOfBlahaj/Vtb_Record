package plugins

import (
	"github.com/fzxiao233/Vtb_Record/plugins/monitor"
	"github.com/fzxiao233/Vtb_Record/plugins/structUtils"
	"github.com/fzxiao233/Vtb_Record/utils"
	"testing"
)

func TestProcessVideo_isNewLive(t *testing.T) {
	type fields struct {
		liveStatus    *LiveStatus
		videoPathList VideoPathList
		liveTrace     LiveTrace
		monitor       monitor.VideoMonitor
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Now false", fields{
			liveStatus: &LiveStatus{video: &structUtils.VideoInfo{
				Title:         "1",
				Provider:      "mock",
				StreamingLink: "3",
				UsersConfig:   utils.UsersConfig{},
			}, isLive: true},
			liveTrace: GetLiveStatus,
			monitor: &monitor.Mock{
				Video: &structUtils.VideoInfo{
					Title:         "",
					StreamingLink: "",
					UsersConfig:   utils.UsersConfig{},
				},
				IsLive: false,
			},
		}, true},
		{"Now true but same", fields{
			liveStatus: &LiveStatus{video: &structUtils.VideoInfo{
				Title:         "1",
				Provider:      "mock",
				StreamingLink: "3",
				UsersConfig:   utils.UsersConfig{},
			}, isLive: true},
			liveTrace: GetLiveStatus,
			monitor: &monitor.Mock{
				Video: &structUtils.VideoInfo{
					Title:         "1",
					StreamingLink: "3",
					UsersConfig:   utils.UsersConfig{},
				},
				IsLive: true,
			},
		}, false},
		{"Now true and title same but new link", fields{
			liveStatus: &LiveStatus{video: &structUtils.VideoInfo{
				Title:         "1",
				Provider:      "mock",
				StreamingLink: "3",
				UsersConfig:   utils.UsersConfig{},
			}, isLive: true},
			liveTrace: GetLiveStatus,
			monitor: &monitor.Mock{
				Video: &structUtils.VideoInfo{
					Title:         "1",
					StreamingLink: "4",
					UsersConfig:   utils.UsersConfig{},
				},
				IsLive: true,
			},
		}, true},
		{"Now true and link same but new title", fields{
			liveStatus: &LiveStatus{video: &structUtils.VideoInfo{
				Title:         "1",
				Provider:      "mock",
				StreamingLink: "3",
				UsersConfig:   utils.UsersConfig{},
			}, isLive: true},
			liveTrace: GetLiveStatus,
			monitor: &monitor.Mock{
				Video: &structUtils.VideoInfo{
					Title:         "2",
					StreamingLink: "3",
					UsersConfig:   utils.UsersConfig{},
				},
				IsLive: true,
			},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProcessVideo{
				liveStatus:    tt.fields.liveStatus,
				videoPathList: tt.fields.videoPathList,
				liveTrace:     tt.fields.liveTrace,
				monitor:       tt.fields.monitor,
			}
			if got := p.isNewLive(); got != tt.want {
				t.Errorf("isNewLive() = %v, want %v", got, tt.want)
			}
		})
	}
}
