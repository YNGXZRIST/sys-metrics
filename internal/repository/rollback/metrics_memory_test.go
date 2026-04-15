package rollback

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
	"testing"
)

func TestMemory(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		snapshot          []metrics.Metrics
		newMetrics        []metrics.Metrics
		gaugesBefore      map[string]float64
		countersBefore    map[string]int64
		wantGauge         map[string]float64
		wantGaugeAbsent   []string
		wantCounter       map[string]int64
		wantCounterAbsent []string
	}{
		{
			name:         "gauge restored from snapshot",
			snapshot:     []metrics.Metrics{{ID: "g1", MType: common.Gauge, Value: ptrFloat64(100.5)}},
			newMetrics:   []metrics.Metrics{{ID: "g1", MType: common.Gauge, Value: ptrFloat64(999.0)}},
			gaugesBefore: map[string]float64{"g1": 999.0},
			wantGauge:    map[string]float64{"g1": 100.5},
		},
		{
			name:            "gauge deleted when not in snapshot",
			snapshot:        []metrics.Metrics{},
			newMetrics:      []metrics.Metrics{{ID: "g1", MType: common.Gauge, Value: ptrFloat64(1.0)}},
			gaugesBefore:    map[string]float64{"g1": 1.0},
			wantGaugeAbsent: []string{"g1"},
		},
		{
			name:           "counter restored from snapshot",
			snapshot:       []metrics.Metrics{{ID: "c1", MType: common.Counter, Delta: ptrInt64(10)}},
			newMetrics:     []metrics.Metrics{{ID: "c1", MType: common.Counter, Delta: ptrInt64(100)}},
			countersBefore: map[string]int64{"c1": 100},
			wantCounter:    map[string]int64{"c1": 10},
		},
		{
			name:              "counter deleted when not in snapshot",
			snapshot:          []metrics.Metrics{},
			newMetrics:        []metrics.Metrics{{ID: "c1", MType: common.Counter, Delta: ptrInt64(5)}},
			countersBefore:    map[string]int64{"c1": 5},
			wantCounterAbsent: []string{"c1"},
		},
		{
			name: "existing restored and new from batch deleted",
			snapshot: []metrics.Metrics{
				{ID: "g1", MType: common.Gauge, Value: ptrFloat64(1.0)},
				{ID: "c1", MType: common.Counter, Delta: ptrInt64(2)},
			},
			newMetrics: []metrics.Metrics{
				{ID: "g1", MType: common.Gauge, Value: ptrFloat64(100.0)},
				{ID: "c1", MType: common.Counter, Delta: ptrInt64(200)},
				{ID: "g_new", MType: common.Gauge, Value: ptrFloat64(50.0)},
				{ID: "c_new", MType: common.Counter, Delta: ptrInt64(7)},
			},
			gaugesBefore:      map[string]float64{"g1": 100.0, "g_new": 50.0},
			countersBefore:    map[string]int64{"c1": 200, "c_new": 7},
			wantGauge:         map[string]float64{"g1": 1.0},
			wantGaugeAbsent:   []string{"g_new"},
			wantCounter:       map[string]int64{"c1": 2},
			wantCounterAbsent: []string{"c_new"},
		},
		{
			name:         "no change when newMetrics is empty",
			snapshot:     []metrics.Metrics{{ID: "g1", MType: common.Gauge, Value: ptrFloat64(3.14)}},
			newMetrics:   []metrics.Metrics{},
			gaugesBefore: map[string]float64{"g1": 3.14},
			wantGauge:    map[string]float64{"g1": 3.14},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gauges := storage.NewMemStorage[string, *metrics.Gauge]()
			counters := storage.NewMemStorage[string, *metrics.Counter]()

			for id, v := range tt.gaugesBefore {
				_ = gauges.Set(ctx, id, &metrics.Gauge{Metrics: metrics.Metrics{ID: id, MType: common.Gauge, Value: new(v)}})
			}
			for id, d := range tt.countersBefore {
				_ = counters.Set(ctx, id, &metrics.Counter{Metrics: metrics.Metrics{ID: id, MType: common.Counter, Delta: new(d)}})
			}

			Memory(ctx, gauges, counters, tt.snapshot, tt.newMetrics)

			for id, want := range tt.wantGauge {
				g, err := gauges.Get(ctx, id)
				if err != nil {
					t.Fatalf("gauge %q: Get: %v", id, err)
				}
				if g.Value == nil || *g.Value != want {
					t.Errorf("gauge %q = %v, want %v", id, valueOrNil(g.Value), want)
				}
			}
			for _, id := range tt.wantGaugeAbsent {
				if gauges.Has(ctx, id) {
					t.Errorf("gauge %q want be absent", id)
				}
			}
			for id, want := range tt.wantCounter {
				c, err := counters.Get(ctx, id)
				if err != nil {
					t.Fatalf("counter %q: Get: %v", id, err)
				}
				if c.Delta == nil || *c.Delta != want {
					t.Errorf("counter %q = %v, want %v", id, deltaOrNil(c.Delta), want)
				}
			}
			for _, id := range tt.wantCounterAbsent {
				if counters.Has(ctx, id) {
					t.Errorf("counter %q want be absent", id)
				}
			}
		})
	}
}

func ptrFloat64(x float64) *float64 { return &x }
func ptrInt64(x int64) *int64       { return &x }

func valueOrNil(v *float64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func deltaOrNil(d *int64) interface{} {
	if d == nil {
		return nil
	}
	return *d
}
