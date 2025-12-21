package metrics

import (
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/memstorage"
)

var (
	CounterStorage = memstorage.NewMemStorage[string, *metrics.Counter]()
	GaugeStorage   = memstorage.NewMemStorage[string, *metrics.Gauge]()
)
