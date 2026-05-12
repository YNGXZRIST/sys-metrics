package metrics

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	"testing"
)

func TestInit(t *testing.T) {
	service := Init(memory.NewService())

	if service == nil {
		t.Fatal("Init() returned nil service")
	}
	if service.Counters() == nil {
		t.Error("Init() did not set counters correctly")
	}
	if service.Gauges() == nil {
		t.Error("Init() did not set gauges correctly")
	}

	if defaultService != service {
		t.Error("Init() did not set defaultService correctly")
	}
}

func TestCounters(t *testing.T) {

	Init(memory.NewService())

	got := Counters()

	if got == nil {
		t.Fatal("Counters() returned nil")
	}
}

func TestGauges(t *testing.T) {
	Init(memory.NewService())

	got := Gauges()

	if got == nil {
		t.Fatal("Gauges() returned nil")
	}
}

func TestGetService(t *testing.T) {
	service := Init(memory.NewService())

	got := GetService()

	if got == nil {
		t.Fatal("GetService() returned nil")
	}
	if got != service {
		t.Error("GetService() returned wrong service")
	}
}

func TestGetAllMetrics(t *testing.T) {
	svc := Init(memory.NewService())

	counter := metrics.NewCounter("test_counter")
	counter.SetValue(42)
	err := svc.Counters().Set(context.Background(), "test_counter", counter)
	if err != nil {
		t.Errorf("Counters() returned %v", err)
	}
	gauge := metrics.NewGauge("test_gauge")
	gauge.SetValue(3.14)
	err = svc.Gauges().Set(context.Background(), "test_gauge", gauge)
	if err != nil {
		t.Errorf("Gauges() returned %v", err)
	}
	got := GetAllMetrics(context.Background())

	if len(got) != 2 {
		t.Errorf("GetAllMetrics() returned %d metrics, expected 2", len(got))
	}
}

func TestGetAllMetrics_empty(t *testing.T) {
	Init(memory.NewService())
	got := GetAllMetrics(context.Background())
	if len(got) != 0 {
		t.Fatalf("want empty, got %d", len(got))
	}
}
