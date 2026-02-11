package metrics

import (
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
)

var defaultService metricsiface.ServiceInterface

type BackupDBService struct {
	backupStorage any
}

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

func GetAllMetrics() []metrics.Metrics {
	return defaultService.GetAllMetrics()
}
