package utils

type VideoInfo struct {
	Title         string
	Date          string
	Target        string
	Provider      string
	FilePath      string
	StreamingLink string
	UsersConfig   UsersConfig
	CQBotMsg      string
}

func (v *VideoInfo) CreateLiveMsg() {
	v.CQBotMsg = "[直播提示]" + "[" + v.Provider + "]" + v.Title + "正在直播" + "链接:" + v.Target
}
