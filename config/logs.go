package config

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/utils"
	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
)

// WriterHook is a hook that writes logs of specified LogLevels to specified Writer
type WriterHook struct {
	Out       io.Writer
	Formatter logrus.Formatter
	LogLevel  logrus.Level
}

// Fire will be called when some logging function is called with current hook
// It will format logrus entry to string and write it to appropriate writer
func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	if entry.Level > hook.LogLevel {
		return nil
	}
	serialized, err := hook.Formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to obtain reader, %v\n", err)
		return err
	}
	if _, err = hook.Out.Write(serialized); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to logrus, %v\n", err)
	}
	return nil
}

// Levels define on which logrus levels this hook would trigger
func (hook *WriterHook) Levels() []logrus.Level {
	//return logrus.AllLevels[:hook.LogLevel+1]
	return logrus.AllLevels[:logrus.DebugLevel+1]
}

var ConsoleHook *WriterHook
var FileHook *lumberjackrus.Hook

// Can't be func init as we need the parsed config
func InitLog() {
	var err error

	logrus.Printf("Init logging!")
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	// Log as JSON instead of the default ASCII formatter.
	formatter := &logrus.TextFormatter{
		ForceColors: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			_, _, shortfname := utils.RPartition(f.Function, ".")
			return fmt.Sprintf("%s()", shortfname), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	}
	logrus.SetFormatter(formatter)

	ConsoleHook = &WriterHook{ // Send logs with level higher than warning to stderr
		Out:       logrus.StandardLogger().Out,
		Formatter: formatter,
		LogLevel:  logrus.InfoLevel,
	}
	logrus.AddHook(ConsoleHook)
	logrus.StandardLogger().Out = ioutil.Discard

	FileHook, err = lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   Config.LogFile,
			MaxSize:    Config.LogFileSize,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  false,
		},
		logrus.DebugLevel,
		&logrus.JSONFormatter{},
		nil,
	)

	if err != nil {
		panic(fmt.Errorf("NewHook Error: %s", err))
	}

	logrus.AddHook(FileHook)

	UpdateLogLevel()
}
