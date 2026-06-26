package main

import (
	"context"
	"testing"
	"time"

	"sys-metrics/internal/common"
	agentcfg "sys-metrics/internal/config/agent"
)

func TestInitAgent(t *testing.T) {
	opt := &agentcfg.Options{
		Host:           "127.0.0.1",
		Port:           "8080",
		Mode:           common.TypeModeDevelopment,
		PollInterval:   100 * time.Millisecond,
		ReportInterval: 100 * time.Millisecond,
		RateLimit:      1,
	}
	a, err := initAgent(opt, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if a == nil || a.ServerAddr == "" {
		t.Fatalf("agent: %#v", a)
	}
	_ = a.Logger.Sync()
}

func TestInitAgent_invalidCryptoKeyFile(t *testing.T) {
	opt := &agentcfg.Options{
		Host:           "127.0.0.1",
		Port:           "8080",
		Mode:           common.TypeModeDevelopment,
		PollInterval:   100 * time.Millisecond,
		ReportInterval: 100 * time.Millisecond,
		RateLimit:      1,
		CryptoKeyPath:  "/this/path/does/not/exist.pem",
	}
	_, err := initAgent(opt, context.Background())
	if err == nil {
		t.Fatal("expected error from NewRequestEncryptor")
	}
}

func TestInitAgent_grpcTransport(t *testing.T) {
	opt := &agentcfg.Options{
		Host:            "127.0.0.1",
		Port:            "8080",
		Mode:            common.TypeModeDevelopment,
		PollInterval:    100 * time.Millisecond,
		ReportInterval:  100 * time.Millisecond,
		RateLimit:       1,
		ReportTransport: common.ReportTransportGRPC,
	}
	a, err := initAgent(opt, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if a == nil {
		t.Fatal("nil agent")
	}
	_ = a.Close(context.Background())
	_ = a.Logger.Sync()
}

func TestGetAgentLocalIpV4(t *testing.T) {
	ip, err := getAgentLocalIpV4()
	if err != nil {
		t.Fatal(err)
	}
	if ip == "" {
		t.Fatal("empty ip")
	}
}

func TestRun_short(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runWithContext(ctx, []string{
			"-m", common.TypeModeDevelopment,
			"-p", "1",
			"-r", "1",
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not finish")
	}
}
