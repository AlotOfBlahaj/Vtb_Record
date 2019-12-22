package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

var Config *MainConfig

type UsersConfig struct {
	TargetId    string
	Name        string
	Live        bool
	Article     bool
	DownloadDir string
}
type ModuleConfig struct {
	Name       string
	Enable     bool
	EnableTemp bool
	Users      []UsersConfig
}
type MainConfig struct {
	EnableProxy     bool
	Proxy           string
	CheckSec        int
	DownloadQuality string
	DownloadDir     string
	Module          []ModuleConfig
}

func init() {
	initConfig()
}

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("config file error: %s\n", err)
		os.Exit(1)
	}
	Config = &MainConfig{}
	err = viper.Unmarshal(Config)
	if err != nil {
		fmt.Println("Struct config error")
	}
}
