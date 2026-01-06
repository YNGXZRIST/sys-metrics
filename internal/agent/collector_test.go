package agent

import (
	"runtime"
	"sys-metrics/internal/common"
	"testing"
)

func TestCollector_ResetPollMetric(t *testing.T) {
	tests := []struct {
		name    string
		initial float64
		want    float64
	}{
		{
			name:    "reset from positive value",
			initial: 12,
			want:    0,
		},
		{
			name:    "reset from zero",
			initial: 0,
			want:    0,
		},
		{
			name:    "reset from large value",
			initial: 999999,
			want:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector()
			c.SetCounter(PollCount, tt.initial)
			c.ResetPollMetric()
			if got := c.GetPollCountMetric(); got != tt.want {
				t.Errorf("ResetPollMetric() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollector_SetPollCounterMetric(t *testing.T) {
	tests := []struct {
		name       string
		initial    float64
		increments int
		want       float64
	}{
		{
			name:       "increment from zero",
			initial:    0,
			increments: 1,
			want:       1,
		},
		{
			name:       "increment from positive",
			initial:    12,
			increments: 1,
			want:       13,
		},
		{
			name:       "multiple increments",
			initial:    0,
			increments: 5,
			want:       5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector()
			c.SetCounter(PollCount, tt.initial)
			for i := 0; i < tt.increments; i++ {
				c.SetPollCounterMetric()
			}
			if got := c.GetPollCountMetric(); got != tt.want {
				t.Errorf("SetPollCounterMetric() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollector_GetPollCountMetric(t *testing.T) {
	tests := []struct {
		name    string
		initial float64
		want    float64
	}{
		{
			name:    "positive value",
			initial: 12,
			want:    12,
		},
		{
			name:    "zero",
			initial: 0,
			want:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector()
			c.SetCounter(PollCount, tt.initial)
			if got := c.GetPollCountMetric(); got != tt.want {
				t.Errorf("GetPollCountMetric() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollector_SetRandomValueMetric(t *testing.T) {
	c := NewCollector()
	for i := 0; i < 10; i++ {
		c.SetRandomValueMetric()
		got := c.GetRandomValueMetric()
		if got < 0 || got >= 1 {
			t.Errorf("SetRandomValueMetric() got = %v, want value in range [0, 1)", got)
		}
	}
}

func TestUpdateFromStats(t *testing.T) {
	tests := []struct {
		name      string
		stats     *runtime.MemStats
		checkName string
		want      float64
	}{
		{
			name:      "Alloc metric",
			stats:     &runtime.MemStats{Alloc: 1024},
			checkName: Alloc,
			want:      1024,
		},
		{
			name:      "HeapAlloc metric",
			stats:     &runtime.MemStats{HeapAlloc: 2048},
			checkName: HeapAlloc,
			want:      2048,
		},
		{
			name:      "TotalAlloc metric",
			stats:     &runtime.MemStats{TotalAlloc: 4096},
			checkName: TotalAlloc,
			want:      4096,
		},
		{
			name:      "Sys metric",
			stats:     &runtime.MemStats{Sys: 8192},
			checkName: Sys,
			want:      8192,
		},
		{
			name:      "zero value",
			stats:     &runtime.MemStats{Alloc: 0},
			checkName: Alloc,
			want:      0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector()
			c.UpdateFromStats(tt.stats)

			v, ok := c.GetGauge(tt.checkName)
			if !ok {
				t.Errorf("GetGauge(%s) returned ok=false, want true", tt.checkName)
			}
			if v != tt.want {
				t.Errorf("GetGauge(%s) = %v, want %v", tt.checkName, v, tt.want)
			}
		})
	}
}

func TestNewCollector(t *testing.T) {
	c := NewCollector()

	if c == nil {
		t.Fatal("NewCollector() returned nil")
	}
	if c.metrics == nil {
		t.Fatal("NewCollector() metrics is nil")
	}
	if _, ok := c.metrics[common.Gauge]; !ok {
		t.Error("NewCollector() missing Gauge map")
	}
	if _, ok := c.metrics[common.Counter]; !ok {
		t.Error("NewCollector() missing Counter map")
	}
}

func TestCollector_SetGauge(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value float64
	}{
		{
			name:  "set positive value",
			key:   "test_metric",
			value: 123.45,
		},
		{
			name:  "set zero",
			key:   "zero_metric",
			value: 0,
		},
		{
			name:  "set negative value",
			key:   "negative_metric",
			value: -100.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector()
			c.SetGauge(tt.key, tt.value)

			got, ok := c.GetGauge(tt.key)
			if !ok {
				t.Errorf("GetGauge(%s) returned ok=false", tt.key)
			}
			if got != tt.value {
				t.Errorf("GetGauge(%s) = %v, want %v", tt.key, got, tt.value)
			}
		})
	}
}

func TestCollector_GetGauge_NotFound(t *testing.T) {
	c := NewCollector()

	_, ok := c.GetGauge("nonexistent")
	if ok {
		t.Error("GetGauge(nonexistent) returned ok=true, want false")
	}
}

func TestCollector_SetCounter(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value float64
	}{
		{
			name:  "set positive value",
			key:   "test_counter",
			value: 100,
		},
		{
			name:  "set zero",
			key:   "zero_counter",
			value: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector()
			c.SetCounter(tt.key, tt.value)

			got, ok := c.GetCounter(tt.key)
			if !ok {
				t.Errorf("GetCounter(%s) returned ok=false", tt.key)
			}
			if got != tt.value {
				t.Errorf("GetCounter(%s) = %v, want %v", tt.key, got, tt.value)
			}
		})
	}
}

func TestCollector_GetCounter_NotFound(t *testing.T) {
	c := NewCollector()

	_, ok := c.GetCounter("nonexistent")
	if ok {
		t.Error("GetCounter(nonexistent) returned ok=true, want false")
	}
}

func TestCollector_Update(t *testing.T) {
	c := NewCollector()
	c.Update()
	metricsToCheck := []string{Alloc, HeapAlloc, Sys, TotalAlloc}
	for _, name := range metricsToCheck {
		if _, ok := c.GetGauge(name); !ok {
			t.Errorf("Update() did not set gauge %s", name)
		}
	}
}
