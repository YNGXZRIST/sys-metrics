package backup

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
	"testing"
	"time"
)

func TestBackupConfig_InitMetricsFromBackup(t *testing.T) {
	type fields struct {
		metrics []metrics.Metrics
		readErr error
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success counter initialization",
			fields: fields{
				metrics: []metrics.Metrics{
					{
						ID:    "test_counter",
						MType: common.Counter,
						Delta: ptr(int64(42)),
					},
				},
				readErr: nil,
			},
			wantErr: false,
		},
		{
			name: "succes gauge initialization",
			fields: fields{
				metrics: []metrics.Metrics{
					{
						ID:    "test_gauge",
						MType: common.Gauge,
						Value: ptr(123.45),
					},
				},
				readErr: nil,
			},
			wantErr: false,
		},
		{
			name: "succes mixed metrics initialization",
			fields: fields{
				metrics: []metrics.Metrics{
					{
						ID:    "counter1",
						MType: common.Counter,
						Delta: ptr(int64(10)),
					},
					{
						ID:    "gauge1",
						MType: common.Gauge,
						Value: ptr(5.5),
					},
				},
				readErr: nil,
			},
			wantErr: false,
		},
		{
			name: "empty backup metrics",
			fields: fields{
				metrics: []metrics.Metrics{},
				readErr: nil,
			},
			wantErr: false,
		},
		{
			name: "mixed metrics with zero values",
			fields: fields{
				metrics: []metrics.Metrics{
					{
						ID:    "counter_a",
						MType: common.Counter,
						Delta: ptr(int64(100)),
					},
					{
						ID:    "counter_b",
						MType: common.Counter,
						Delta: ptr(int64(200)),
					},
					{
						ID:    "counter_c",
						MType: common.Counter,
						Delta: ptr(int64(0)),
					},
				},
				readErr: nil,
			},
			wantErr: false,
		},
		{
			name: "multiple gauges initialization",
			fields: fields{
				metrics: []metrics.Metrics{
					{
						ID:    "gauge_a",
						MType: common.Gauge,
						Value: ptr(1.1),
					},
					{
						ID:    "gauge_b",
						MType: common.Gauge,
						Value: ptr(2.2),
					},
					{
						ID:    "gauge_c",
						MType: common.Gauge,
						Value: ptr(0.0),
					},
				},
				readErr: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := memstorage.NewMemStorage[string, *metrics.Counter]()
			gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
			svc.Init(counters, gauges)
			config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer config.Close()
			if tt.fields.readErr == nil {
				for _, m := range tt.fields.metrics {
					err := config.Writer.WriteMetricToBackup(&m)
					if err != nil {
						t.Fatalf("Failed to write test metric: %v", err)
					}
				}
				err = config.Reader.Reset()
				if err != nil {
					t.Fatalf("Failed to reset reader: %v", err)
				}
			}

			err = config.InitMetricsFromBackup()
			if (err != nil) != tt.wantErr {
				t.Errorf("InitMetricsFromBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackupConfig_InitBackupRoutine_Disabled(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *metrics.Counter]()
	gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
	svc.Init(counters, gauges)

	config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Millisecond*50, false)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Reader.Close()
	defer config.Writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err = config.InitBackupRoutine(ctx)
	if err != nil {
		t.Errorf("InitBackupRoutine() error = %v", err)
	}
}

func TestBackupConfig_InitBackupRoutine_SyncBackup(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *metrics.Counter]()
	gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
	svc.Init(counters, gauges)

	config, err := NewBackupConfig(common.TypeModeTest, "./test", 0*time.Second, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Reader.Close()
	defer config.Writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err = config.InitBackupRoutine(ctx)
	if err != nil {
		t.Errorf("InitBackupRoutine() error = %v", err)
	}
}

func TestBackupConfig_InitBackupRoutine_AsyncBackup(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *metrics.Counter]()
	gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
	svc.Init(counters, gauges)

	config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Millisecond*50, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Reader.Close()
	defer config.Writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	err = config.InitBackupRoutine(ctx)
	if err != nil {
		t.Errorf("InitBackupRoutine() error = %v", err)
	}
}

func TestBackupConfig_InitBackupRoutine_ContextCancelled(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *metrics.Counter]()
	gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
	svc.Init(counters, gauges)

	config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Millisecond*10, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Reader.Close()
	defer config.Writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = config.InitBackupRoutine(ctx)
	if err != nil {
		t.Errorf("InitBackupRoutine() error = %v", err)
	}
}

func ptr[T any](v T) *T {
	return &v
}
