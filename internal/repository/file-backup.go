package repository

import (
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
)

type FileBackupStorage struct {
	counters       *storage.MemStorage[string, *metrics.Counter]
	gauges         *storage.MemStorage[string, *metrics.Gauge]
	config         BackupConfig
	metricsHandler MetricsBackupHandler
}
