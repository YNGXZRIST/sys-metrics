package agent

import (
	"context"
	"fmt"
	"sync"
	"sys-metrics/internal/config/agent"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/errors/timeerrors"
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
	return &Agent{cfg, sync.Mutex{}, NewCollector(), NewReporter(cfg.ServerAddr, cfg.Logger, cfg.Authenticator)}
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
	err := a.reporter.sendMetricsToServer(a.collector)
	if err != nil {
		return timeerrors.NewTimeError(labelerrors.NewLabelError("REPORTER", fmt.Errorf("reporter send error: %w", err)))
	}
	a.collector.ResetPollMetric()
	return nil
}
