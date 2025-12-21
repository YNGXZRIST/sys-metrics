package storagemetrics

import (
	m "sys-metrics/internal/model/metrics"
	s "sys-metrics/internal/service/metrics"
)

type Metric interface {
	*m.Counter | *m.Gauge
}

func StorageFactory(t string) (any, error) {
	switch t {
	case "counter":
		return s.CounterStorage, nil
	case "gauge":
		return s.GaugeStorage, nil
	default:
		return nil, s.ErrUnknownMetricType
	}
}
