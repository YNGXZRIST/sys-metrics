package main

import (
	"context"
	"errors"
	"strings"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/repository/memory"
	"sys-metrics/internal/secure"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"testing"
	"time"
)

func TestInitDB_emptyDSN(t *testing.T) {
	_, err := initDB(&server.Options{DNS: ""})
	if err == nil || !strings.Contains(err.Error(), "not set") {
		t.Fatalf("initDB: %v", err)
	}
}

func TestIsDSNSet(t *testing.T) {
	if isDSNSet(&server.Options{DNS: "postgres://x"}) != true {
		t.Fatal()
	}
	if isDSNSet(&server.Options{}) {
		t.Fatal()
	}
}

func TestInitLogger(t *testing.T) {
	lg, err := initLogger(common.TypeModeDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	_ = lg.Sync()
}

func TestInitBackupConfig(t *testing.T) {
	o := &server.Options{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
	}
	cfg, err := initBackupConfig(o)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("nil config")
	}
}

func TestCreateService_memory(t *testing.T) {
	o := &server.Options{
		Mode:              common.TypeModeDevelopment,
		DNS:               "",
		Restore:           false,
		StoreInterval:     time.Minute,
		BackupStoragePath: t.TempDir(),
		Host:              "localhost",
		Port:              "8080",
	}
	svc, conn, _, need, err := createService(o)
	if err != nil {
		t.Fatal(err)
	}
	if conn != nil {
		t.Fatal("expected no DB conn")
	}
	if need {
		t.Fatal("expected no restore flag")
	}
	if svc == nil {
		t.Fatal("nil service")
	}
	_ = svc.Close(context.Background())
}

func TestCreateService_fileBackup(t *testing.T) {
	o := &server.Options{
		Mode:              common.TypeModeTest,
		DNS:               "",
		Restore:           true,
		StoreInterval:     time.Minute,
		BackupStoragePath: t.TempDir(),
		Host:              "localhost",
		Port:              "8080",
	}
	svc, conn, cfg, need, err := createService(o)
	if err != nil {
		t.Fatal(err)
	}
	defer cfg.Cleanup()
	defer svc.Close(context.Background())

	if conn != nil {
		t.Fatal("expected no DB conn")
	}
	if cfg == nil {
		t.Fatal("expected backup config")
	}
	if !need {
		t.Fatal("expected restore flag for enabled backup")
	}
}

type readBackupFailSvc struct {
	*memory.Service
}

func (readBackupFailSvc) ReadBackup(context.Context) error {
	return errors.New("read backup failed")
}

func TestRestoreFromBackup_error(t *testing.T) {
	s := &readBackupFailSvc{Service: memory.NewService()}
	err := restoreFromBackup(context.Background(), s)
	if err == nil || !strings.Contains(err.Error(), "read backup") {
		t.Fatalf("err = %v", err)
	}
}

func TestRun_invalidMode(t *testing.T) {
	_, err := run(context.Background(), []string{"-m", "not-a-valid-mode"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_invalidAddress(t *testing.T) {
	_, err := run(context.Background(), []string{"-a", "no-port-here"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInitAuthenticator_nilKey(t *testing.T) {
	a := initAuthenticator(&server.Options{})
	if a != nil {
		t.Fatalf("want nil authenticator, got %T", a)
	}
}

func TestInitAuthenticator_withKey(t *testing.T) {
	k := "secret"
	a := initAuthenticator(&server.Options{HashKey: &k})
	if a == nil {
		t.Fatal("expected authenticator")
	}
}

func TestIsDatabaseConnected_nil(t *testing.T) {
	if isDatabaseConnected(nil) {
		t.Fatal("nil conn should be disconnected")
	}
}

func TestRestoreFromBackup_memory(t *testing.T) {
	s := memory.NewService()
	if err := restoreFromBackup(context.Background(), s); err != nil {
		t.Fatal(err)
	}
}

func TestStartBackupRoutine_memory(t *testing.T) {
	lg, err := initLogger(common.TypeModeDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	s := memory.NewService()
	ctx, cancel := context.WithCancel(context.Background())
	startBackupRoutine(ctx, s, lg)
	cancel()
	time.Sleep(20 * time.Millisecond)
	_ = lg.Sync()
}

func TestAppClose(t *testing.T) {
	app := &App{
		Service: memory.NewService(),
	}

	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestInitHandler(t *testing.T) {
	lg, err := initLogger(common.TypeModeDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := secure.NewRequestDecryptor("")
	if err != nil {
		t.Fatal(err)
	}
	h := initHandler(nil, lg, nil, dec, nil, serviceMetrics.NewService(nil))
	if h == nil {
		t.Fatal("nil handler")
	}
	_ = lg.Sync()
}

func TestInitHandler_withAuthenticator(t *testing.T) {
	lg, err := initLogger(common.TypeModeDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := secure.NewRequestDecryptor("")
	if err != nil {
		t.Fatal(err)
	}
	k := "secret"
	h := initHandler(nil, lg, initAuthenticator(&server.Options{HashKey: &k}), dec, nil, serviceMetrics.NewService(nil))
	if h == nil || h.Auth == nil {
		t.Fatal("expected handler with auth")
	}
	_ = lg.Sync()
}

func TestInitMetricsObserver(t *testing.T) {
	o := &server.Options{
		Mode: common.TypeModeDevelopment,
	}
	obs, err := initMetricsObserver(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if obs == nil {
		t.Fatal("nil observer")
	}
}
