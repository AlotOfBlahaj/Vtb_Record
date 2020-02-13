package monitor

import (
	"Vtb_Record/src/plugins/structUtils"
	"Vtb_Record/src/utils"
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
