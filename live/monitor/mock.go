package monitor

import (
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
)

type Mock struct {
	Video  *interfaces.VideoInfo
	IsLive bool
}

func (m *Mock) CheckLive(usersConfig utils.UsersConfig) bool {
	return m.IsLive
}
func (m *Mock) CreateVideo(usersConfig utils.UsersConfig) *interfaces.VideoInfo {
	return m.Video
}
