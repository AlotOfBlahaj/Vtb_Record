package main

import (
	"flag"
	"fmt"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	"github.com/fzxiao233/Vtb_Record/utils"
	"github.com/orandin/lumberjackrus"
	"github.com/rclone/rclone/fs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	_ "net/http/pprof"
	"path"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// Can't be func init as we need the parsed config
func initLog() {
	log.Printf("Init logging!")
	log.SetReportCaller(true)
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			_, _, shortfname := utils.RPartition(f.Function, ".")
			return fmt.Sprintf("%s()", shortfname), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	},
	)
	hook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   config.Config.LogFile,
			MaxSize:    config.Config.LogFileSize,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  false,
		},
		log.DebugLevel,
		&log.JSONFormatter{},
		nil,
	)

	if err != nil {
		panic(fmt.Errorf("NewHook Error: %s", err))
	}

	log.AddHook(hook)

	fs.LogPrint = func(level fs.LogLevel, text string) {
		log.WithField("src", "rclone").Infof(fmt.Sprintf("%-6s: %s", level, text))
	}
}

func arrangeTask() {
	log.Printf("Arrange tasks...")
	status := make([]map[string]bool, len(config.Config.Module))
	for i, module := range config.Config.Module {
		status[i] = make(map[string]bool, len(module.Users))
		/*for j, _ := range status[i] {
			status[i][j] = false
		}*/
	}

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(1))
		for {
			if config.ConfigChanged {
				allDone := true
				/*for mod_i, _ := range status {
					for _, ch := range status[mod_i] {
						if ch != false {
							allDone = false
						}
					}
				}*/
				if allDone {
					ret, err := config.ReloadConfig()
					if ret {
						if err == nil {
							log.Infof("Config changed! New config: %s", config.Config)
						} else {
							log.Warnf("Config changed but loading failed: %s", err)
						}
					}
				}
			}
			<-ticker.C
		}

	}()

	defer func() {
		panic("arrangeTask goes out!!!")
	}()

	var statusMx sync.Mutex
	for {
		living := make([]string, 0, 128)
		changed := make([]string, 0, 128)
		for mod_i, module := range config.Config.Module {
			if module.Enable {
				for _, usersConfig := range module.Users {
					statusMx.Lock()
					if status[mod_i][usersConfig.Name] != false {
						living = append(living, fmt.Sprintf("\"%s-%s\"", usersConfig.Name, usersConfig.TargetId))
						statusMx.Unlock()
						continue
					}
					status[mod_i][usersConfig.Name] = true
					statusMx.Unlock()
					changed = append(changed, usersConfig.Name)
					go func(i int, j string, mon monitor.VideoMonitor, userCon config.UsersConfig) {
						live.StartMonitor(mon, userCon)
						statusMx.Lock()
						status[i][j] = false
						statusMx.Unlock()
					}(mod_i, usersConfig.Name, monitor.CreateVideoMonitor(module), usersConfig)
					time.Sleep(time.Millisecond * 20)
				}
			}
		}
		log.Infof("current living %s", living)
		log.Debugf("checked %s", changed)
		if time.Now().Minute() > 55 || time.Now().Minute() < 5 || (time.Now().Minute() > 25 && time.Now().Minute() < 35) {
			time.Sleep(time.Duration(config.Config.CriticalCheckSec) * time.Second)
		}
		time.Sleep(time.Duration(config.Config.NormalCheckSec) * time.Second)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	log.Warnf("Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tGoroutines = %v\tNumGC = %v",
		bToMb(m.Alloc),
		bToMb(m.TotalAlloc),
		bToMb(m.Sys),
		runtime.NumGoroutine(),
		m.NumGC)
}

func printMem() {
	log.Warnf("Starting pprof server")
	//go http.ListenAndServe("0.0.0.0:49314", nil)
	go http.ListenAndServe(config.Config.PprofHost, nil)

	go func() {
		ticker := time.NewTicker(time.Minute * 1)
		for {
			PrintMemUsage()
			<-ticker.C
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Millisecond * 600)
		for {
			//start := time.Now()
			runtime.GC()
			//log.Debugf("GC & scvg use %s", time.Now().Sub(start))
			<-ticker.C
		}
	}()

	ticker := time.NewTicker(time.Second * 5)
	for {
		start := time.Now()
		debug.FreeOSMemory()
		log.Debugf("scvg use %s", time.Now().Sub(start))
		<-ticker.C
	}
}

func main() {
	http.DefaultTransport = &http.Transport{
		DisableKeepAlives:  true, // disable keep alive to avoid connection reset
		DisableCompression: false,
	}
	http.DefaultClient.Transport = http.DefaultTransport
	fs.Config.Transfers = 20
	//fs.Config.ConnectTimeout = time.Second * 4
	//fs.Config.Timeout = time.Second * 8
	//fs.Config.TPSLimit = 0
	//fs.Config.NoGzip = false
	confPath := flag.String("config", "config.json", "config.json location")
	flag.Parse()
	viper.SetConfigFile(*confPath)
	config.InitConfig()
	initLog()
	go printMem()
	arrangeTask()
}
