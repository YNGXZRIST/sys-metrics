package metrics

import (
	m "sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/memstorage"
)

var (
	CounterStorage = memstorage.NewMemStorage[string, *m.Counter]()
	GaugeStorage   = memstorage.NewMemStorage[string, *m.Gauge]()
)

func StorageFactory(t string) (any, error) {
	switch t {
	case "counter":
		return CounterStorage, nil
	case "gauge":
		return GaugeStorage, nil
	default:
		return nil, ErrUnknownMetricType
	}
}
