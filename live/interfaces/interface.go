package interfaces

import (
	"github.com/fzxiao233/Vtb_Record/utils"
)

type VideoInfo struct {
	Title           string
	Date            string
	Target          string
	Provider        string
	FileName        string
	FilePath        string
	StreamingLink   string
	UsersConfig     utils.UsersConfig
	TransRecordPath string
}

type LiveStatus struct {
	IsLive bool
	Video  *VideoInfo
}
