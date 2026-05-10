// Package metricsiface defines storage-layer contracts for metrics (service, backup, DB/file handlers).
package metricsiface

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
)

// ServiceInterface describes the metrics service: gauge/counter access, backups, and batch writes.
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

// MetricStorage is a name-keyed metric store extending the base Storage contract.
type MetricStorage[V any] interface {
	storage.Storage[string, V]
}

// BackupMetricStorage adds a flag indicating whether sync with persistent storage is required.
type BackupMetricStorage[V any] interface {
	MetricStorage[V]
	NeedSync() bool
}

// BackupConfig reports whether periodic backup synchronization is enabled.
type BackupConfig interface {
	NeedSync() bool
}

// Handler is low-level persistence access (file or PostgreSQL): read/write and upsert.
type Handler interface {
	Upsert(ctx context.Context, metric *metrics.Metrics) error
	Read(ctx context.Context) ([]metrics.Metrics, error)
	Write(ctx context.Context, metric *metrics.Metrics) error
	WriteBatch(ctx context.Context, metrics []metrics.Metrics) error
}
