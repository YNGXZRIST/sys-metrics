package file

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
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
	svc := NewBackupService(storage)
	bc := svc.BackupCounters()
	if bc == nil {
		t.Error("BackupCounters() returned nil")
	}
	_, ok := bc.(metricsiface.BackupMetricStorage[*metrics.Counter])
	if !ok {
		t.Error("BackupCounters() returned wrong type")
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
	svc := NewBackupService(storage)
	bg := svc.BackupGauges()
	if bg == nil {
		t.Error("BackupGauges() returned nil")
	}
	_, ok := bg.(metricsiface.BackupMetricStorage[*metrics.Gauge])
	if !ok {
		t.Error("BackupGauges() returned wrong type")
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
	svc := NewBackupService(storage)
	c := svc.Counters()
	if c == nil {
		t.Error("Counters() returned nil")
	}
	_, ok := c.(metricsiface.MetricStorage[*metrics.Counter])
	if !ok {
		t.Error("Counters() returned wrong type")
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
	svc := NewBackupService(storage)
	g := svc.Gauges()
	if g == nil {
		t.Error("Gauges() returned nil")
	}
	_, ok := g.(metricsiface.MetricStorage[*metrics.Gauge])
	if !ok {
		t.Error("Gauges() returned wrong type")
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
	svc := NewBackupService(storage)
	counter := metrics.NewCounter("test_counter")
	counter.SetValue(10)
	gauge := metrics.NewGauge("test_gauge")
	gauge.SetValue(2.5)
	err = svc.Counters().Set(counter.ID, counter)
	if err != nil {
		t.Fatalf("Failed to set counter: %v", err)
	}
	err = svc.Gauges().Set(gauge.ID, gauge)
	if err != nil {
		t.Fatalf("Failed to set gauge: %v", err)
	}
	metricsList := svc.GetAllMetrics()
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
	svc := NewBackupService(storage)
	err = svc.ReadBackup()
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
	svc := NewBackupService(storage)
	counter := metrics.NewCounter("test_counter")
	counter.SetValue(5)
	err = svc.Counters().Set(counter.ID, counter)
	if err != nil {
		t.Fatalf("Failed to set counter: %v", err)
	}
	err = svc.WriteBackup()
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
	bs := NewBackupService(storage)
	if bs == nil {
		t.Error("NewBackupService() returned nil")
	}
	if bs.BackupStorage != storage {
		t.Error("NewBackupService() storage mismatch")
	}
}
