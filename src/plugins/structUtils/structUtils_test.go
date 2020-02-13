package structUtils

import (
	"Vtb_Record/src/utils"
	"testing"
)

func TestVideoInfo_CreateLiveMsg(t *testing.T) {
	type fields struct {
		Title           string
		Date            string
		Target          string
		Provider        string
		FileName        string
		FilePath        string
		StreamingLink   string
		UsersConfig     utils.UsersConfig
		CQBotMsg        string
		TransRecordPath string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"1", fields{
			Title:    "test",
			Target:   "test",
			Provider: "test",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VideoInfo{
				Title:           tt.fields.Title,
				Date:            tt.fields.Date,
				Target:          tt.fields.Target,
				Provider:        tt.fields.Provider,
				FileName:        tt.fields.FileName,
				FilePath:        tt.fields.FilePath,
				StreamingLink:   tt.fields.StreamingLink,
				UsersConfig:     tt.fields.UsersConfig,
				CQBotMsg:        tt.fields.CQBotMsg,
				TransRecordPath: tt.fields.TransRecordPath,
			}
			v.CreateLiveMsg()
		})
	}
}

func TestVideoInfo_CreateNoticeMsg(t *testing.T) {
	type fields struct {
		Title           string
		Date            string
		Target          string
		Provider        string
		FileName        string
		FilePath        string
		StreamingLink   string
		UsersConfig     utils.UsersConfig
		CQBotMsg        string
		TransRecordPath string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"1", fields{
			Title:    "test",
			Target:   "test",
			Provider: "test",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VideoInfo{
				Title:           tt.fields.Title,
				Date:            tt.fields.Date,
				Target:          tt.fields.Target,
				Provider:        tt.fields.Provider,
				FileName:        tt.fields.FileName,
				FilePath:        tt.fields.FilePath,
				StreamingLink:   tt.fields.StreamingLink,
				UsersConfig:     tt.fields.UsersConfig,
				CQBotMsg:        tt.fields.CQBotMsg,
				TransRecordPath: tt.fields.TransRecordPath,
			}
			v.CreateNoticeMsg()
		})
	}
}
