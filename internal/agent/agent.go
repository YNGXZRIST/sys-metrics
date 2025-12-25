package agent

import (
	"context"
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
func (a *Agent) StartReport(ctx context.Context) {
	ticker := time.NewTicker(a.ReportInterval * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := a.Report()
			if err != nil {
				a.Logger.Println(err)
				return
			}
		}
	}
}
func (a *Agent) StartPool(ctx context.Context) {
	ticker := time.NewTicker(a.PoolInterval * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.mu.Lock()
			a.collector.Update()
			a.collector.SetPoolCounterMetric()
			a.collector.SetRandomValueMetric()
			a.mu.Unlock()
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
