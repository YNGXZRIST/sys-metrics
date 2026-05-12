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
