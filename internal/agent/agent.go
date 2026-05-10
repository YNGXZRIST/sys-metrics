// Package agent implements the system metrics collection loop and reporting to the server.
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

// generate:reset

// Agent ties together config, metric collection, reporting, and a report worker pool.
type Agent struct {
	*agent.Config
	collector  *Collector
	reporter   *Reporter
	ReportPool *workerpool.Pool
	mu         sync.Mutex
}

// NewAgent creates an agent with a report pool and collectors sized by cfg.RateLimit.
func NewAgent(cfg *agent.Config, ctx context.Context) *Agent {
	reportPool := workerpool.NewPool(cfg.RateLimit)
	reportPool.StartBg(ctx)
	return &Agent{
		Config:     cfg,
		collector:  NewCollector(ctx, cfg.RateLimit),
		reporter:   NewReporter(cfg.ServerAddr, cfg.Logger, cfg.Authenticator),
		ReportPool: reportPool,
		mu:         sync.Mutex{},
	}
}

// StartReport calls Report on every ReportInterval tick until the context is canceled.
func (a *Agent) StartReport(ctx context.Context) {
	err := a.Report(ctx)
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
			err := a.Report(ctx)
			if err != nil {
				a.Logger.Error("report error", zap.Error(err))
			}
		}
	}
}

// StartPoll refreshes system metrics and auxiliary gauges on every PollInterval tick.
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

// Report asynchronously sends buffered metrics to the server and resets the poll counter.
func (a *Agent) Report(ctx context.Context) error {
	task := workerpool.NewTask(func(x any) (any, error) {

		err := a.reporter.sendMetricsToServer(a.collector)
		if err != nil {
			return nil, timeerrors.NewTimeError(labelerrors.NewLabelError("REPORTER", fmt.Errorf("reporter send error: %w", err)))
		}
		a.collector.ResetPollMetric()
		return nil, nil
	})
	a.ReportPool.Add(task)
	res := a.ReportPool.Get(ctx)
	return res.Err
}
