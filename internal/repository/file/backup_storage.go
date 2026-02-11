package file

import (
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/pkg/storage"
)

type BackupStorage struct {
	counters       *storage.MemStorage[string, *metrics.Counter]
	gauges         *storage.MemStorage[string, *metrics.Gauge]
	config         metricsiface.BackupConfig
	metricsHandler metricsiface.Handler
}

func NewFileBackupStorage(config metricsiface.BackupConfig, metricsHandler metricsiface.Handler) *BackupStorage {
	return &BackupStorage{
		counters:       storage.NewMemStorage[string, *metrics.Counter](),
		gauges:         storage.NewMemStorage[string, *metrics.Gauge](),
		config:         config,
		metricsHandler: metricsHandler,
	}
}

func (s *BackupStorage) Counters() metricsiface.BackupMetricStorage[*metrics.Counter] {
	return &counterBackupStorage{
		MemStorage:     s.counters,
		config:         s.config,
		metricsHandler: s.metricsHandler,
	}
}

func (s *BackupStorage) Gauges() metricsiface.BackupMetricStorage[*metrics.Gauge] {
	return &gaugeBackupStorage{
		MemStorage:     s.gauges,
		config:         s.config,
		metricsHandler: s.metricsHandler,
	}
}
