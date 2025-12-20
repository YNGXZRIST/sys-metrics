package storage_metrics

import (
	"sys-metrics/pkg/mem_storage"
	"sys-metrics/pkg/metrics"
	"sys-metrics/pkg/metrics/counter"
	"sys-metrics/pkg/metrics/gauge"
)

type Metric interface {
	*counter.Metric | *gauge.Metric
}

func StorageFactory(t string) (any, error) {
	switch t {
	case "counter":
		return metrics.CounterStorage, nil
	case "gauge":
		return metrics.GaugeStorage, nil
	default:
		return nil, metrics.ErrUnknownMetricType
	}
}
func UpdateStorageMetric[T Metric](s *mem_storage.MemStorage[string, T], name string, metric T) error {
	err := s.Set(name, metric)
	if err != nil {
		return err
	}

	return nil
}
