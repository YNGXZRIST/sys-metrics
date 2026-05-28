package server

import (
	"sys-metrics/internal/common"
	"testing"
	"time"
)

func TestParseArgs_defaults(t *testing.T) {
	opt := new(Options)

	err := opt.parseArgs([]string{})
	if err != nil {
		t.Fatal(err)
	}

	err = applyDefaults(opt, common.ServerHTTP)
	if err != nil {
		t.Fatal(err)
	}

	if opt.Host != "localhost" || opt.Port != "8080" {
		t.Fatalf("host:port = %q:%q", opt.Host, opt.Port)
	}

	if opt.Mode != common.TypeModeDefault {
		t.Fatalf("mode = %q", opt.Mode)
	}
}

func TestParseArgs_hashKey(t *testing.T) {
	opt := new(Options)

	err := opt.parseArgs([]string{"-k", "secret"})
	if err != nil {
		t.Fatal(err)
	}

	if opt.HashKey == nil || *opt.HashKey != "secret" {
		t.Fatalf("HashKey = %v", opt.HashKey)
	}
}

func TestParseArgs_cryptoKey(t *testing.T) {
	opt := new(Options)

	err := opt.parseArgs([]string{
		"-crypto-key",
		"/path/to/key.pem",
	})
	if err != nil {
		t.Fatal(err)
	}

	if opt.CryptoKeyPath != "/path/to/key.pem" {
		t.Fatalf("CryptoKeyPath = %q", opt.CryptoKeyPath)
	}
}

func TestOptions_SetHostPort(t *testing.T) {
	var opt Options

	opt.SetHostPort("example.com", "9090")

	if opt.Host != "example.com" || opt.Port != "9090" {
		t.Fatalf("SetHostPort: %q:%q", opt.Host, opt.Port)
	}
}

func TestNewOption_development(t *testing.T) {
	for _, k := range []string{
		"ADDRESS",
		"GRPC_ADDRESS",
		"STORE_INTERVAL",
		"KEY",
		"MODE",
		"STORE_FILE",
		"DATABASE_DSN",
		"AUDIT_FILE",
		"AUDIT_URL",
		"RESTORE",
		"CRYPTO_KEY",
		"CONFIG",
	} {
		t.Setenv(k, "")
	}

	opt, err := NewOption(common.ServerHTTP, []string{
		"-m",
		common.TypeModeDevelopment,
	})

	if err != nil {
		t.Fatal(err)
	}

	if opt.Mode != common.TypeModeDevelopment {
		t.Fatalf("mode %q", opt.Mode)
	}

	if opt.Host != "localhost" || opt.Port != "8080" {
		t.Fatalf("addr %s:%s", opt.Host, opt.Port)
	}
}

func TestOptions_parseEnv_storeInterval(t *testing.T) {
	for _, k := range []string{
		"ADDRESS",
		"STORE_INTERVAL",
		"KEY",
		"MODE",
		"STORE_FILE",
		"DATABASE_DSN",
		"AUDIT_FILE",
		"AUDIT_URL",
		"RESTORE",
		"CRYPTO_KEY",
		"CONFIG",
	} {
		t.Setenv(k, "")
	}

	t.Setenv("STORE_INTERVAL", "42")

	var opt Options

	if err := opt.parseEnv(); err != nil {
		t.Fatal(err)
	}

	if opt.StoreIntervalSec == nil {
		t.Fatal("StoreIntervalSec is nil")
	}

	if opt.StoreInterval != 42*time.Second {
		t.Fatalf("StoreInterval = %v", opt.StoreInterval)
	}
}

func TestNewOption_withAddressEnv(t *testing.T) {
	for _, k := range []string{
		"ADDRESS",
		"GRPC_ADDRESS",
		"STORE_INTERVAL",
		"KEY",
		"MODE",
		"STORE_FILE",
		"DATABASE_DSN",
		"AUDIT_FILE",
		"AUDIT_URL",
		"RESTORE",
		"CRYPTO_KEY",
		"CONFIG",
	} {
		t.Setenv(k, "")
	}

	t.Setenv("ADDRESS", "192.168.0.2:6000")

	opt, err := NewOption(common.ServerGRPC, []string{"-m", common.TypeModeDevelopment})

	if err != nil {
		t.Fatal(err)
	}

	if opt.Host != "192.168.0.2" || opt.Port != "6000" {
		t.Fatalf("host:port = %s:%s", opt.Host, opt.Port)
	}
}

func TestNewOption_grpcAddressWhenBothSet(t *testing.T) {
	for _, k := range []string{
		"ADDRESS",
		"GRPC_ADDRESS",
		"STORE_INTERVAL",
		"KEY",
		"MODE",
		"STORE_FILE",
		"DATABASE_DSN",
		"AUDIT_FILE",
		"AUDIT_URL",
		"RESTORE",
		"CRYPTO_KEY",
		"CONFIG",
	} {
		t.Setenv(k, "")
	}

	t.Setenv("ADDRESS", "localhost:8080")
	t.Setenv("GRPC_ADDRESS", "192.168.0.3:7000")

	optGRPC, err := NewOption(common.ServerGRPC, []string{"-m", common.TypeModeDevelopment})
	if err != nil {
		t.Fatal(err)
	}
	if optGRPC.Host != "192.168.0.3" || optGRPC.Port != "7000" {
		t.Fatalf("grpc mode host:port = %s:%s", optGRPC.Host, optGRPC.Port)
	}

	optHTTP, err := NewOption(common.ServerHTTP, []string{"-m", common.TypeModeDevelopment})
	if err != nil {
		t.Fatal(err)
	}
	if optHTTP.Host != "localhost" || optHTTP.Port != "8080" {
		t.Fatalf("http mode host:port = %s:%s", optHTTP.Host, optHTTP.Port)
	}
}

func TestListenAddressForMode(t *testing.T) {
	httpAddr := "localhost:8080"
	grpcAddr := "localhost:9090"
	opt := &Options{
		ServerAddress: &httpAddr,
		GRPCAddress:   &grpcAddr,
	}

	if got := opt.listenAddressForMode(common.ServerGRPC); got != grpcAddr {
		t.Fatalf("grpc mode = %q, want %q", got, grpcAddr)
	}
	if got := opt.listenAddressForMode(common.ServerHTTP); got != httpAddr {
		t.Fatalf("http mode = %q, want %q", got, httpAddr)
	}
}
