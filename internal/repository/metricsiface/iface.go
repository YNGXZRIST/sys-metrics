package metricsiface

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
)

type ServiceInterface interface {
	GetAllMetrics(ctx context.Context) []metrics.Metrics
	Gauges() MetricStorage[*metrics.Gauge]
	Counters() MetricStorage[*metrics.Counter]
	ReadBackup(ctx context.Context) error
	WriteBackup(ctx context.Context) error
	InitRoutine(ctx context.Context) error
	Close(ctx context.Context) error
	WriteBatchMetrics(ctx context.Context, m []metrics.Metrics) error
}

type MetricStorage[V any] interface {
	storage.Storage[string, V]
}

type BackupMetricStorage[V any] interface {
	MetricStorage[V]
	NeedSync() bool
}

type BackupConfig interface {
	NeedSync() bool
}

type Handler interface {
	Upsert(ctx context.Context, metric *metrics.Metrics) error
	Read(ctx context.Context) ([]metrics.Metrics, error)
	Write(ctx context.Context, metric *metrics.Metrics) error
	WriteBatch(ctx context.Context, metrics []metrics.Metrics) error
}
