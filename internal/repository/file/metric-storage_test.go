package file

import (
	"context"
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	m "sys-metrics/internal/repository/metrics"
	"testing"
	"time"
)

func TestMetricBackupStorage_NeedSync(t *testing.T) {
	type fields struct {
		BackupStorage *m.BackupStorage
		Config        *Config
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Sync mode (interval=0)",
			fields: fields{
				BackupStorage: &m.BackupStorage{},
				Config:        &Config{Interval: 0},
			},
			want: true,
		},
		{
			name: "Async mode (interval>0)",
			fields: fields{
				BackupStorage: &m.BackupStorage{},
				Config:        &Config{Interval: 1},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MetricBackupStorage{
				BackupStorage: tt.fields.BackupStorage,
				Config:        tt.fields.Config,
			}
			if got := s.NeedSync(); got != tt.want {
				t.Errorf("NeedSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMetricFileBackupStorage(t *testing.T) {
	t.Run("Nil config", func(t *testing.T) {
		got, err := NewMetricFileBackupStorage(nil)
		if err == nil {
			t.Errorf("NewMetricFileBackupStorage() error = nil, want error")
		}
		if got != nil {
			t.Errorf("NewMetricFileBackupStorage() got = %v, want nil", got)
		}
	})

	t.Run("Valid config", func(t *testing.T) {
		config, err := NewConfig(common.TypeModeTest, "./test", time.Second, true)
		if err != nil {
			t.Fatalf("NewConfig: %v", err)
		}
		defer config.Cleanup()
		got, err := NewMetricFileBackupStorage(config)
		if err != nil {
			t.Errorf("NewMetricFileBackupStorage() error = %v", err)
		}
		if got == nil {
			t.Errorf("NewMetricFileBackupStorage() got = nil, want not nil")
		}
		if got != nil {
			defer got.Close()
		}
	})
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
			f, err := os.CreateTemp("", "backup_test_*.json")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(f.Name()) }()
			cfg, err := NewConfig(common.TypeModeTest, "./test", time.Second, true)
			if err != nil {
				t.Fatalf("Failed to create config: %v", err)
			}
			defer cfg.Cleanup()
			backupStorage, err := NewMetricFileBackupStorage(cfg)
			if err != nil {
				t.Fatalf("Failed to create backup storage: %v", err)
			}
			defer backupStorage.Close()
			err = backupStorage.MetricsHandler.Write(context.TODO(), tt.args.metric)
			if err != nil {
				t.Fatalf("Failed to write metric to backup: %v", err)
			}
			service := NewBackupService(backupStorage)
			err = service.ReadBackup(context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackupService_InitRoutine_Disabled(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Millisecond*50, false)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Cleanup()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err = svc.InitRoutine(ctx)
	if err != nil {
		t.Errorf("InitRoutine() error = %v", err)
	}
}
func TestBackupService_InitRoutine_SyncBackup(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", 0*time.Second, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Cleanup()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err = svc.InitRoutine(ctx)
	if err != nil {
		t.Errorf("InitRoutine() error = %v", err)
	}

}
func TestBackupService_InitRoutine_AsyncBackup(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Millisecond*50, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Cleanup()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	err = svc.InitRoutine(ctx)
	if err != nil {
		t.Errorf("InitRoutine() error = %v", err)
	}
}
func TestBackupService_InitRoutine_ContextCancelled(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Millisecond*10, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Cleanup()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = svc.InitRoutine(ctx)
	if err != nil {
		t.Errorf("InitRoutine() error = %v", err)
	}
}
