package repository

import (
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"testing"
	"time"
)

func TestInitBackup(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*10, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Close()

	storage, err := NewMetricBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}

	service := InitBackup(storage)

	if service == nil {
		t.Fatal("InitBackup() returned nil service")
	}
	if service.Counters() == nil {
		t.Error("InitBackup() counters is nil")
	}
	if service.Gauges() == nil {
		t.Error("InitBackup() gauges is nil")
	}
}

func TestWriteBackup(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", time.Second*10, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Close()

	storage, err := NewMetricBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create metric backup storage: %v", err)
	}
	service := InitBackup(storage)

	counter := metrics.NewCounter("test_counter")
	counter.SetValue(42)
	_ = storage.Counters().Set("test_counter", counter)

	gauge := metrics.NewGauge("test_gauge")
	gauge.SetValue(3.14)
	_ = storage.Gauges().Set("test_gauge", gauge)

	err = service.WriteBackup()
	if err != nil {
		t.Errorf("WriteBackup() error = %v", err)
	}
}

func TestMetricBackupStorage_NeedSync(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", 0, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Close()

	storage, err := NewMetricBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	if !storage.NeedSync() {
		t.Error("NeedSync() should return true for interval=0")
	}
}

func TestMetricBackupStorage_SetWithSync(t *testing.T) {
	config, err := NewConfig(common.TypeModeTest, "./test", 0, true)
	if err != nil {
		t.Fatalf("Failed to create backup config: %v", err)
	}
	defer config.Close()

	storage, err := NewMetricBackupStorage(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	counter := metrics.NewCounter("test")
	counter.SetValue(100)

	err = storage.Counters().Set("test", counter)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	if err := config.Reader.Reset(); err != nil {
		t.Fatalf("Failed to reset reader: %v", err)
	}

	metricsData, err := config.MetricsHandler.Read()
	if err != nil {
		t.Fatalf("Failed to read from backup: %v", err)
	}

	if len(metricsData) != 1 {
		t.Errorf("Expected 1 metric in backup, got %d", len(metricsData))
	}

	if len(metricsData) > 0 && metricsData[0].ID != "test" {
		t.Errorf("Expected metric ID 'test', got '%s'", metricsData[0].ID)
	}
}
