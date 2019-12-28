package main

import (
	. "Vtb_Record/src/plugins"
	. "Vtb_Record/src/utils"
	"fmt"
	"time"
)

type ScheduleTask func(UsersConfig)

func RunScheduleTask(userConfig UsersConfig, task ScheduleTask) {
	ticker := time.NewTicker(time.Second * time.Duration(Config.CheckSec))
	go func() {
		for {
			task(userConfig)
			<-ticker.C
		}
	}()
}
func logUp(moduleName, TargetId string) {
	fmt.Printf("%s: %s up\n", moduleName, TargetId)
}
func arrangeTask() {
	var ch chan int
	for _, module := range Config.Module {
		if module.Enable {
			for _, Users := range module.Users {
				switch module.Name {
				case "Youtube":
					logUp(module.Name, Users.TargetId)
					go RunScheduleTask(Users, YoutubeCheckLive)
				case "Twitcasting":
					logUp(module.Name, Users.TargetId)
					go RunScheduleTask(Users, TwitcastingCheckLive)
				}
			}
		}
	}
	<-ch
}
func main() {
	arrangeTask()
}
