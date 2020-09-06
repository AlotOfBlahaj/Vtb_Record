package videoworker

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

type PluginCallback interface {
	LiveStart(p *ProcessVideo) error
	DownloadStart(p *ProcessVideo) error
	LiveEnd(p *ProcessVideo) error
}

type PluginManager struct {
	plugins []PluginCallback
}

func (p *PluginManager) AddPlugin(plug PluginCallback) {
	p.plugins = append(p.plugins, plug)
}

func (p *PluginManager) OnLiveStart(video *ProcessVideo) {
	var wg sync.WaitGroup
	wg.Add(len(p.plugins))
	for _, plug := range p.plugins {
		go func(callback PluginCallback) {
			defer wg.Done()
			err := callback.LiveStart(video)
			if err != nil {
				video.getLogger().Errorf("plugin %s livestart error: %s", callback, err)
			}
		}(plug)
	}
	wg.Wait()
}

func (p *PluginManager) OnDownloadStart(video *ProcessVideo) {
	var wg sync.WaitGroup
	wg.Add(len(p.plugins))
	for _, plug := range p.plugins {
		go func(callback PluginCallback) {
			defer wg.Done()
			err := callback.LiveStart(video)
			if err != nil {
				video.getLogger().Errorf("plugin %s downloadstart error: %s", callback, err)
			}
		}(plug)
	}
	wg.Wait()
}

func (p *PluginManager) OnLiveEnd(video *ProcessVideo) {
	var wg sync.WaitGroup
	wg.Add(len(p.plugins))
	for _, plug := range p.plugins {
		go func(callback PluginCallback) {
			defer wg.Done()
			err := callback.LiveStart(video)
			if err != nil {
				log.Errorf("plugin %s liveend error: %s", callback, err)
			}
		}(plug)
	}
	wg.Wait()
}
