package app

import (
	"context"
	"errors"
	"net"
	"strings"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/repository/memory"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func testApp(t *testing.T, opts *server.Options) *App {
	t.Helper()
	return &App{opts: opts}
}

func TestInitDB_emptyDSN(t *testing.T) {
	a := testApp(t, &server.Options{DNS: ""})
	_, err := a.initDB()
	if err == nil || !strings.Contains(err.Error(), "not set") {
		t.Fatalf("initDB: %v", err)
	}
}

func TestIsDSNSet(t *testing.T) {
	if testApp(t, &server.Options{DNS: "postgres://x"}).isDSNSet() != true {
		t.Fatal()
	}
	if testApp(t, &server.Options{}).isDSNSet() {
		t.Fatal()
	}
}

func TestInitLogger(t *testing.T) {
	a := testApp(t, &server.Options{Mode: common.TypeModeDevelopment})
	lg, err := a.initLogger()
	if err != nil {
		t.Fatal(err)
	}
	_ = lg.Sync()
}

func TestInitBackupConfig(t *testing.T) {
	a := testApp(t, &server.Options{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
	})
	cfg, err := a.initBackupConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("nil config")
	}
}

func TestInitStorage_memory(t *testing.T) {
	a := testApp(t, &server.Options{
		Mode:              common.TypeModeDevelopment,
		DNS:               "",
		Restore:           false,
		StoreInterval:     time.Minute,
		BackupStoragePath: t.TempDir(),
	})
	if err := a.initStorage(); err != nil {
		t.Fatal(err)
	}
	if a.db != nil {
		t.Fatal("expected no DB conn")
	}
	if a.needRestore {
		t.Fatal("expected no restore flag")
	}
	if a.service == nil {
		t.Fatal("nil service")
	}
	_ = a.service.Close(context.Background())
}

func TestInitStorage_fileBackup(t *testing.T) {
	a := testApp(t, &server.Options{
		Mode:              common.TypeModeTest,
		DNS:               "",
		Restore:           true,
		StoreInterval:     time.Minute,
		BackupStoragePath: t.TempDir(),
	})
	if err := a.initStorage(); err != nil {
		t.Fatal(err)
	}
	defer a.backupConfig.Cleanup()
	defer a.service.Close(context.Background())

	if a.db != nil {
		t.Fatal("expected no DB conn")
	}
	if a.backupConfig == nil {
		t.Fatal("expected backup config")
	}
	if !a.needRestore {
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
	a := testApp(t, &server.Options{})
	a.service = &readBackupFailSvc{Service: memory.NewService()}
	err := a.restoreFromBackup(context.Background())
	if err == nil || !strings.Contains(err.Error(), "read backup") {
		t.Fatalf("err = %v", err)
	}
}

func TestIsDatabaseConnected_nil(t *testing.T) {
	if isDatabaseConnected(nil) {
		t.Fatal("nil conn should be disconnected")
	}
}

func TestRestoreFromBackup_memory(t *testing.T) {
	a := testApp(t, &server.Options{})
	a.service = memory.NewService()
	if err := a.restoreFromBackup(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestStartBackupRoutine_memory(t *testing.T) {
	a := testApp(t, &server.Options{Mode: common.TypeModeDevelopment})
	lg, err := a.initLogger()
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
		service: memory.NewService(),
	}

	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestInitMetricsObserver(t *testing.T) {
	a := testApp(t, &server.Options{Mode: common.TypeModeDevelopment})
	obs, err := a.initMetricsObserver(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if obs == nil {
		t.Fatal("nil observer")
	}
}

func TestBootstrap_memory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o := &server.Options{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
	}
	a, err := Bootstrap(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	if a.metricsService == nil {
		t.Fatal("nil MetricsService")
	}
	if a.logger == nil {
		t.Fatal("nil Logger")
	}
	_ = a.Close(context.Background())
}

func TestBootstrap_trustedSubnet(t *testing.T) {
	ctx := context.Background()
	a, err := Bootstrap(ctx, &server.Options{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
		TrustedSubnetMask: "127.0.0.0/8",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.ipNet == nil {
		t.Fatal("expected IpNet")
	}
	_ = a.Close(ctx)
}

func TestShutdownGRPCServer(t *testing.T) {
	srv := grpc.NewServer()
	w := &shutdownGRPCServer{server: srv}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Serve(lis) }()
	if err := w.shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestBootstrap_withAuditFile(t *testing.T) {
	ctx := context.Background()
	a, err := Bootstrap(ctx, &server.Options{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
		AuditFile:         t.TempDir() + "/audit.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.metricsService == nil {
		t.Fatal("expected metrics service")
	}
	_ = a.service.Close(ctx)
}

func TestInitStorage_invalidDSN(t *testing.T) {
	a := testApp(t, &server.Options{
		Mode:              common.TypeModeDevelopment,
		DNS:               "postgres://127.0.0.1:1/nodb?sslmode=disable&connect_timeout=1",
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
	})
	if err := a.initStorage(); err == nil {
		t.Fatal("expected migrate/init error")
	}
}

func TestBootstrap_invalidTrustedSubnet(t *testing.T) {
	_, err := Bootstrap(context.Background(), &server.Options{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
		TrustedSubnetMask: "invalid",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInitAuthenticator_nilKey(t *testing.T) {
	if initAuthenticator(&server.Options{}) != nil {
		t.Fatal("expected nil authenticator")
	}
}

func TestInitAuthenticator_withKey(t *testing.T) {
	k := "secret"
	if initAuthenticator(&server.Options{HashKey: &k}) == nil {
		t.Fatal("expected authenticator")
	}
}

func TestParseTrustedSubnet_invalid(t *testing.T) {
	a := testApp(t, &server.Options{TrustedSubnetMask: "not-a-cidr"})
	if _, err := a.parseTrustedSubnet(); err == nil {
		t.Fatal("expected error")
	}
}
