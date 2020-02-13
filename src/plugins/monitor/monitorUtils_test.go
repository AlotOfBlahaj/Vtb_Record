package monitor

import (
	"reflect"
	"testing"
)

func TestCreateVideoMonitor(t *testing.T) {
	type args struct {
		moduleName string
	}
	tests := []struct {
		name string
		args args
		want VideoMonitor
	}{
		{"Youtube", args{moduleName: "Youtube"}, &Youtube{}},
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
}

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
