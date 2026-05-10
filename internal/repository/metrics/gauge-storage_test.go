package metrics

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/pkg/storage"
	"testing"
	"time"
)

func Test_gaugeBackupStorage_NeedSync(t *testing.T) {
	type fields struct {
		MemStorage     *storage.MemStorage[string, *metrics.Gauge]
		config         metricsiface.BackupConfig
		metricsHandler metricsiface.Handler
	}
	tests := []struct {
		fields fields
		name   string
		want   bool
	}{
		{
			name: "Sync mode (interval=0)",
			fields: fields{
				MemStorage:     storage.NewMemStorage[string, *metrics.Gauge](),
				config:         &mockBackupConfig{interval: 0},
				metricsHandler: nil,
			},
			want: true,
		},
		{
			name: "Async mode (interval>0)",
			fields: fields{
				MemStorage:     storage.NewMemStorage[string, *metrics.Gauge](),
				config:         &mockBackupConfig{interval: time.Second},
				metricsHandler: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GaugeBackupStorage{
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

func Test_gaugeBackupStorage_Set(t *testing.T) {
	type args struct {
		value *metrics.Gauge
		key   string
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "Set valid gauge",
			args: args{
				key:   "gauge1",
				value: metrics.NewGauge("gauge1"),
			},
			wantErr: false,
		},
		{
			name: "Set nil value",
			args: args{
				key:   "gauge2",
				value: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &mockBackupConfig{interval: time.Second}
			s := &GaugeBackupStorage{
				MemStorage:     storage.NewMemStorage[string, *metrics.Gauge](),
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
