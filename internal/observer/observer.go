package observer

import (
	"context"
	"fmt"
	"sys-metrics/pkg/workerpool"
)

type observer interface {
	OnNotify(data any)
	Register(data any) error
}
type MetricObserverConfig struct {
	FilePath  string
	URL       string
	RateLimit int
}
type MetricsObserver struct {
	cfg              MetricObserverConfig
	ReportWorkerPool *workerpool.Pool
}

type MetricsEvent struct {
	Ts      int64
	IP      string
	Metrics []string
}
type MetricObserverOptions func(*MetricsObserver)

func NewMetricsObserver(cfg MetricObserverConfig) *MetricsObserver {
	obs := &MetricsObserver{
		cfg,
		nil,
	}
	return obs
}
func (o *MetricsObserver) Register(data any) error {
	ctx, ok := data.(context.Context)
	if !ok {
		return fmt.Errorf("observer: Register observer: invalid data type")

	}
	reportPool := workerpool.NewPool(o.cfg.RateLimit)
	reportPool.StartBg(ctx)
	o.ReportWorkerPool = reportPool
	return nil
}
func (o *MetricsObserver) OnNotify(data any) {
	if o.ReportWorkerPool == nil {
		//return fmt.Errorf("observer: OnNotify observer: register observer")
		return
	}
	event, ok := data.(MetricsEvent)
	if !ok {
		//fmt.Errorf("observer: OnNotify observer: invalid data type")
		return
	}
	task := o.ReportTask(event)
	o.ReportWorkerPool.Add(task)
}
func (o *MetricsObserver) ReportTask(event MetricsEvent) *workerpool.Task {
	task := workerpool.NewTask(func(a any) (any, error) {
		fmt.Println("here", event)
		return nil, nil
	})
	return task
}
