package metrics

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	s "sys-metrics/pkg/storage"
)

var defaultService metricsiface.ServiceInterface

func Init(service metricsiface.ServiceInterface) metricsiface.ServiceInterface {
	defaultService = service
	return service
}

func Counters() metricsiface.MetricStorage[*metrics.Counter] {
	return defaultService.Counters()
}

func Gauges() metricsiface.MetricStorage[*metrics.Gauge] {
	return defaultService.Gauges()
}

func GetService() metricsiface.ServiceInterface {
	return defaultService
}

func GetAllMetrics(ctx context.Context) []metrics.Metrics {
	return defaultService.GetAllMetrics(ctx)
}
func WriteBatchMetrics(ctx context.Context, metrics []metrics.Metrics) error {
	return defaultService.WriteBatchMetrics(ctx, metrics)
}

type BackupStorage struct {
	counters       *s.MemStorage[string, *metrics.Counter]
	gauges         *s.MemStorage[string, *metrics.Gauge]
	config         metricsiface.BackupConfig
	metricsHandler metricsiface.Handler
}

func NewBackupStorage(config metricsiface.BackupConfig, metricsHandler metricsiface.Handler) *BackupStorage {
	return &BackupStorage{
		counters: s.NewMemStorage[string, *metrics.Counter](),
		gauges:   s.NewMemStorage[string, *metrics.Gauge](),
		config:   config, metricsHandler: metricsHandler}
}

func (s *BackupStorage) Counters() metricsiface.BackupMetricStorage[*metrics.Counter] {
	return &CounterBackupStorage{
		MemStorage:     s.counters,
		Config:         s.config,
		MetricsHandler: s.metricsHandler,
	}
}

func (s *BackupStorage) Gauges() metricsiface.BackupMetricStorage[*metrics.Gauge] {
	return &GaugeBackupStorage{
		MemStorage:     s.gauges,
		Config:         s.config,
		MetricsHandler: s.metricsHandler,
	}
}
