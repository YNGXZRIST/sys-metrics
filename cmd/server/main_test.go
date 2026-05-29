package main

import (
	"context"
	"sys-metrics/internal/app"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/handler"
	"sys-metrics/internal/secure"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"testing"
	"time"
)

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

func bootstrapForHandlerTest(t *testing.T) *app.App {
	t.Helper()
	a, err := app.Bootstrap(context.Background(), &app.Option{
		Mode:              common.TypeModeDevelopment,
		BackupStoragePath: t.TempDir(),
		StoreInterval:     time.Minute,
		Restore:           false,
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestInitHandler(t *testing.T) {
	base := bootstrapForHandlerTest(t)
	dec, err := secure.NewRequestDecryptor("")
	if err != nil {
		t.Fatal(err)
	}
	h := handler.NewHandler(handler.InitProperties{
		Logger:           base.Logger,
		RequestDecryptor: dec,
		MetricService:    serviceMetrics.NewService(nil),
	})
	if h == nil {
		t.Fatal("nil handler")
	}
	_ = base.Logger.Sync()
}

func TestInitHandler_withAuthenticator(t *testing.T) {
	base := bootstrapForHandlerTest(t)
	dec, err := secure.NewRequestDecryptor("")
	if err != nil {
		t.Fatal(err)
	}
	k := "secret"
	h := handler.NewHandler(handler.InitProperties{
		Logger:           base.Logger,
		Authenticator:    initAuthenticator(&server.Options{HashKey: &k}),
		RequestDecryptor: dec,
		MetricService:    serviceMetrics.NewService(nil),
	})
	if h == nil || h.Authenticator == nil {
		t.Fatal("expected handler with auth")
	}
	_ = base.Logger.Sync()
}

func TestRun_development(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := run(ctx, []string{"-m", common.TypeModeDevelopment})
	if err != nil {
		t.Fatal(err)
	}
	if a == nil || a.App == nil || a.h == nil {
		t.Fatal("expected initialized HTTP app")
	}
	if srv, ok := a.App.Server.(interface{ Shutdown(context.Context) error }); ok {
		_ = srv.Shutdown(context.Background())
	}
	_ = a.App.Service.Close(context.Background())
}
