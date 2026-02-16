package metrics

import (
	"context"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Update(context.Background(), tt.args.metricType, tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
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
