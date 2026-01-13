package metrics

import (
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/memstorage"
	"testing"
)

func TestInit(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *metrics.Counter]()
	gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()

	Init(counters, gauges)

	if defaultService == nil {
		t.Fatal("Init() did not initialize defaultService")
	}
	if defaultService.counters != counters {
		t.Error("Init() did not set counters correctly")
	}
	if defaultService.gauges != gauges {
		t.Error("Init() did not set gauges correctly")
	}
}

func TestCounters(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *metrics.Counter]()
	gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
	Init(counters, gauges)

	got := Counters()

	if got == nil {
		t.Fatal("Counters() returned nil")
	}
	if got != counters {
		t.Errorf("Counters() returned wrong storage")
	}
}

func TestGauges(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *metrics.Counter]()
	gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
	Init(counters, gauges)

	got := Gauges()

	if got == nil {
		t.Fatal("Gauges() returned nil")
	}
	if got != gauges {
		t.Errorf("Gauges() returned wrong storage")
	}
}
