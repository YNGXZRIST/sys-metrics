package metrics

import (
	"sys-metrics/internal/model/metrics"
)

type MetricStorage[V any] interface {
	Set(key string, value V) error
	Get(key string) (V, error)
	All() map[string]V
}

type Service struct {
	counters MetricStorage[*metrics.Counter]
	gauges   MetricStorage[*metrics.Gauge]
}

var defaultService *Service

func Init(
	counters MetricStorage[*metrics.Counter],
	gauges MetricStorage[*metrics.Gauge],
) {
	defaultService = &Service{counters: counters, gauges: gauges}
}

func Counters() MetricStorage[*metrics.Counter] {
	return defaultService.counters
}
func Gauges() MetricStorage[*metrics.Gauge] {
	return defaultService.gauges
}
