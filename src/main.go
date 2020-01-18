package main

import (
	. "Vtb_Record/src/plugins"
	. "Vtb_Record/src/utils"
	"log"
	"time"
)

type ScheduleTask func(monitor VideoMonitor, userConfig UsersConfig)

func RunScheduleTask(userConfig UsersConfig, monitor VideoMonitor, task ScheduleTask) {
	ticker := time.NewTicker(time.Second * time.Duration(Config.CheckSec))
	go func() {
		for {
			task(monitor, userConfig)
			<-ticker.C
		}
	}()
}
func arrangeTask() {
	var ch chan int
	for _, module := range Config.Module {
		if module.Enable {
			for _, usersConfig := range module.Users {
				log.Printf("%s|%s is up", module.Name, usersConfig.Name)
				go RunScheduleTask(usersConfig, CreateVideoMonitor(module.Name), StartMonitor)
			}
		}
	}
	<-ch
}
func main() {
	arrangeTask()
}
