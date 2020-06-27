package videoworker

import (
	"log"
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
		go func() {
			defer wg.Done()
			err := plug.LiveStart(video)
			if err != nil {
				log.Print(err)
			}
		}()
	}
	wg.Wait()
}

func (p *PluginManager) OnDownloadStart(video *ProcessVideo) {
	var wg sync.WaitGroup
	wg.Add(len(p.plugins))
	for _, plug := range p.plugins {
		go func() {
			defer wg.Done()
			err := plug.DownloadStart(video)
			if err != nil {
				log.Print(err)
			}
		}()
	}
	wg.Wait()
}

func (p *PluginManager) OnLiveEnd(video *ProcessVideo) {
	var wg sync.WaitGroup
	wg.Add(len(p.plugins))
	for _, plug := range p.plugins {
		go func() {
			defer wg.Done()
			err := plug.LiveEnd(video)
			if err != nil {
				log.Print(err)
			}
		}()
	}
	wg.Wait()
}
