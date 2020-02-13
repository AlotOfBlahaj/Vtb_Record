package plugins

import (
	"Vtb_Record/src/plugins/monitor"
	"Vtb_Record/src/plugins/structUtils"
	"Vtb_Record/src/utils"
	"testing"
)

func TestProcessVideo_distributeVideo(t *testing.T) {
	ch := make(chan int)
	go func() {
		<-ch
	}()
	type fields struct {
		liveStatus    *LiveStatus
		videoPathList VideoPathList
		liveTrace     LiveTrace
		monitor       monitor.VideoMonitor
	}
	type args struct {
		end      chan int
		fileName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"1", fields{
			liveStatus: &LiveStatus{video: &structUtils.VideoInfo{
				Title:       "Shiny Smily Story",
				Date:        "2020-01-22 04:31:44",
				UsersConfig: utils.UsersConfig{DownloadDir: "/home/ubuntu/Matsuri"},
			}},
		}, args{
			end:      ch,
			fileName: "Shiny Smily Story.ts",
		}, "/home/ubuntu/Matsuri/Shiny Smily Story.ts"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProcessVideo{
				liveStatus:    tt.fields.liveStatus,
				videoPathList: tt.fields.videoPathList,
				liveTrace:     tt.fields.liveTrace,
				monitor:       tt.fields.monitor,
			}
			if got := p.distributeVideo(tt.args.end, tt.args.fileName); got != tt.want {
				t.Errorf("distributeVideo() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
