package memory

import (
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
	"testing"
)

func TestService_GetAllMetrics(t *testing.T) {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()

	counter := metrics.NewCounter("counter1")
	counter.SetValue(10)
	gauge := metrics.NewGauge("gauge1")
	gauge.SetValue(2.5)

	_ = counters.Set(counter.ID, counter)
	_ = gauges.Set(gauge.ID, gauge)

	svc := NewService(counters, gauges)
	metricsList := svc.GetAllMetrics()
	if len(metricsList) != 2 {
		t.Errorf("GetAllMetrics() returned %d metrics, want 2", len(metricsList))
	}
}

func TestService_Gauges(t *testing.T) {
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	gauge := metrics.NewGauge("gauge1")
	gauge.SetValue(1.23)
	_ = gauges.Set(gauge.ID, gauge)
	svc := NewService(nil, gauges)
	result, _ := svc.Gauges().Get(gauge.ID)
	if result == nil || result.Value == nil || *result.Value != 1.23 {
		t.Errorf("Gauges() returned %v, want 1.23", result)
	}
}

func TestService_Counters(t *testing.T) {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	counter := metrics.NewCounter("counter1")
	counter.SetValue(42)
	_ = counters.Set(counter.ID, counter)
	svc := NewService(counters, nil)
	result, _ := svc.Counters().Get(counter.ID)
	if result == nil || result.Delta == nil || *result.Delta != 42 {
		t.Errorf("Counters() returned %v, want 42", result)
	}
}

func TestService_InitRoutine(t *testing.T) {
	svc := NewService(nil, nil)
	if err := svc.InitRoutine(nil); err != nil {
		t.Errorf("InitRoutine() error = %v", err)
	}
}

func TestService_ReadBackup_WriteBackup(t *testing.T) {
	svc := NewService(nil, nil)
	if err := svc.ReadBackup(); err != nil {
		t.Errorf("ReadBackup() error = %v", err)
	}
	if err := svc.WriteBackup(); err != nil {
		t.Errorf("WriteBackup() error = %v", err)
	}
}
