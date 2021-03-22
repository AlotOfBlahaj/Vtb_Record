package main

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"github.com/fzxiao233/Vtb_Record/live/plugins"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var SafeStop bool

func initPluginManager() videoworker.PluginManager {
	pm := videoworker.PluginManager{}
	pm.AddPlugin(&plugins.PluginCQBot{})
	return pm
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
			if config.ConfigChanged {
				allDone := true
				if allDone {
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
			}
			<-ticker.C
		}

	}()
	for _, dir := range config.Config.DownloadDir {
		utils.MakeDir(dir)
	}

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
		}
		time.Sleep(time.Duration(config.Config.NormalCheckSec) * time.Second)

		if SafeStop {
			break
		}
	}
	for {
		living := make([]string, 0, 128)
		statusMx.Lock()
		for _, mod := range status {
			for name, val := range mod {
				if val {
					living = append(living, name)
				}
			}
		}
		statusMx.Unlock()
		if len(living) == 0 {
			break
		}
		log.Infof("Waiting to finish: current living %s", living)
		time.Sleep(time.Second * 5)
	}
	log.Infof("All tasks finished! Wait an additional time to ensure everything's saved")
	time.Sleep(time.Second * 300)
	log.Infof("Everything finished, exiting now~~")
}

func handleInterrupt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Warnf("Ctrl+C pressed in Terminal!")
		time.Sleep(5 * time.Second) // wait rclone upload finish..
		os.Exit(0)
	}()
}

func handleUpdate() {
	c := make(chan os.Signal)
	SIGUSR1 := syscall.Signal(10)
	signal.Notify(c, SIGUSR1)
	go func() {
		<-c
		log.Warnf("Received update signal! Waiting everything done!")
		SafeStop = true
	}()
}

func main() {
	handleInterrupt()
	handleUpdate()
	rand.Seed(time.Now().UnixNano())

	http.DefaultClient.Transport = http.DefaultTransport
	config.PrepareConfig()
	config.InitLog()
	go config.InitProfiling()
	arrangeTask()
}
