package monitor

import (
	. "github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	stdlog "log"
	"net/http"
	"net/url"
	"time"
)

type MonitorCtx struct {
	Client         *http.Client
	ExtraModConfig map[string]interface{}
}

func (c *MonitorCtx) HttpGet(url string, header map[string]string) ([]byte, error) {
	finalHeaders := make(map[string]string, 10)
	for k, v := range header {
		finalHeaders[k] = v
	}
	for k, v := range c.GetHeaders() {
		finalHeaders[k] = v
	}
	return utils.HttpGet(c.Client, url, finalHeaders)
}

type HeadersConfig struct {
	HttpHeaders map[string]string
}

func (c *MonitorCtx) GetHeaders() map[string]string {
	config := HeadersConfig{}
	utils.MapToStruct(c.ExtraModConfig, &config)
	return config.HttpHeaders
}

func (c *MonitorCtx) GetProxy() (string, bool) {
	enableProxy, ok1 := c.ExtraModConfig["EnableProxy"]
	proxy, ok2 := c.ExtraModConfig["Proxy"]
	if ok1 && ok2 && enableProxy == true {
		return proxy.(string), true
	} else {
		return "", false
	}
}

func createMonitorCtx(module utils.ModuleConfig) MonitorCtx {
	ctx := MonitorCtx{ExtraModConfig: module.ExtraConfig}
	var client *http.Client
	proxy, ok := ctx.GetProxy()
	if ok && proxy != "" {
		proxyUrl, _ := url.Parse("socks5://" + proxy)
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}

		//adding the Transport object to the http Client
		client = &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		}
	} else {
		client = http.DefaultClient
	}
	ctx.Client = client
	return ctx
}

type BaseMonitor struct {
	ctx MonitorCtx
}

func (b *BaseMonitor) CreateVideo(usersConfig utils.UsersConfig) *VideoInfo {
	return nil
}
func (b *BaseMonitor) CheckLive(usersConfig utils.UsersConfig) bool {
	return false
}

func (b *BaseMonitor) GetCtx() *MonitorCtx {
	return &b.ctx
}

type VideoMonitor interface {
	CheckLive(usersConfig utils.UsersConfig) bool
	CreateVideo(usersConfig utils.UsersConfig) *VideoInfo
	GetCtx() *MonitorCtx
}

type LiveTrace func(monitor VideoMonitor) *LiveStatus

func CreateVideoMonitor(module utils.ModuleConfig) VideoMonitor {
	var monitor VideoMonitor
	//var monitor *BaseMonitor
	ctx := createMonitorCtx(module)
	switch module.Name {
	case "Youtube":
		monitor = &Youtube{BaseMonitor: BaseMonitor{ctx}}
	case "Twitcasting":
		monitor = &Twitcasting{BaseMonitor: BaseMonitor{ctx}}
	case "Bilibili":
		monitor = &Bilibili{BaseMonitor: BaseMonitor{ctx}}
	default:
		return nil
	}
	return monitor
}

func NoLiving(Provide string, Name string) {
	stdlog.Printf("%s|%s|is not living\r", Provide, Name)
}
