package main

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	"testing"
)

func TestRun_invalidMode(t *testing.T) {
	_, err := run(context.Background(), []string{"-m", "invalid-mode"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInitGRPCApp(t *testing.T) {
	o, err := server.NewOption([]string{"-m", common.TypeModeDevelopment})
	if err != nil {
		t.Fatal(err)
	}

	ag, err := initGRPCApp(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if ag.app == nil || ag.grpc == nil {
		t.Fatal("expected initialized app")
	}
	_ = ag.grpc.Shutdown(context.Background())
	_ = ag.app.Service.Close(context.Background())
}

func TestRun_startsServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ag, err := run(ctx, []string{"-m", common.TypeModeDevelopment})
	if err != nil {
		t.Fatal(err)
	}
	if ag == nil {
		t.Fatal("nil app")
	}
	_ = ag.grpc.Shutdown(context.Background())
	_ = ag.app.Service.Close(context.Background())
}
