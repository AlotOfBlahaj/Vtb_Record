package main

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"github.com/fzxiao233/Vtb_Record/live/plugins"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func initPluginManager() videoworker.PluginManager {
	pm := videoworker.PluginManager{}
	pm.AddPlugin(&plugins.PluginCQBot{})
	return pm
}

func checkStreamlink() {
	c := exec.Command("streamlink", "--version")
	output, _ := c.CombinedOutput()
	if !strings.Contains(string(output), "streamlink") {
		log.Fatal("Cannot find streamlink")
	}
}

func arrangeTask() {
	log.Printf("Arrange tasks...")
	pm := initPluginManager()
	status := make([]map[string]bool, len(config.Config.Module))
	for i, module := range config.Config.Module {
		status[i] = make(map[string]bool, len(module.Users))
	}

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(1))
		for {
			if config.Changed {
				time.Sleep(4 * time.Second) // wait to ensure the config is fully written
				ret, err := config.ReloadConfig()
				if ret {
					if err == nil {
						log.Infof("\n\n\t\tConfig changed and load successfully!\n\n")
					} else {
						log.Warnf("Config changed but loading failed: %s", err)
					}
				}
			}
			<-ticker.C
		}

	}()

	var statusMx sync.Mutex
	for {
		var mods []config.ModuleConfig
		living := make([]string, 0, 128)
		changed := make([]string, 0, 128)
		mods = make([]config.ModuleConfig, len(config.Config.Module))
		copy(mods, config.Config.Module)
		for mod_i, module := range mods {
			if module.Enable {
				for _, usersConfig := range module.Users {
					identifier := fmt.Sprintf("\"%s-%s\"", usersConfig.Name, usersConfig.TargetId)
					statusMx.Lock()
					if status[mod_i][identifier] != false {
						living = append(living, fmt.Sprintf("\"%s-%s\"", usersConfig.Name, usersConfig.TargetId))
						statusMx.Unlock()
						continue
					}
					status[mod_i][identifier] = true
					statusMx.Unlock()
					changed = append(changed, identifier)
					go func(i int, j string, mon monitor.VideoMonitor, userCon config.UsersConfig) {
						live.StartMonitor(mon, userCon, pm)
						statusMx.Lock()
						status[i][j] = false
						statusMx.Unlock()
					}(mod_i, identifier, monitor.CreateVideoMonitor(module), usersConfig)
					time.Sleep(time.Millisecond * 20)
				}
			}
		}
		log.Infof("current living %s", living)
		log.Tracef("checked %s", changed)
		if time.Now().Minute() > 55 || time.Now().Minute() < 5 || (time.Now().Minute() > 25 && time.Now().Minute() < 35) {
			time.Sleep(time.Duration(config.Config.CriticalCheckSec) * time.Second)
		} else {
			time.Sleep(time.Duration(config.Config.NormalCheckSec) * time.Second)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	config.PrepareConfig()
	config.InitLog()
	go config.InitProfiling()
	checkStreamlink()
	arrangeTask()
}
