package repository

import "sys-metrics/internal/model/metrics"

type MetricsBackupHandler interface {
	Read() ([]metrics.Metrics, error)
	Write(metric *metrics.Metrics) error
	WriteBatch(metrics []metrics.Metrics) error
	Upsert(metric *metrics.Metrics) error
	UpsertBatch(metrics []metrics.Metrics) error
}
