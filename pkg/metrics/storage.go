package metrics

import (
	"sys-metrics/pkg/mem_storage"
	"sys-metrics/pkg/metrics/counter"
	"sys-metrics/pkg/metrics/gauge"
)

var (
	CounterStorage = mem_storage.NewMemStorage[string, *counter.Metric]()
	GaugeStorage   = mem_storage.NewMemStorage[string, *gauge.Metric]()
)
