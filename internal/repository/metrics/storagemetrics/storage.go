package storagemetrics

import (
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
