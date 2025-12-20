package metrics

import (
	"sys-metrics/pkg/memstorage"
	"sys-metrics/pkg/metrics/counter"
	"sys-metrics/pkg/metrics/gauge"
)

var (
	CounterStorage = memstorage.NewMemStorage[string, *counter.Metric]()
	GaugeStorage   = memstorage.NewMemStorage[string, *gauge.Metric]()
)
