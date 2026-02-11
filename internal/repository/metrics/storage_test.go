package metrics

import (
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	file "sys-metrics/internal/repository/file"
	"sys-metrics/internal/repository/memory"
	"sys-metrics/pkg/storage"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	service := Init(memory.NewService(counters, gauges))

	if service == nil {
		t.Fatal("Init() returned nil service")
	}
	if service.Counters() != counters {
		t.Error("Init() did not set counters correctly")
	}
	if service.Gauges() != gauges {
		t.Error("Init() did not set gauges correctly")
	}
	if defaultService != service {
		t.Error("Init() did not set defaultService correctly")
	}
}

func TestCounters(t *testing.T) {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	Init(memory.NewService(counters, gauges))

	got := Counters()

	if got == nil {
		t.Fatal("Counters() returned nil")
	}
	if got != counters {
		t.Errorf("Counters() returned wrong storage")
	}
}

func TestGauges(t *testing.T) {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	Init(memory.NewService(counters, gauges))

	got := Gauges()

	if got == nil {
		t.Fatal("Gauges() returned nil")
	}
	if got != gauges {
		t.Errorf("Gauges() returned wrong storage")
	}
}

func TestGetService(t *testing.T) {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	service := Init(memory.NewService(counters, gauges))

	got := GetService()

	if got == nil {
		t.Fatal("GetService() returned nil")
	}
	if got != service {
		t.Error("GetService() returned wrong service")
	}
}

func TestGetAllMetrics(t *testing.T) {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	Init(memory.NewService(counters, gauges))

	counter := metrics.NewCounter("test_counter")
	counter.SetValue(42)
	_ = counters.Set("test_counter", counter)

	gauge := metrics.NewGauge("test_gauge")
	gauge.SetValue(3.14)
	_ = gauges.Set("test_gauge", gauge)

	got := GetAllMetrics()

	if len(got) != 2 {
		t.Errorf("GetAllMetrics() returned %d metrics, expected 2", len(got))
	}
}

func TestBackupMemoryService_ReadBackup(t *testing.T) {
	type args struct {
		metric *metrics.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "gauge metric",
			args: args{
				metric: &metrics.Metrics{
					ID:    "test_gauge",
					MType: "gauge",
					Value: func() *float64 { v := 123.45; return &v }(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := file.NewConfig(common.TypeModeTest, "", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer func() { _ = config.Close() }()
			err = config.MetricsHandler.Write(tt.args.metric)
			if err != nil {
				t.Fatalf("Failed to write metric to backup: %v", err)
			}
		})
	}
}
