package downloader

import (
	"github.com/fzxiao233/Vtb_Record/downloader/base"
	"github.com/fzxiao233/Vtb_Record/downloader/native"
	"github.com/fzxiao233/Vtb_Record/downloader/streamlink"
	log "github.com/sirupsen/logrus"
)

func GetDownloader(providerName string) *base.Downloader {
	switch providerName {
	case "streamlink":
		return &base.Downloader{&streamlink.Streamlink{}}
	case "native":
		return &base.Downloader{&native.Native{}}
	default:
		log.Fatalf("Unknown download provider %s", providerName)
		return nil
	}
}
