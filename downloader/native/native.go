package native

import (
	"github.com/fzxiao233/Vtb_Record/downloader/base"
	"github.com/fzxiao233/Vtb_Record/interfaces"
)

type Native struct {
	base.Downloader
}

func (n Native) StartDownload(video *interfaces.VideoInfo, proxy string, cookie string, filepath string) error {
	panic("implement me")
}
