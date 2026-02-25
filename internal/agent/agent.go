package agent

import (
	"context"
	"fmt"
	"sync"
	"sys-metrics/internal/config/agent"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/errors/timeerrors"
	"sys-metrics/pkg/workerpool"
	"time"

	"go.uber.org/zap"
)

type Agent struct {
	*agent.Config
	mu         sync.Mutex
	collector  *Collector
	reporter   *Reporter
	ReportPool *workerpool.Pool
}

func NewAgent(cfg *agent.Config, ctx context.Context) *Agent {
	reportPool := workerpool.NewPool(cfg.RateLimit)
	reportPool.StartBg(ctx)
	return &Agent{cfg, sync.Mutex{}, NewCollector(ctx, cfg.RateLimit), NewReporter(cfg.ServerAddr, cfg.Logger, cfg.Authenticator), reportPool}
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
			err := a.collector.Update()
			if err != nil {
				a.Logger.Error("update collector error", zap.Error(err))
			}
			a.collector.SetPollCounterMetric()
			a.collector.SetRandomValueMetric()
			a.mu.Unlock()
		}
	}
}
func (a *Agent) Report() error {
	task := workerpool.NewTask(func(x any) (any, error) {

		err := a.reporter.sendMetricsToServer(a.collector)
		if err != nil {
			return nil, timeerrors.NewTimeError(labelerrors.NewLabelError("REPORTER", fmt.Errorf("reporter send error: %w", err)))
		}
		a.collector.ResetPollMetric()
		return nil, nil
	})
	a.ReportPool.Add(task)
	res := a.ReportPool.Get()
	return res.Err
}
