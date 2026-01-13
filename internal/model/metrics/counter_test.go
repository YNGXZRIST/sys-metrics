package metrics

import (
	"sys-metrics/internal/common"
	"testing"
)

func TestCounter_SetValue(t *testing.T) {
	type fields struct {
		Metrics Metrics
	}
	type args struct {
		v int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "SetValue",
			fields: fields{
				Metrics: Metrics{ID: "test", MType: common.Counter},
			},
			args: args{
				v: 1,
			},
			want: 1,
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
			g := &Counter{
				Metrics: tt.fields.Metrics,
			}
			if got := g.SetValue(tt.args.v); got != tt.want {
				t.Errorf("SetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCounter(t *testing.T) {
	m := NewCounter("test")
	if m == nil {
		t.Fatal("NewCounter() returned nil")
	}
	if got := m.Metrics.MType; got != common.Counter {
		t.Fatalf("NewCounter() = %v, want %v", got, common.Counter)
	}
}
