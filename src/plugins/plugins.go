package plugins

import (
	"Vtb_Record/src/utils"
	"log"
)

type VideoMonitor interface {
	CheckLive(usersConfig utils.UsersConfig) bool
	CreateVideo(usersConfig utils.UsersConfig) *utils.VideoInfo
}

func StartMonitor(monitor VideoMonitor, usersConfig utils.UsersConfig) {
	if monitor.CheckLive(usersConfig) {
		ProcessVideo(monitor.CreateVideo(usersConfig))
	}
}
func NoLiving(Provide string, Name string) {
	log.Printf("%s|%s|is not living\n", Provide, Name)
}
func CreateVideoMonitor(moduleName string) VideoMonitor {
	switch moduleName {
	case "Youtube":
		return &Youtube{}
	case "Twitcasting":
		return &Twitacasting{}
	default:
		return nil
	}
}
