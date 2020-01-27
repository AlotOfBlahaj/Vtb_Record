package structUtils

import "Vtb_Record/src/utils"

type VideoInfo struct {
	Title         string
	Date          string
	Target        string
	Provider      string
	FileName      string
	FilePath      string
	StreamingLink string
	UsersConfig   utils.UsersConfig
	CQBotMsg      string
}

func (v *VideoInfo) CreateLiveMsg() {
	v.CQBotMsg = "[直播提示]" + "[" + v.Provider + "]" + v.Title + "正在直播" + "链接: " + v.Target + " [CQ:at,qq=all]"
}

func (v *VideoInfo) CreateNoticeMsg() {
	v.CQBotMsg = "[" + v.Provider + "]" + v.Title + "链接: " + v.Target
}
