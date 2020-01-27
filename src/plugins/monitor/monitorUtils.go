package monitor

import (
	"Vtb_Record/src/plugins/structUtils"
	"Vtb_Record/src/utils"
	"log"
)

func CreateVideoMonitor(moduleName string) VideoMonitor {
	switch moduleName {
	case "Youtube":
		return &Youtube{}
	case "Twitcasting":
		return &Twitacasting{}
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
