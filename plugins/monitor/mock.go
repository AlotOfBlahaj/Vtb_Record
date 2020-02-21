package monitor

import (
	"github.com/fzxiao233/Vtb_Record/plugins/structUtils"
	"github.com/fzxiao233/Vtb_Record/utils"
)

type Mock struct {
	Video  *structUtils.VideoInfo
	IsLive bool
}

func (m *Mock) CheckLive(usersConfig utils.UsersConfig) bool {
	return m.IsLive
}
func (m *Mock) CreateVideo(usersConfig utils.UsersConfig) *structUtils.VideoInfo {
	return m.Video
}
