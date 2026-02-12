package memory

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
	"testing"
)

func TestService_GetAllMetrics(t *testing.T) {

	counter := metrics.NewCounter("counter1")
	counter.SetValue(10)
	gauge := metrics.NewGauge("gauge1")
	gauge.SetValue(2.5)
	svc := NewService()
	err := svc.Counters().Set(context.TODO(), counter.ID, counter)
	if err != nil {
		t.Errorf("NewService().Counters().Set(counter.ID, counter): got %v, want nil", err)
	}
	err = svc.Gauges().Set(context.TODO(), gauge.ID, gauge)
	if err != nil {
		t.Errorf("NewService().Gauges().Set(gauge.ID, gauge): got %v, want nil", err)
	}
	metricsList := svc.GetAllMetrics(context.TODO())

	if len(metricsList) != 2 {
		t.Errorf("GetAllMetrics() returned %d metrics, want 2", len(metricsList))
	}
}

func TestService_Gauges(t *testing.T) {
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	gauge := metrics.NewGauge("gauge1")
	gauge.SetValue(1.23)
	_ = gauges.Set(context.TODO(), gauge.ID, gauge)
	svc := NewService()
	err := svc.Gauges().Set(context.TODO(), gauge.ID, gauge)
	if err != nil {
		t.Error("Gauges.Set() should not return an error %w", err)
	}
	result, _ := svc.Gauges().Get(context.TODO(), gauge.ID)
	if result == nil || result.Value == nil || *result.Value != 1.23 {
		t.Errorf("Gauges() returned %v, want 1.23", result)
	}
}

func TestService_Counters(t *testing.T) {
	counter := metrics.NewCounter("counter1")
	counter.SetValue(42)
	svc := NewService()
	err := svc.Counters().Set(context.TODO(), counter.ID, counter)
	if err != nil {
		t.Error("Counters.Set() should not return an error %w", err)
	}
	result, _ := svc.Counters().Get(context.TODO(), counter.ID)
	if result == nil || result.Delta == nil || *result.Delta != 42 {
		t.Errorf("Counters() returned %v, want 42", result)
	}
}

func TestService_InitRoutine(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	if err := svc.InitRoutine(ctx); err != nil {
		t.Errorf("InitRoutine() error = %v", err)
	}
	ctx.Done()
}

func TestService_ReadBackup_WriteBackup(t *testing.T) {
	svc := NewService()
	if err := svc.ReadBackup(context.TODO()); err != nil {
		t.Errorf("ReadBackup() error = %v", err)
	}
	if err := svc.WriteBackup(context.TODO()); err != nil {
		t.Errorf("WriteBackup() error = %v", err)
	}
}
