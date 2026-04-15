// Package metrics provides a global facade to the active ServiceInterface implementation and BackupStorage helper.
package metrics

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	s "sys-metrics/pkg/storage"
)

var defaultService metricsiface.ServiceInterface

// Init sets the default metrics service implementation and returns it.
func Init(service metricsiface.ServiceInterface) metricsiface.ServiceInterface {
	defaultService = service
	return service
}

// Counters returns the counter storage of the active service.
func Counters() metricsiface.MetricStorage[*metrics.Counter] {
	return defaultService.Counters()
}

// Gauges returns the gauge storage of the active service.
func Gauges() metricsiface.MetricStorage[*metrics.Gauge] {
	return defaultService.Gauges()
}

// GetService returns the current ServiceInterface implementation (after Init).
func GetService() metricsiface.ServiceInterface {
	return defaultService
}

// GetAllMetrics delegates to the active service.
func GetAllMetrics(ctx context.Context) []metrics.Metrics {
	return defaultService.GetAllMetrics(ctx)
}

// WriteBatchMetrics writes a batch of metrics through the active service.
func WriteBatchMetrics(ctx context.Context, metrics []metrics.Metrics) error {
	return defaultService.WriteBatchMetrics(ctx, metrics)
}

// BackupStorage holds in-memory gauges/counters and a Handler for DB/file synchronization.
type BackupStorage struct {
	counters       *s.MemStorage[string, *metrics.Counter]
	gauges         *s.MemStorage[string, *metrics.Gauge]
	config         metricsiface.BackupConfig
	metricsHandler metricsiface.Handler
}

// NewBackupStorage wraps in-memory stores and a persistence handler according to backup config.
func NewBackupStorage(config metricsiface.BackupConfig, metricsHandler metricsiface.Handler) *BackupStorage {
	return &BackupStorage{
		counters: s.NewMemStorage[string, *metrics.Counter](),
		gauges:   s.NewMemStorage[string, *metrics.Gauge](),
		config:   config, metricsHandler: metricsHandler}
}

// Counters returns a counter storage adapter that supports NeedSync.
func (s *BackupStorage) Counters() metricsiface.BackupMetricStorage[*metrics.Counter] {
	return &CounterBackupStorage{
		MemStorage:     s.counters,
		Config:         s.config,
		MetricsHandler: s.metricsHandler,
	}
}

// Gauges returns a gauge storage adapter that supports NeedSync.
func (s *BackupStorage) Gauges() metricsiface.BackupMetricStorage[*metrics.Gauge] {
	return &GaugeBackupStorage{
		MemStorage:     s.gauges,
		Config:         s.config,
		MetricsHandler: s.metricsHandler,
	}
}
