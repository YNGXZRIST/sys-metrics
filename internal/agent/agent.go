package agent

import (
	"sync"
	"sys-metrics/internal/config/agent"
	"time"
)

type Agent struct {
	*agent.Config
	mu        sync.RWMutex
	collector *Collector
}

func NewAgent(cfg *agent.Config) *Agent {
	return &Agent{cfg, sync.RWMutex{}, NewCollector()}
}
func (a *Agent) StartReport() {
	ticker := time.NewTicker(a.ReportInterval * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := a.Report()
			if err != nil {
				a.Logger.Println(err)
				return
			}
			break
		}
	}
}
func (a *Agent) StartPool() {
	ticker := time.NewTicker(a.PoolInterval * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.mu.Lock()
			a.collector.Update()
			a.collector.SetPoolMetric()
			a.collector.SetRandomValueMetric()
			a.mu.Unlock()
			break
		}
	}
}
func (a *Agent) Report() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	reporter := NewReporter(a.Config.ServerAddr, a.Logger)
	err := reporter.Send(*a.collector)
	if err != nil {
		return err
	}
	a.collector.ResetPoolMetric()
	return nil
}
