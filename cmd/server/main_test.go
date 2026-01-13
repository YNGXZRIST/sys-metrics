package main

import (
	model "sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
	"testing"
)

func Test_initStorage(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *model.Counter]()
	gauges := memstorage.NewMemStorage[string, *model.Gauge]()
	svc.Init(counters, gauges)
	if svc.Gauges() != gauges {
		t.Errorf("Gauges(): expected %v, got %v", gauges, svc.Gauges())
	}
	if svc.Counters() != counters {
		t.Errorf("Counters(): expected %v, got %v", counters, svc.Counters())
	}
}
