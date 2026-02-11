package file

import (
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/pkg/storage"
	"testing"
	"time"
)

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
				config:         &Config{Interval: 0},
				metricsHandler: nil,
			},
			want: true,
		},
		{
			name: "Async mode (interval>0)",
			fields: fields{
				MemStorage:     storage.NewMemStorage[string, *metrics.Counter](),
				config:         &Config{Interval: time.Second},
				metricsHandler: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &counterBackupStorage{
				MemStorage:     tt.fields.MemStorage,
				config:         tt.fields.config,
				metricsHandler: tt.fields.metricsHandler,
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
			config, err := NewConfig("test", "./test", time.Second*1, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer func() {
				_ = config.Close()
				_ = config.Cleanup()
			}()
			s := &counterBackupStorage{
				MemStorage:     storage.NewMemStorage[string, *metrics.Counter](),
				config:         config,
				metricsHandler: nil,
			}
			err = s.Set(tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.args.value != nil {
				stored, _ := s.MemStorage.Get(tt.args.key)
				if stored == nil {
					t.Errorf("Set() did not store value for key %v", tt.args.key)
				}
			}
		})
	}
}
