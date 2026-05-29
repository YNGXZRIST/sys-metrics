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

func testApp(t *testing.T, opt *Option) *App {
	t.Helper()
	return &App{option: opt}
}

func TestInitDB_emptyDSN(t *testing.T) {
	a := testApp(t, &Option{DNS: ""})
	_, err := a.initDB()
	if err == nil || !strings.Contains(err.Error(), "not set") {
		t.Fatalf("initDB: %v", err)
	}
}

func TestIsDSNSet(t *testing.T) {
	if testApp(t, &Option{DNS: "postgres://x"}).isDSNSet() != true {
		t.Fatal()
	}
	if testApp(t, &Option{}).isDSNSet() {
		t.Fatal()
	}
}

func TestInitLogger(t *testing.T) {
	a := testApp(t, &Option{Mode: common.TypeModeDevelopment})
	lg, err := a.initLogger()
	if err != nil {
		t.Fatal(err)
	}
	_ = lg.Sync()
}

func TestInitBackupConfig(t *testing.T) {
	a := testApp(t, &Option{
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
	a := testApp(t, &Option{
		Mode:              common.TypeModeDevelopment,
		DNS:               "",
		Restore:           false,
		StoreInterval:     time.Minute,
		BackupStoragePath: t.TempDir(),
	})
	if err := a.initStorage(); err != nil {
		t.Fatal(err)
	}
	if a.DB != nil {
		t.Fatal("expected no DB conn")
	}
	if a.needRestore {
		t.Fatal("expected no restore flag")
	}
	if a.Service == nil {
		t.Fatal("nil service")
	}
	_ = a.Service.Close(context.Background())
}

func TestInitStorage_fileBackup(t *testing.T) {
	a := testApp(t, &Option{
		Mode:              common.TypeModeTest,
		DNS:               "",
		Restore:           true,
		StoreInterval:     time.Minute,
		BackupStoragePath: t.TempDir(),
	})
	if err := a.initStorage(); err != nil {
		t.Fatal(err)
	}
	defer a.BackupConfig.Cleanup()
	defer a.Service.Close(context.Background())

	if a.DB != nil {
		t.Fatal("expected no DB conn")
	}
	if a.BackupConfig == nil {
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
	a := testApp(t, &Option{})
	a.Service = &readBackupFailSvc{Service: memory.NewService()}
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
	a := testApp(t, &Option{})
	a.Service = memory.NewService()
	if err := a.restoreFromBackup(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestStartBackupRoutine_memory(t *testing.T) {
	a := testApp(t, &Option{Mode: common.TypeModeDevelopment})
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
		Service: memory.NewService(),
	}

	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestInitMetricsObserver(t *testing.T) {
	a := testApp(t, &Option{Mode: common.TypeModeDevelopment})
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

	o := &Option{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
	}
	a, err := Bootstrap(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	if a.MetricsService == nil {
		t.Fatal("nil MetricsService")
	}
	if a.Logger == nil {
		t.Fatal("nil Logger")
	}
	_ = a.Close(context.Background())
}

func TestOptionFromServer(t *testing.T) {
	hash := "key"
	addr := "localhost:8080"
	o := &server.Options{
		Mode:              common.TypeModeDevelopment,
		ServerAddress:     &addr,
		HashKey:           &hash,
		BackupStoragePath: "/tmp/backups",
		DNS:               "postgres://x",
		AuditFile:         "/tmp/audit",
		AuditURL:          "http://audit",
		Restore:           false,
		TrustedSubnetMask: "127.0.0.0/8",
		StoreInterval:     time.Minute,
	}
	opt := OptionFromServer(o)
	if opt.Mode != o.Mode || opt.DNS != o.DNS || opt.TrustedSubnetMask != o.TrustedSubnetMask {
		t.Fatalf("opt = %+v", opt)
	}
}

func TestBootstrap_trustedSubnet(t *testing.T) {
	ctx := context.Background()
	a, err := Bootstrap(ctx, &Option{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
		TrustedSubnetMask: "127.0.0.0/8",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.IpNet == nil {
		t.Fatal("expected IpNet")
	}
	_ = a.Close(ctx)
}

func TestShutdownGRPCServer(t *testing.T) {
	srv := grpc.NewServer()
	w := &ShutdownGRPCServer{Server: srv}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = w.ListenAndServe(lis) }()
	if err := w.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestAsGRPCServer(t *testing.T) {
	srv := grpc.NewServer()
	app := &App{Server: &ShutdownGRPCServer{Server: srv}}
	got, ok := app.AsGRPCServer()
	if !ok || got == nil {
		t.Fatal("expected GRPCServer")
	}
}

func TestBootstrap_withAuditFile(t *testing.T) {
	ctx := context.Background()
	a, err := Bootstrap(ctx, &Option{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
		AuditFilePath:     t.TempDir() + "/audit.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.MetricsService == nil {
		t.Fatal("expected metrics service")
	}
	_ = a.Service.Close(ctx)
}

func TestInitStorage_invalidDSN(t *testing.T) {
	a := testApp(t, &Option{
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
	_, err := Bootstrap(context.Background(), &Option{
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

func TestParseTrustedSubnet_invalid(t *testing.T) {
	a := testApp(t, &Option{TrustedSubnetMask: "not-a-cidr"})
	if _, err := a.parseTrustedSubnet(); err == nil {
		t.Fatal("expected error")
	}
}
