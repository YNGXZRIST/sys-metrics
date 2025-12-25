package metrics

import (
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
				Metrics: Metrics{ID: "test", MType: MTypeGauge},
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
	if got := m.Metrics.MType; got != MTypeGauge {
		t.Fatalf("NewGauge() = %v, want %v", got, MTypeGauge)
	}
}
