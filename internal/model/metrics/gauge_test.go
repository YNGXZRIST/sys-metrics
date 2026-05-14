package metrics

import (
	"sys-metrics/internal/common"
	"testing"
)

func TestGauge_SetValue(t *testing.T) {
	type fields struct {
		Metrics Metrics
	}
	type args struct {
		v float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: "SetValue",
			fields: fields{
				Metrics: Metrics{ID: "test", MType: common.Gauge},
			},
			args: args{
				v: 1.0,
			},
			want: 1.0,
		},
		{
			name:   "empty",
			fields: fields{},
			args:   args{},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				Metrics: tt.fields.Metrics,
			}
			if got := g.SetValue(tt.args.v); got != tt.want {
				t.Errorf("SetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGauge(t *testing.T) {
	m := NewGauge("test")
	if m == nil {
		t.Fatal("NewGauge() returned nil")
	}
	if got := m.Metrics.MType; got != common.Gauge {
		t.Fatalf("NewGauge() = %v, want %v", got, common.Gauge)
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := &Metrics{
		ID:    "metric",
		MType: common.Gauge,
		Delta: new(int64(7)),
		Value: new(3.14),
		Hash:  "hash",
	}

	m.Reset()

	if m.ID != "" || m.MType != "" || m.Delta != nil || m.Value != nil || m.Hash != "" {
		t.Fatalf("Reset() = %#v, want zeroed metric", m)
	}
}

func TestGauge_Reset_NilSafe(t *testing.T) {
	var g *Gauge
	g.Reset()
}
