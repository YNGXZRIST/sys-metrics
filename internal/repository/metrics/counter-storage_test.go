package metrics

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/pkg/storage"
	"testing"
	"time"
)

// mockBackupConfig is a test implementation of BackupConfig to avoid import cycle
type mockBackupConfig struct {
	interval time.Duration
}

func (m *mockBackupConfig) NeedSync() bool {
	return m.interval == 0
}

func Test_counterBackupStorage_NeedSync(t *testing.T) {
	type fields struct {
		MemStorage     *storage.MemStorage[string, *metrics.Counter]
		config         metricsiface.BackupConfig
		metricsHandler metricsiface.Handler
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Sync mode (interval=0)",
			fields: fields{
				MemStorage:     storage.NewMemStorage[string, *metrics.Counter](),
				config:         &mockBackupConfig{interval: 0},
				metricsHandler: nil,
			},
			want: true,
		},
		{
			name: "Async mode (interval>0)",
			fields: fields{
				MemStorage:     storage.NewMemStorage[string, *metrics.Counter](),
				config:         &mockBackupConfig{interval: time.Second},
				metricsHandler: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CounterBackupStorage{
				MemStorage:     tt.fields.MemStorage,
				Config:         tt.fields.config,
				MetricsHandler: tt.fields.metricsHandler,
			}
			if got := s.NeedSync(); got != tt.want {
				t.Errorf("NeedSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_counterBackupStorage_Set(t *testing.T) {
	type fields struct {
		MemStorage     *storage.MemStorage[string, *metrics.Counter]
		config         metricsiface.BackupConfig
		metricsHandler metricsiface.Handler
	}
	type args struct {
		key   string
		value *metrics.Counter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Set valid counter",
			args: args{
				key:   "counter1",
				value: metrics.NewCounter("counter1"),
			},
			wantErr: false,
		},
		{
			name: "Set nil value",
			args: args{
				key:   "counter2",
				value: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &mockBackupConfig{interval: time.Second * 1}
			s := &CounterBackupStorage{
				MemStorage:     storage.NewMemStorage[string, *metrics.Counter](),
				Config:         config,
				MetricsHandler: nil,
			}
			err := s.Set(context.Background(), tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.args.value != nil {
				stored, _ := s.MemStorage.Get(context.Background(), tt.args.key)
				if stored == nil {
					t.Errorf("Set() did not store value for key %v", tt.args.key)
				}
			}
		})
	}
}
