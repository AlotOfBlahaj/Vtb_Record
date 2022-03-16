package base

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/interfaces"
)

type Mock struct {
	Video  *interfaces.VideoInfo
	IsLive bool
}

func (m *Mock) CheckLive(usersConfig config.UsersConfig) bool {
	return m.IsLive
}
func (m *Mock) CreateVideo(usersConfig config.UsersConfig) *interfaces.VideoInfo {
	return m.Video
}
