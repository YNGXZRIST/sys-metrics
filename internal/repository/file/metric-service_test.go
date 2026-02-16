package file

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"testing"
	"time"
)

func TestBackupService_BackupCounters(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)
	bc := svc.BackupCounters(context.Background())
	if bc == nil {
		t.Error("BackupCounters() returned nil")
	}
}

func TestBackupService_BackupGauges(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)
	bg := svc.BackupGauges(context.Background())
	if bg == nil {
		t.Error("BackupGauges() returned nil")
	}
}

func TestBackupService_Counters(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)
	c := svc.Counters()
	if c == nil {
		t.Error("Counters() returned nil")
	}
}

func TestBackupService_Gauges(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)
	g := svc.Gauges()
	if g == nil {
		t.Error("Gauges() returned nil")
	}
}

func TestBackupService_GetAllMetrics(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)
	counter := metrics.NewCounter("test_counter")
	counter.SetValue(10)
	gauge := metrics.NewGauge("test_gauge")
	gauge.SetValue(2.5)
	err = svc.Counters().Set(context.Background(), counter.ID, counter)
	if err != nil {
		t.Fatalf("Failed to set counter: %v", err)
	}
	err = svc.Gauges().Set(context.Background(), gauge.ID, gauge)
	if err != nil {
		t.Fatalf("Failed to set gauge: %v", err)
	}
	metricsList := svc.GetAllMetrics(context.Background())
	if len(metricsList) != 2 {
		t.Errorf("GetAllMetrics() returned %d metrics, want 2", len(metricsList))
	}
}

func TestBackupService_InitRoutine(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
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

func TestBackupService_ReadBackup(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)
	err = svc.ReadBackup(context.Background())
	if err != nil {
		t.Errorf("ReadBackup() error = %v", err)
	}
}

func TestBackupService_WriteBackup(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	svc := NewBackupService(storage)
	counter := metrics.NewCounter("test_counter")
	counter.SetValue(5)
	err = svc.Counters().Set(context.Background(), counter.ID, counter)
	if err != nil {
		t.Fatalf("Failed to set counter: %v", err)
	}
	err = svc.WriteBackup(context.Background())
	if err != nil {
		t.Errorf("WriteBackup() error = %v", err)
	}
}

func TestNewBackupService(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*1, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer func() {
		_ = config.Close()
		_ = config.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	defer storage.Close()
	bs := NewBackupService(storage)
	if bs == nil {
		t.Error("NewBackupService() returned nil")
		return
	}
	if bs.BackupStorage != storage {
		t.Error("NewBackupService() storage mismatch")
	}
}
