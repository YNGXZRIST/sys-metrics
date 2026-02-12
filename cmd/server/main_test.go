package main

import (
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
	"testing"
)

func Test_initStorage(t *testing.T) {

	svc.Init(memory.NewService())
	if svc.Gauges() == nil {
		t.Errorf("Gauges()==nil")
	}
	if svc.Counters() == nil {
		t.Errorf("Counters()==nil")
	}
}
