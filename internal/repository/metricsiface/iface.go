package metricsiface

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
)

type ServiceInterface interface {
	GetAllMetrics() []metrics.Metrics
	Gauges() MetricStorage[*metrics.Gauge]
	Counters() MetricStorage[*metrics.Counter]
	ReadBackup() error
	WriteBackup() error
	InitRoutine(ctx context.Context) error
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
	Upsert(metric *metrics.Metrics) error
	Read() ([]metrics.Metrics, error)
	Write(metric *metrics.Metrics) error
	WriteBatch(metrics []metrics.Metrics) error
}
