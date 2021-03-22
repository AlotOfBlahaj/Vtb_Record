package config

import (
	log "github.com/sirupsen/logrus"
	"runtime"
	"time"
)

func PrintMemUsage() {
	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	log.WithField("prof", true).Warnf("Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tGoroutines = %v\tNumGC = %v",
		bToMb(m.Alloc),
		bToMb(m.TotalAlloc),
		bToMb(m.Sys),
		runtime.NumGoroutine(),
		m.NumGC)
}

func InitProfiling() {
	go func() {
		ticker := time.NewTicker(time.Minute * 1)
		for {
			PrintMemUsage()
			<-ticker.C
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Millisecond * 3000)
		for {
			runtime.GC()
			<-ticker.C
		}
	}()
	ticker := time.NewTicker(time.Second * 3)
	for {
		start := time.Now()
		debug.FreeOSMemory()
		log.WithField("prof", true).Debugf("scvg use %s", time.Now().Sub(start))
		<-ticker.C
	}
}
