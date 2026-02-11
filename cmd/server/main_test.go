package main

import (
	model "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
	"sys-metrics/pkg/storage"
	"testing"
)

func Test_initStorage(t *testing.T) {
	counters := storage.NewMemStorage[string, *model.Counter]()
	gauges := storage.NewMemStorage[string, *model.Gauge]()
	svc.Init(memory.NewService(counters, gauges))
	if svc.Gauges() != gauges {
		t.Errorf("Gauges(): expected %v, got %v", gauges, svc.Gauges())
	}
	if svc.Counters() != counters {
		t.Errorf("Counters(): expected %v, got %v", counters, svc.Counters())
	}
}
