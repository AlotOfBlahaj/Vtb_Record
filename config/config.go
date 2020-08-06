package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/rclone/rclone/fs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"os"
	"reflect"
	"strings"
)

var Config *MainConfig
var ConfigChanged bool

type UsersConfig struct {
	TargetId     string
	Name         string
	DownloadDir  string
	NeedDownload bool
	TransBiliId  string
	UserHeaders  map[string]string
	ExtraConfig  map[string]interface{}
}
type ModuleConfig struct {
	//EnableProxy     bool
	//Proxy           string
	Name             string
	Enable           bool
	Users            []UsersConfig
	DownloadProvider string
	ExtraConfig      map[string]interface{}
}
type MainConfig struct {
	CriticalCheckSec int
	NormalCheckSec   int
	LogFile          string
	LogFileSize      int
	LogLevel         string
	RLogLevel        string
	DownloadQuality  string
	DownloadDir      []string
	UploadDir        string
	Module           []ModuleConfig
	PprofHost        string
	RedisHost        string
	ExpressPort      string
	EnableTS2MP4     bool
	ExtraConfig      map[string]interface{}
}

func InitConfig() {
	log.Print("Init config!")
	initConfig()
	log.Print("Load config!")
	_, _ = ReloadConfig()
	//fmt.Println(Config)
}

func initConfig() {
	/*viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")
	viper.SetConfigType("json")*/
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("config file error: %s\n", err)
		os.Exit(1)
	}

	ConfigChanged = true
	viper.OnConfigChange(func(in fsnotify.Event) {
		ConfigChanged = true
	})
}

func ReloadConfig() (bool, error) {
	if !ConfigChanged {
		return false, nil
	}
	ConfigChanged = false
	err := viper.ReadInConfig()
	if err != nil {
		return true, err
	}
	config := &MainConfig{}
	mdMap := make(map[string]*mapstructure.Metadata, 10)
	mdMap[""] = &mapstructure.Metadata{}
	err = viper.Unmarshal(config, func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			func(inType reflect.Type, outType reflect.Type, input interface{}) (interface{}, error) {
				if inType.Kind() == reflect.Map && outType.Kind() == reflect.Struct { // we'll decoding a struct
					fieldsMap := make(map[string]reflect.StructField, 10)
					for i := 0; i < outType.NumField(); i++ {
						fieldsMap[strings.ToLower(outType.Field(i).Name)] = outType.Field(i)
					}
					inputMap := input.(map[string]interface{})
					extraConfig := make(map[string]interface{}, 5)
					inputMap["ExtraConfig"] = extraConfig
					for key := range inputMap {
						_, ok := fieldsMap[strings.ToLower(key)]
						if !ok {
							extraConfig[key] = inputMap[key]
						}
					}
				}
				return input, nil
			},
			c.DecodeHook)
	})
	if err != nil {
		fmt.Println("Struct config error")
	}
	/*modules := viper.AllSettings()["module"].([]interface{})
	for i := 0; i < len(modules); i++ {
		Config.Module[i].ExtraConfig = modules[i].(map[string]interface{})
	}*/
	Config = config

	fs.Config.LogLevel = fs.LogLevelInfo
	if Config.RLogLevel == "debug" {
		fs.Config.LogLevel = fs.LogLevelDebug
	} else if Config.RLogLevel == "info" {
		fs.Config.LogLevel = fs.LogLevelInfo
	} else if Config.RLogLevel == "warn" {
		fs.Config.LogLevel = fs.LogLevelWarning
	} else if Config.RLogLevel == "error" {
		fs.Config.LogLevel = fs.LogLevelError
	}
	log.Printf("Set rclone log level to %s", fs.Config.LogLevel)

	level := logrus.InfoLevel
	if Config.LogLevel == "debug" {
		level = logrus.DebugLevel
	} else if Config.LogLevel == "info" {
		level = logrus.InfoLevel
	} else if Config.LogLevel == "warn" {
		level = logrus.WarnLevel
	} else if Config.LogLevel == "error" {
		level = logrus.ErrorLevel
	}

	logrus.SetLevel(level)
	log.Printf("Set log level to %s", level)
	return true, nil
}
