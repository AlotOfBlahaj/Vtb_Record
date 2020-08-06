package interfaces

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/sirupsen/logrus"
)

type VideoInfoLogHook struct {
}

func (h *VideoInfoLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *VideoInfoLogHook) Fire(entry *logrus.Entry) error {
	_ret, ok := entry.Data["video"]
	if !ok {
		return nil
	}
	v, ok := _ret.(*VideoInfo)
	if !ok {
		return nil
	}
	entry.Data["video"] = fmt.Sprintf("%s|%s|%s", v.Provider, v.UsersConfig.Name, v.Title)
	return nil
}

func init() {
	logrus.AddHook(&VideoInfoLogHook{})
}

type VideoInfo struct {
	Title    string
	Date     string
	Target   string
	Provider string
	FileName string
	FilePath string
	//Target          string
	UsersConfig     config.UsersConfig
	TransRecordPath string
}

type LiveStatus struct {
	IsLive bool
	Video  *VideoInfo
}
