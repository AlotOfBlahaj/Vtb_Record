package plugins

import (
	"Vtb_Record/src/utils"
	"fmt"
)

type VideoMonitor interface {
	CheckLive(userConfig utils.UsersConfig)
}

func NoLiving(Provide string, Name string) {
	fmt.Printf("%s|%s|is not living\n", Provide, Name)
}
