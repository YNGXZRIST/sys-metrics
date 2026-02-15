package main

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/repository/memory"
	"sys-metrics/internal/repository/metrics"
	"testing"
	"time"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"test mode", common.TypeModeTest, false},
		{"development mode", common.TypeModeDevelopment, false},
		{"invalid mode", "invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := initLogger(tt.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("initLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("initLogger() returned nil logger")
			}
		})
	}
}

func TestIsDSNSet(t *testing.T) {
	tests := []struct {
		name string
		opt  *server.Options
		want bool
	}{
		{"empty DSN", &server.Options{DNS: ""}, false},
		{"DSN set", &server.Options{DNS: "postgres://localhost/db"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDSNSet(tt.opt); got != tt.want {
				t.Errorf("isDSNSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitBackupConfig(t *testing.T) {
	opt := &server.Options{
		Mode:              common.TypeModeTest,
		BackupStoragePath: "./backups",
		StoreInterval:     time.Second * 300,
		Restore:           true,
	}
	cfg, err := initBackupConfig(opt)
	if err != nil {
		t.Fatalf("initBackupConfig() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("initBackupConfig() returned nil config")
	}
	defer cfg.Close()
	if !cfg.Enabled {
		t.Error("initBackupConfig() config.Enabled = false, want true when Restore=true")
	}
}

func TestCreateService_memoryWhenNoDSNAndRestoreFalse(t *testing.T) {
	opt := &server.Options{
		Mode:              common.TypeModeTest,
		BackupStoragePath: "./backups",
		StoreInterval:     time.Second,
		Restore:           false,
		DNS:               "",
	}
	svc, conn, backupCfg, needRestore, err := createService(opt)
	if err != nil {
		t.Fatalf("createService() error = %v", err)
	}
	if svc == nil {
		t.Fatal("createService() returned nil service")
	}
	if conn != nil {
		t.Error("createService() conn should be nil when using memory")
	}
	if backupCfg != nil {
		t.Error("createService() backupCfg should be nil when using memory")
		defer backupCfg.Close()
	}
	if needRestore {
		t.Error("createService() needRestore should be false for memory")
	}
	_ = svc.Close(context.Background())
}

func TestRun_invalidModeReturnsError(t *testing.T) {
	err := run([]string{"-m", "invalid"})
	if err == nil {
		t.Error("run() with invalid mode should return error")
	}
}

func TestMetricsInit(t *testing.T) {
	metrics.Init(memory.NewService())
	if metrics.Gauges() == nil {
		t.Error("Gauges() == nil")
	}
	if metrics.Counters() == nil {
		t.Error("Counters() == nil")
	}
}
