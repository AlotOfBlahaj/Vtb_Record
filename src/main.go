package main

import (
	. "Vtb_Record/src/plugins"
	"Vtb_Record/src/plugins/monitor"
	. "Vtb_Record/src/utils"
	"log"
)

type ScheduleTask func(monitor monitor.VideoMonitor, userConfig UsersConfig)

func arrangeTask() {
	var ch chan int
	for _, module := range Config.Module {
		if module.Enable {
			for _, usersConfig := range module.Users {
				log.Printf("%s|%s is up", module.Name, usersConfig.Name)
				go StartMonitor(monitor.CreateVideoMonitor(module.Name), usersConfig)
			}
		}
	}
	<-ch
}
func main() {
	arrangeTask()
}
