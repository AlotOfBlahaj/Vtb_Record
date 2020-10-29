package provgo

import (
	"crypto/tls"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/videoworker/downloader/stealth"
	log "github.com/sirupsen/logrus"
	"net/http"
	"testing"
	"time"
)

var STUB_URL = "http://127.0.0.1:8000/a.m3u8"

func TestMain(m *testing.M) {
	config.PrepareConfig()

	m.Run()
}

func CreateDownloader() *HLSDownloader {
	clients := []*http.Client{
		{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 20 * time.Second,
				TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
				//DisableCompression:    true,
				DisableKeepAlives: false,
				TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
				DialContext:       http.DefaultTransport.(*http.Transport).DialContext,
				DialTLS:           http.DefaultTransport.(*http.Transport).DialTLS,
			},
			Timeout: 60 * time.Second,
		},
	}

	downloader := &HLSDownloader{
		Logger:          log.WithField("test", true),
		AltAsMain:       false,
		HLSUrl:          STUB_URL,
		HLSHeader:       map[string]string{},
		AltHLSUrl:       STUB_URL,
		AltHLSHeader:    map[string]string{},
		Clients:         clients,
		AltClients:      []*http.Client{},
		Video:           &interfaces.VideoInfo{},
		OutPath:         "stub",
		Cookie:          "stub",
		M3U8UrlRewriter: stealth.GetRewriter(),
	}
	return downloader
}
