package monitor

import (
	"github.com/fzxiao233/Vtb_Record/plugins/structUtils"
	"github.com/fzxiao233/Vtb_Record/utils"
	"log"
)

func CreateVideoMonitor(moduleName string) VideoMonitor {
	switch moduleName {
	case "Youtube":
		return &Youtube{}
	case "Twitcasting":
		return &Twitcasting{}
	case "Bilibili":
		return &Bilibili{}
	default:
		return nil
	}
}

type VideoMonitor interface {
	CheckLive(usersConfig utils.UsersConfig) bool
	CreateVideo(usersConfig utils.UsersConfig) *structUtils.VideoInfo
}

func NoLiving(Provide string, Name string) {
	log.Printf("%s|%s|is not living\n", Provide, Name)
}
