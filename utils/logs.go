package utils

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/knq/sdhook"
	"github.com/orandin/lumberjackrus"
	"github.com/rclone/rclone/fs"
	log "github.com/sirupsen/logrus"
	"path"
	"runtime"
)

// Can't be func init as we need the parsed config
func InitLog() {
	log.Printf("Init logging!")
	log.SetReportCaller(true)
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			_, _, shortfname := RPartition(f.Function, ".")
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

	h, err := sdhook.New(
		sdhook.GoogleLoggingAgent(),
		sdhook.LogName(config.Config.LogFile),
	)
	if err != nil {
		log.WithField("prof", true).Warnf("Failed to initialize the sdhook: %v", err)
	} else {
		log.AddHook(h)
	}

	fs.LogPrint = func(level fs.LogLevel, text string) {
		log.WithField("src", "rclone").Infof(fmt.Sprintf("%-6s: %s", level, text))
	}
}
