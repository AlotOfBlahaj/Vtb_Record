package main

import (
	"github.com/fzxiao233/Vtb_Record/live"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	. "github.com/fzxiao233/Vtb_Record/utils"
	"log"
)

func arrangeTask() {
	var ch chan int
	for _, module := range Config.Module {
		if module.Enable {
			for _, usersConfig := range module.Users {
				log.Printf("%s|%s is up", module.Name, usersConfig.Name)
				go live.StartMonitor(monitor.CreateVideoMonitor(module), usersConfig)
			}
		}
	}
	<-ch
}
func main() {
	arrangeTask()
}
