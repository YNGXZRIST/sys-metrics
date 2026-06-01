package server

import (
	"os"
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

	err = applyDefaults(opt)
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

	opt, err := NewOption([]string{
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

	opt, err := NewOption([]string{"-m", common.TypeModeDevelopment})

	if err != nil {
		t.Fatal(err)
	}

	if opt.Host != "192.168.0.2" || opt.Port != "6000" {
		t.Fatalf("host:port = %s:%s", opt.Host, opt.Port)
	}
}

func TestParseConfig(t *testing.T) {
	path := t.TempDir() + "/server.json"
	content := `{
		"address": "127.0.0.1:9090",
		"store_interval": "42s",
		"store_file": "/tmp/backups",
		"database_dsn": "postgres://localhost/db",
		"restore": false,
		"trusted_subnet": "127.0.0.0/8"
	}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opt := &Options{}
	if err := opt.ParseConfig(path); err != nil {
		t.Fatal(err)
	}
	if opt.ServerAddressHTTP == nil || *opt.ServerAddressHTTP != "127.0.0.1:9090" {
		t.Fatalf("address = %v", opt.ServerAddressHTTP)
	}
	if opt.BackupStoragePath != "/tmp/backups" {
		t.Fatalf("backup path %q", opt.BackupStoragePath)
	}
	if opt.DNS != "postgres://localhost/db" {
		t.Fatalf("dsn %q", opt.DNS)
	}
	if opt.Restore {
		t.Fatal("expected restore=false from config")
	}
	if opt.TrustedSubnetMask != "127.0.0.0/8" {
		t.Fatalf("trusted subnet %q", opt.TrustedSubnetMask)
	}
}

func TestNewOption_trustedSubnetFlag(t *testing.T) {
	for _, k := range []string{"ADDRESS", "MODE", "CONFIG", "TRUSTED_SUBNET"} {
		t.Setenv(k, "")
	}

	opt, err := NewOption([]string{
		"-m", common.TypeModeDevelopment,
		"-t", "10.0.0.0/8",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opt.TrustedSubnetMask != "10.0.0.0/8" {
		t.Fatalf("trusted subnet %q", opt.TrustedSubnetMask)
	}
}

func TestNewOption_allFlags(t *testing.T) {
	for _, k := range []string{
		"ADDRESS", "STORE_INTERVAL", "KEY", "MODE", "STORE_FILE",
		"DATABASE_DSN", "AUDIT_FILE", "AUDIT_URL", "RESTORE", "CRYPTO_KEY", "CONFIG",
	} {
		t.Setenv(k, "")
	}

	opt, err := NewOption([]string{
		"-m", common.TypeModeDevelopment,
		"-a", "127.0.0.1:8080",
		"-a-grpc", "127.0.0.1:9090",
		"-i", "60",
		"-d", "postgres://localhost/db",
		"-k", "secret",
		"-f", "/tmp/backup",
		"-audit-file", "/tmp/audit",
		"-audit-url", "http://audit.local",
		"-crypto-key", "/tmp/key.pem",
		"-t", "192.168.0.0/16",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opt.Host != "127.0.0.1" || opt.Port != "8080" {
		t.Fatalf("http addr %s:%s", opt.Host, opt.Port)
	}
	if opt.HostGRPC != "127.0.0.1" || opt.PortGRPC != "9090" {
		t.Fatalf("grpc addr %s:%s", opt.HostGRPC, opt.PortGRPC)
	}
	if opt.StoreInterval != 60*time.Second {
		t.Fatalf("interval %v", opt.StoreInterval)
	}
	if opt.DNS != "postgres://localhost/db" {
		t.Fatalf("dsn %q", opt.DNS)
	}
	if opt.HashKey == nil || *opt.HashKey != "secret" {
		t.Fatalf("hash key %v", opt.HashKey)
	}
	if opt.BackupStoragePath != "/tmp/backup" {
		t.Fatalf("backup %q", opt.BackupStoragePath)
	}
	if !opt.Restore {
		t.Fatal("restore should default to true")
	}
	if opt.AuditFile != "/tmp/audit" || opt.AuditURL != "http://audit.local" {
		t.Fatalf("audit cfg")
	}
	if opt.CryptoKeyPath != "/tmp/key.pem" {
		t.Fatalf("crypto %q", opt.CryptoKeyPath)
	}
	if opt.TrustedSubnetMask != "192.168.0.0/16" {
		t.Fatalf("subnet %q", opt.TrustedSubnetMask)
	}
}

func TestNewConfig_addresses(t *testing.T) {
	cfg := NewConfig(SchemeHTTP, "localhost", "8080", nil, nil)
	if cfg.ServerAddr() != "http://localhost:8080" {
		t.Fatalf("ServerAddr = %q", cfg.ServerAddr())
	}
	if cfg.InternalAddr() != "localhost:8080" {
		t.Fatalf("InternalAddr = %q", cfg.InternalAddr())
	}
}

func TestStringPtrValue(t *testing.T) {
	if stringPtrValue(nil) != "" {
		t.Fatal("nil ptr")
	}
	s := "value"
	if stringPtrValue(&s) != "value" {
		t.Fatal("value mismatch")
	}
}

func TestParseConfigPath_fromEnv(t *testing.T) {
	t.Setenv("CONFIG", "/etc/metrics.json")
	path, err := parseConfigPath(nil)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/etc/metrics.json" {
		t.Fatalf("path = %q", path)
	}
}

func TestNewOption_invalidMode(t *testing.T) {
	for _, k := range []string{"ADDRESS", "MODE", "CONFIG"} {
		t.Setenv(k, "")
	}
	_, err := NewOption([]string{"-m", "staging"})
	if err == nil {
		t.Fatal("expected error")
	}
}
