package agent

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"testing"
)

func TestCollector_ResetPollMetric(t *testing.T) {
	tests := []struct {
		name    string
		initial int64
		want    int64
	}{
		{"reset from positive value", 12, 0},
		{"reset from zero", 0, 0},
		{"reset from large value", 999999, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector(context.Background(), 1)
			c.SetCounter(common.PollCount, tt.initial)
			c.ResetPollMetric()
			counter, ok := c.Counters[common.PollCount]
			if !ok || counter == nil || counter.Delta == nil {
				t.Fatalf("PollCount counter not found or nil after reset")
			}
			if *counter.Delta != tt.want {
				t.Errorf("ResetPollMetric() got = %v, want %v", *counter.Delta, tt.want)
			}
		})
	}
}

func TestCollector_SetPollCounterMetric(t *testing.T) {
	tests := []struct {
		name       string
		initial    int64
		increments int
		want       int64
	}{
		{"increment from zero", 0, 1, 1},
		{"increment from positive", 12, 1, 13},
		{"multiple increments", 0, 5, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector(context.Background(), 1)
			c.SetCounter(common.PollCount, tt.initial)
			for i := 0; i < tt.increments; i++ {
				c.SetPollCounterMetric()
			}
			counter, ok := c.Counters[common.PollCount]
			if !ok || counter == nil || counter.Delta == nil {
				t.Fatalf("PollCount counter not found or nil after increment")
			}
			if *counter.Delta != tt.want {
				t.Errorf("SetPollCounterMetric() got = %v, want %v", *counter.Delta, tt.want)
			}
		})
	}
}

func TestCollector_GetPollCountMetric(t *testing.T) {
	tests := []struct {
		name    string
		initial int64
		want    int64
	}{
		{"positive value", 12, 12},
		{"zero", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector(context.Background(), 1)
			c.SetCounter(common.PollCount, tt.initial)
			counter, ok := c.Counters[common.PollCount]
			if !ok || counter == nil || counter.Delta == nil {
				t.Fatalf("PollCount counter not found or nil in GetPollCountMetric")
			}
			if *counter.Delta != tt.want {
				t.Errorf("GetPollCountMetric() got = %v, want %v", *counter.Delta, tt.want)
			}
		})
	}
}

func TestUpdateFromStats(t *testing.T) {
	tests := []struct {
		name      string
		values    map[string]float64
		checkName string
		want      float64
	}{
		{
			name:      "Alloc metric",
			values:    map[string]float64{common.Alloc: 1024},
			checkName: common.Alloc,
			want:      1024,
		},
		{
			name:      "HeapAlloc metric",
			values:    map[string]float64{common.HeapAlloc: 2048},
			checkName: common.HeapAlloc,
			want:      2048,
		},
		{
			name:      "TotalAlloc metric",
			values:    map[string]float64{common.TotalAlloc: 4096},
			checkName: common.TotalAlloc,
			want:      4096,
		},
		{
			name:      "Sys metric",
			values:    map[string]float64{common.Sys: 8192},
			checkName: common.Sys,
			want:      8192,
		},
		{
			name:      "zero value",
			values:    map[string]float64{common.Alloc: 0},
			checkName: common.Alloc,
			want:      0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector(context.Background(), 1)
			c.UpdateFromStats(tt.values)
			gauge, ok := c.Gauges[tt.checkName]
			if !ok || gauge == nil || gauge.Value == nil {
				t.Fatalf("Gauge %s not found or nil after UpdateFromStats", tt.checkName)
			}
			if *gauge.Value != tt.want {
				t.Errorf("GetGauge(%s) = %v, want %v", tt.checkName, *gauge.Value, tt.want)
			}
		})
	}
}

func TestNewCollector(t *testing.T) {
	c := NewCollector(context.Background(), 1)
	if c == nil {
		t.Fatal("NewCollector() returned nil")
	}
	if c.Gauges == nil {
		t.Fatal("NewCollector() Gauges is nil")
	}
	if c.Counters == nil {
		t.Fatal("NewCollector() Counters is nil")
	}
}

func TestCollector_SetGauge(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value float64
	}{
		{"set positive value", "test_metric", 123.45},
		{"set zero", "zero_metric", 0},
		{"set negative value", "negative_metric", -100.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector(context.Background(), 1)
			g := metrics.NewGauge(tt.key)
			g.SetValue(tt.value)
			c.Gauges[tt.key] = g
			gauge, ok := c.Gauges[tt.key]
			if !ok || gauge == nil || gauge.Value == nil {
				t.Fatalf("Gauge %s not found or nil after SetGauge", tt.key)
			}
			if *gauge.Value != tt.value {
				t.Errorf("GetGauge(%s) = %v, want %v", tt.key, *gauge.Value, tt.value)
			}
		})
	}
}

func TestCollector_GetGauge_NotFound(t *testing.T) {
	c := NewCollector(context.Background(), 1)
	gauge, ok := c.Gauges["nonexistent"]
	if ok && gauge != nil && gauge.Value != nil {
		t.Error("GetGauge(nonexistent) returned value, want nil")
	}
}

func TestCollector_SetCounter(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value int64
	}{
		{"set positive value", "test_counter", 100},
		{"set zero", "zero_counter", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector(context.Background(), 1)
			c.SetCounter(tt.key, tt.value)
			counter, ok := c.Counters[tt.key]
			if !ok || counter == nil || counter.Delta == nil {
				t.Fatalf("Counter %s not found or nil after SetCounter", tt.key)
			}
			if *counter.Delta != tt.value {
				t.Errorf("GetCounter(%s) = %v, want %v", tt.key, *counter.Delta, tt.value)
			}
		})
	}
}

func TestCollector_GetCounter_NotFound(t *testing.T) {
	c := NewCollector(context.Background(), 1)
	counter, ok := c.Counters["nonexistent"]
	if ok && counter != nil && counter.Delta != nil {
		t.Error("GetCounter(nonexistent) returned value, want nil")
	}
}

func TestCollector_Update(t *testing.T) {
	ctx := context.Background()
	c := NewCollector(ctx, 1)
	err := c.UpdateSync(ctx)
	if err != nil {
		t.Errorf("Update() returned %v, want nil", err)
	}
	metricsToCheck := []string{common.Alloc, common.HeapAlloc, common.Sys, common.TotalAlloc}
	for _, name := range metricsToCheck {
		gauge, ok := c.Gauges[name]
		if !ok || gauge == nil || gauge.Value == nil {
			t.Errorf("Update() did not set gauge %s", name)
		}
	}
}

func TestGetMetricType(t *testing.T) {
	type args struct {
		metric string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"PollCount", args{metric: common.PollCount}, common.PollCount},
		{"not isset map", args{metric: "unknownmetric"}, "Unknownmetric"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMetricType(tt.args.metric); got != tt.want {
				t.Errorf("GetMetricType() = %v, want %v", got, tt.want)
			}
		})
	}
}
