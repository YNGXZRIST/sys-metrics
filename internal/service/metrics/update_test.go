package metrics

import (
	"context"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	"sys-metrics/internal/repository/metrics"
	"testing"
)

func TestUpdate(t *testing.T) {
	type args struct {
		metricType string
		name       string
		value      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "empty",
			args:    args{},
			wantErr: true,
		},
		{
			name: "unknown type",
			args: args{
				metricType: "histogram",
				name:       "x",
				value:      "1",
			},
			wantErr: true,
		},
		{
			name: "bad counter value",
			args: args{
				metricType: common.Counter,
				name:       "c",
				value:      "nope",
			},
			wantErr: true,
		},
		{
			name: "bad gauge value",
			args: args{
				metricType: "gauge",
				name:       "g",
				value:      "x.y.z",
			},
			wantErr: true,
		},
		{
			name: "metric counter",
			args: args{
				metricType: common.Counter,
				name:       "c1",
				value:      "42",
			},
			wantErr: false,
		},
		{
			name: "metric gauge",
			args: args{
				metricType: "gauge",
				name:       "test",
				value:      "10",
			},
			wantErr: false,
		},
	}
	metrics.Init(memory.NewService())
	ms := NewService(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ms.Update(context.Background(), tt.args.metricType, tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_updateCounter(t *testing.T) {
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				name:  "test",
				value: 12,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics.Init(memory.NewService())
			if err := updateCounter(context.Background(), tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("updateCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
			val, err := metrics.Counters().Get(context.Background(), tt.args.name)
			if err != nil {
				t.Errorf("counters.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if *val.Delta != tt.args.value {
				t.Errorf("updateCounter() delta got %v, want %v", *val.Delta, tt.args.value)
			}
		})
	}
}

func Test_updateGauge(t *testing.T) {
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				name:  "test",
				value: 12,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics.Init(memory.NewService())
			if err := updateGauge(context.Background(), tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("updateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
			val, err := metrics.Gauges().Get(context.Background(), tt.args.name)
			if err != nil {
				t.Errorf("updateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if *val.Value != tt.args.value {
				t.Errorf("updateGauge() Value got %v, want %v", *val.Delta, tt.args.value)
			}
		})
	}
}

func TestBatchUpdateMetrics_empty(t *testing.T) {
	metrics.Init(memory.NewService())
	ms := NewService(nil)
	if err := ms.BatchUpdateMetrics(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := ms.BatchUpdateMetrics(context.Background(), []models.Metrics{}); err != nil {
		t.Fatal(err)
	}
}

func TestBatchUpdateMetrics_oneGauge(t *testing.T) {
	metrics.Init(memory.NewService())
	ms := NewService(nil)
	v := 3.0
	err := ms.BatchUpdateMetrics(context.Background(), []models.Metrics{
		{ID: "g1", MType: "gauge", Value: &v},
	})
	if err != nil {
		t.Fatal(err)
	}
	g, err := metrics.Gauges().Get(context.Background(), "g1")
	if err != nil || g.Value == nil || *g.Value != v {
		t.Fatalf("gauge: err=%v val=%v", err, g.Value)
	}
}
