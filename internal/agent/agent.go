package agent

import (
	"context"
	"sync"
	"sys-metrics/internal/config/agent"
	"time"

	"go.uber.org/zap"
)

type Agent struct {
	*agent.Config
	mu        sync.Mutex
	collector *Collector
	reporter  *Reporter
}

func NewAgent(cfg *agent.Config) *Agent {
	return &Agent{cfg, sync.Mutex{}, NewCollector(), NewReporter(cfg.ServerAddr, cfg.Logger)}
}
func (a *Agent) StartReport(ctx context.Context) {
	err := a.Report()
	if err != nil {
		a.Logger.Error("report error", zap.Error(err))
	}
	ticker := time.NewTicker(a.ReportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			err := a.Report()
			if err != nil {
				a.Logger.Error("report error", zap.Error(err))
			}
		}
	}
}
func (a *Agent) StartPoll(ctx context.Context) {
	ticker := time.NewTicker(a.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.mu.Lock()
			a.collector.Update()
			a.collector.SetPollCounterMetric()
			a.collector.SetRandomValueMetric()
			a.mu.Unlock()
		}
	}
}
func (a *Agent) Report() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	err := a.reporter.Send(a.collector)
	if err != nil {
		return err
	}
	a.collector.ResetPollMetric()
	return nil
}
