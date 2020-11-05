package base

import (
	"crypto/tls"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	"net/http"
	"net/url"
	"time"
)

type LiveInfo struct {
	Title         string
	StreamingLink string
}

type MonitorCtx struct {
	Client         *http.Client
	ExtraModConfig map[string]interface{}
}

// HttpGet wraps the raw HttpGet with monitor's global header
func (c *MonitorCtx) HttpGet(url string, header map[string]string) ([]byte, error) {
	finalHeaders := make(map[string]string, 10)
	for k, v := range c.GetHeaders() {
		finalHeaders[k] = v
	}
	for k, v := range header {
		finalHeaders[k] = v
	}
	return utils.HttpGet(c.Client, url, finalHeaders)
}

func (c *MonitorCtx) HttpPost(url string, header map[string]string, data []byte) ([]byte, error) {
	finalHeaders := make(map[string]string, 10)
	for k, v := range c.GetHeaders() {
		finalHeaders[k] = v
	}
	for k, v := range header {
		finalHeaders[k] = v
	}
	return utils.HttpPost(c.Client, url, finalHeaders, data)
}

type HeadersConfig struct {
	HttpHeaders map[string]string
}

func (c *MonitorCtx) GetHeaders() map[string]string {
	headerConfig := HeadersConfig{}
	_ = utils.MapToStruct(c.ExtraModConfig, &headerConfig)
	return headerConfig.HttpHeaders
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

type VideoMonitor interface {
	CheckLive(usersConfig config.UsersConfig) bool
	CreateVideo(usersConfig config.UsersConfig) *interfaces.VideoInfo
	GetCtx() *MonitorCtx
	DownloadProvider() string
}

type BaseMonitor struct {
	Ctx      MonitorCtx
	Provider string
}

func (b *BaseMonitor) CreateVideo(usersConfig config.UsersConfig) *interfaces.VideoInfo {
	return nil
}
func (b *BaseMonitor) CheckLive(usersConfig config.UsersConfig) bool {
	return false
}

func (b *BaseMonitor) GetCtx() *MonitorCtx {
	return &b.Ctx
}

func (b *BaseMonitor) DownloadProvider() string {
	return b.Provider
}

// monitorCtx contains mod's extraConfig and its own http client
func CreateMonitorCtx(module config.ModuleConfig) MonitorCtx {
	ctx := MonitorCtx{ExtraModConfig: module.ExtraConfig}
	var client *http.Client
	proxy, ok := ctx.GetProxy()
	if ok && proxy != "" {
		proxyUrl, _ := url.Parse("socks5://" + proxy)
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyUrl),
		}

		//adding the Transport object to the http Client
		client = &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		}
	} else {
		//client = http.DefaultClient
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 60 * time.Second,
		}
	}
	ctx.Client = client
	return ctx
}

// get mod config & ctx statically
func GetMod(modName string) *config.ModuleConfig {
	for _, m := range config.Config.Module {
		if m.Name == modName {
			return &m
		}
	}
	return nil
}

func GetCtx(modName string) *MonitorCtx {
	var ctx *MonitorCtx
	ret := GetMod(modName)
	if ret == nil {
		return nil
	}
	_ctx := CreateMonitorCtx(*ret)
	ctx = &_ctx
	return ctx
}

func NoLiving(Provide string, Name string) {
	//stdlog.Printf("%s|%s|is not living\r", Provide, Name)
}
