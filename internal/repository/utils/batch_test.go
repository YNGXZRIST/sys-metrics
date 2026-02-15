package utils

import (
	"context"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
	"testing"
)

func newTestStorages() (gauges *storage.MemStorage[string, *models.Gauge], counters *storage.MemStorage[string, *models.Counter]) {
	return storage.NewMemStorage[string, *models.Gauge](), storage.NewMemStorage[string, *models.Counter]()
}

func TestApplyGauge(t *testing.T) {
	ctx := context.Background()
	gauges, _ := newTestStorages()
	v := models.Metrics{ID: "Alloc", MType: common.Gauge}
	val := 123.45
	v.Value = &val

	err := ApplyGauge(ctx, v, gauges)
	if err != nil {
		t.Fatalf("ApplyGauge() err = %v", err)
	}
	got, err := gauges.Get(ctx, v.ID)
	if err != nil {
		t.Fatalf("Gauges().Get() err = %v", err)
	}
	if got.Metrics.ID != v.ID || got.Metrics.Value == nil || *got.Metrics.Value != val {
		t.Errorf("Gauges().Get() = %+v, want ID=%s Value=%g", got.Metrics, v.ID, val)
	}
}

func TestApplyCounter_New(t *testing.T) {
	ctx := context.Background()
	_, counters := newTestStorages()
	delta := int64(10)
	v := models.Metrics{ID: "PollCount", MType: common.Counter, Delta: &delta}

	m, err := ApplyCounter(ctx, v, counters)
	if err != nil {
		t.Fatalf("ApplyCounter() err = %v", err)
	}
	if m.ID != v.ID || m.Delta == nil || *m.Delta != 10 {
		t.Errorf("ApplyCounter() returned %+v, want Delta=10", m)
	}
	got, _ := counters.Get(ctx, v.ID)
	if got == nil || got.Delta == nil || *got.Delta != 10 {
		t.Errorf("Counters().Get() = %+v, want Delta=10", got)
	}
}

func TestApplyCounter_Accumulate(t *testing.T) {
	ctx := context.Background()
	_, counters := newTestStorages()
	c := models.NewCounter("PollCount")
	c.SetValue(5)
	_ = counters.Set(ctx, c.ID, c)

	delta := int64(3)
	v := models.Metrics{ID: "PollCount", MType: common.Counter, Delta: &delta}

	m, err := ApplyCounter(ctx, v, counters)
	if err != nil {
		t.Fatalf("ApplyCounter() err = %v", err)
	}
	if m.Delta == nil || *m.Delta != 8 {
		t.Errorf("ApplyCounter() returned Delta = %v, want 8", m.Delta)
	}
	got, _ := counters.Get(ctx, v.ID)
	if got == nil || got.Delta == nil || *got.Delta != 8 {
		var d int64 = -1
		if got != nil && got.Delta != nil {
			d = *got.Delta
		}
		t.Errorf("Counters().Get() Delta = %d, want 8", d)
	}
}

func TestApplyBatchToStorages(t *testing.T) {
	ctx := context.Background()
	gauges, counters := newTestStorages()

	val1 := 1.0
	val2 := 2.0
	m := []models.Metrics{
		{ID: "g1", MType: common.Gauge, Value: &val1},
		{ID: "c1", MType: common.Counter, Delta: ptrInt64(1)},
		{ID: "c1", MType: common.Counter, Delta: ptrInt64(2)},
		{ID: "g2", MType: common.Gauge, Value: &val2},
	}

	byID, err := ApplyBatchToStorages(ctx, m, gauges, counters)
	if err != nil {
		t.Fatalf("ApplyBatchToStorages() err = %v", err)
	}
	if len(byID) != 3 {
		t.Errorf("byID len = %d, want 3 (g1, c1, g2)", len(byID))
	}
	if byID["g1"].Value == nil || *byID["g1"].Value != 1.0 {
		t.Errorf("g1 = %v", byID["g1"])
	}
	if byID["g2"].Value == nil || *byID["g2"].Value != 2.0 {
		t.Errorf("g2 = %v", byID["g2"])
	}
	if byID["c1"].Delta == nil || *byID["c1"].Delta != 3 {
		t.Errorf("c1 Delta = %v, want 3 (1+2)", byID["c1"].Delta)
	}
}

func TestApplyBatchToStorages_SkipsUnknownType(t *testing.T) {
	ctx := context.Background()
	gauges, counters := newTestStorages()
	m := []models.Metrics{
		{ID: "unknown", MType: "unknown"},
	}
	byID, err := ApplyBatchToStorages(ctx, m, gauges, counters)
	if err != nil {
		t.Fatalf("ApplyBatchToStorages() err = %v", err)
	}
	if len(byID) != 0 {
		t.Errorf("byID len = %d, want 0", len(byID))
	}
}

func ptrInt64(n int64) *int64 {
	return &n
}
