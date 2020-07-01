package main

import (
	"github.com/fzxiao233/Vtb_Record/live"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	. "github.com/fzxiao233/Vtb_Record/utils"
	"log"
	"time"
)

func arrangeTask() {
	//var ch chan int
	chans := make([][]bool, len(Config.Module))
	for i, module := range Config.Module {
		chans[i] = make([]bool, len(module.Users))
		for j, _ := range chans[i] {
			chans[i][j] = false
		}
	}

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(1))
		for {
			if ConfigChanged {
				allDone := true
				for mod_i, _ := range chans {
					for _, ch := range chans[mod_i] {
						if ch != false {
							allDone = false
						}
					}
				}
				if allDone {
					ret, err := ReloadConfig()
					if ret {
						if err == nil {
							log.Println("Config changed! New config: %s", Config)
						} else {
							log.Println("Config changed but loading failed: %s", err)
						}
					}
				}
			}
			<-ticker.C
		}

	}()

	for {
		for mod_i, module := range Config.Module {
			if module.Enable {
				for user_i, usersConfig := range module.Users {
					if chans[mod_i][user_i] != false {
						continue
					}
					chans[mod_i][user_i] = true //make(chan int)
					//log.Printf("%s|%s is up", module.Name, usersConfig.Name)
					go func(i, j int, mon monitor.VideoMonitor, userCon UsersConfig) {
						live.StartMonitor(mon, userCon)
						//chans[mod_i][user_i] <- 1
						chans[i][j] = false
					}(mod_i, user_i, monitor.CreateVideoMonitor(module), usersConfig)
				}
			}
		}
		time.Sleep(time.Duration(Config.CheckSec) * time.Second)
		for mod_i, _ := range chans {
			for _, ch := range chans[mod_i] {
				if ch != false {
					/*select {
					case <-ch:
						chans[mod_i][user_i] = nil
					default:

					}*/
				}
			}
		}
	}
	//<-ch
}
func main() {
	arrangeTask()
}
