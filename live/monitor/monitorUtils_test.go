package monitor

import (
	"testing"
)

/*func TestCreateVideoMonitor(t *testing.T) {
	tests := []struct {
		name string
		args args
		want VideoMonitor
	}{
		{"Youtube", utils.ModuleConfig{
			Name:        "",
			Enable:      false,
			Users:       utils.UsersConfig{
				TargetId:     "",
				Name:         "",
				DownloadDir:  "",
				NeedDownload: false,
				TransBiliId:  "",
				ExtraConfig:  nil,
			},
			ExtraConfig: map[string]interface{}{},
		}, &Youtube{}},
		{"Twitcasting", args{moduleName: "Twitcasting"}, &Twitcasting{}},
		{"Bilibili", args{moduleName: "Bilibili"}, &Bilibili{}},
		{"Nil", args{moduleName: "7216"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateVideoMonitor(tt.args.moduleName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateVideoMonitor() = %v, want %v", got, tt.want)
			}
		})
	}
}*/

func TestNoLiving(t *testing.T) {
	type args struct {
		Provide string
		Name    string
	}
	tests := []struct {
		name string
		args args
	}{
		{"1", args{
			Provide: "moke",
			Name:    "moke",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
