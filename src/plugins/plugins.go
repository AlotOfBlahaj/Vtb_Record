package plugins

import (
	"Vtb_Record/src/utils"
	"fmt"
)

type VideoMonitor interface {
	CheckLive(userConfig utils.UsersConfig) bool
}

func NoLiving(Provide string, Name string) {
	fmt.Printf("%s|%s|is not living\n", Provide, Name)
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
