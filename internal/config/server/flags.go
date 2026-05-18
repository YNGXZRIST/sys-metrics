package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config"
	"sys-metrics/internal/errors/labelerrors"
	"time"

	"github.com/caarlos0/env/v11"
)

// generate:reset

// Options holds server CLI flags and env vars: address, mode, DSN, backup, hash key, audit.
type Options struct {
	ServerAddress     *string `json:"address" env:"ADDRESS"`
	StoreIntervalSec  *int    `env:"STORE_INTERVAL" default:"300"`
	HashKey           *string `env:"KEY"`
	Mode              string  `env:"MODE"`
	Host              string
	Port              string
	StoreIntervalJSON string `json:"store_interval"`
	BackupStoragePath string `json:"store_file" env:"STORE_FILE" envDefault:"./backups"`
	DNS               string `json:"database_dsn" env:"DATABASE_DSN"`
	AuditFile         string `env:"AUDIT_FILE"`
	AuditURL          string `env:"AUDIT_URL"`
	CryptoKeyPath     string `json:"crypto_key" env:"CRYPTO_KEY"`
	ConfigFilePath    string
	StoreInterval     time.Duration
	Restore           bool `json:"restore" env:"RESTORE" envDefault:"true"`
}

// NewOption parses argv, environment and config, validates mode, and returns Options.
func NewOption(args []string) (*Options, error) {
	configPath, err := parseConfigPath(args)
	if err != nil {
		return nil, fmt.Errorf("error getting config path: %w", err)
	}

	opt := new(Options)

	if configPath != "" {
		err = opt.ParseConfig(configPath)
		if err != nil {
			return nil, labelerrors.NewLabelError("CONFIG", fmt.Errorf("error parsing config: %w", err))
		}
	}
	err = opt.parseEnv()
	if err != nil {
		return nil, labelerrors.NewLabelError("ENV", fmt.Errorf("error parsing env: %w", err))
	}
	err = opt.parseArgs(args)
	if err != nil {
		return nil, labelerrors.NewLabelError(
			"ARGS",
			fmt.Errorf("error parsing args: %w", err),
		)
	}
	err = applyDefaults(opt)
	if err != nil {
		return nil, fmt.Errorf("error applying defaults: %w", err)
	}

	err = config.ValidateMode(opt.Mode)
	if err != nil {
		return nil, labelerrors.NewLabelError(
			"MODE",
			fmt.Errorf("error validation mode: %w", err),
		)
	}

	return opt, nil
}

// ParseConfig parsing json config. Lowest priority
func (opt *Options) ParseConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	var cfg Options

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&cfg)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	if cfg.StoreIntervalJSON != "" {
		duration, err := time.ParseDuration(cfg.StoreIntervalJSON)
		if err != nil {
			return fmt.Errorf("error parsing store interval: %w", err)
		}

		cfg.StoreInterval = duration
	}

	mergeOptions(opt, &cfg)

	return nil
}

// parseEnv parsing os.ENV. medium priority
func (opt *Options) parseEnv() error {
	cfg := new(Options)

	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("error parsing env: %w", err)
	}
	if cfg.StoreIntervalSec != nil {
		cfg.StoreInterval =
			time.Duration(*cfg.StoreIntervalSec) * time.Second
	}
	mergeOptions(opt, cfg)

	return nil
}

// parseArgs parsiong os.Args. Highest priority
func (opt *Options) parseArgs(args []string) error {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)

	var (
		serverAddr string
		mode       string
		interval   int
		hashKey    string
		dns        string
		backup     string
		restore    bool
		auditFile  string
		auditURL   string
		cryptoKey  string
	)

	flags.StringVar(&serverAddr, "a", "", "Address of the server")
	flags.StringVar(&mode, "m", "", "Server mode. Possible values: production, development")
	flags.IntVar(&interval, "i", 0, "Storage interval in seconds")
	flags.StringVar(&dns, "d", "", "Database DSN for backup storage")
	flags.StringVar(&hashKey, "k", "", "Hash key")
	flags.StringVar(&backup, "f", "", "Backup storage path")
	flags.BoolVar(&restore, "r", false, "Restore backups")
	flags.StringVar(&auditFile, "audit-file", "", "File to store audit log")
	flags.StringVar(&auditURL, "audit-url", "", "URL to store audit log in remote server")
	flags.StringVar(&cryptoKey, "crypto-key", "", "crypto key for decoding request")
	flags.StringVar(&opt.ConfigFilePath, "config", "", "config file path")
	flags.StringVar(&opt.ConfigFilePath, "c", "", "config file path (shorthand)")

	err := flags.Parse(args)
	if err != nil {
		return err
	}

	visited := map[string]bool{}

	flags.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	if visited["a"] {
		opt.ServerAddress = &serverAddr
	}

	if visited["m"] {
		opt.Mode = mode
	}
	if visited["i"] {
		opt.StoreInterval = time.Duration(interval) * time.Second
		opt.StoreIntervalSec = new(interval)
	}
	if visited["d"] {
		opt.DNS = dns
	}
	if visited["k"] {
		opt.HashKey = &hashKey
	}
	if visited["f"] {
		opt.BackupStoragePath = backup
	}
	if visited["r"] {
		opt.Restore = restore
	}

	if visited["audit-file"] {
		opt.AuditFile = auditFile
	}

	if visited["audit-url"] {
		opt.AuditURL = auditURL
	}

	if visited["crypto-key"] {
		opt.CryptoKeyPath = cryptoKey
	}

	return nil
}

// SetHostPort implements config.HostPortSetter.
func (opt *Options) SetHostPort(host, port string) {
	opt.Host = host
	opt.Port = port
}

// mergeOptions merge new and source options. Not overriding existed options
func mergeOptions(dst, src *Options) {
	if dst.ServerAddress == nil && src.ServerAddress != nil {
		dst.ServerAddress = src.ServerAddress
	}

	if dst.StoreIntervalSec == nil && src.StoreIntervalSec != nil {
		dst.StoreIntervalSec = src.StoreIntervalSec
		dst.StoreInterval = src.StoreInterval
	}

	if dst.HashKey == nil && src.HashKey != nil {
		dst.HashKey = src.HashKey
	}

	if dst.Mode == "" && src.Mode != "" {
		dst.Mode = src.Mode
	}

	if dst.BackupStoragePath == "" && src.BackupStoragePath != "" {
		dst.BackupStoragePath = src.BackupStoragePath
	}

	if dst.DNS == "" && src.DNS != "" {
		dst.DNS = src.DNS
	}

	if dst.AuditFile == "" && src.AuditFile != "" {
		dst.AuditFile = src.AuditFile
	}

	if dst.AuditURL == "" && src.AuditURL != "" {
		dst.AuditURL = src.AuditURL
	}

	if !dst.Restore && src.Restore {
		dst.Restore = src.Restore
	}

	if dst.CryptoKeyPath == "" && src.CryptoKeyPath != "" {
		dst.CryptoKeyPath = src.CryptoKeyPath
	}
}

// applyDefaults set default fields if not exist
func applyDefaults(opt *Options) error {

	if opt.ServerAddress == nil {
		opt.ServerAddress = new("localhost:8080")
	}

	if opt.Mode == "" {
		opt.Mode = common.TypeModeDefault
	}

	if opt.StoreIntervalSec == nil {
		opt.StoreIntervalSec = new(300)
		opt.StoreInterval = 300 * time.Second
	}

	if opt.BackupStoragePath == "" {
		opt.BackupStoragePath = "./backups"
	}

	if !opt.Restore {
		opt.Restore = true
	}

	err := config.ParseAndSetHostPort(*opt.ServerAddress, opt)
	if err != nil {
		return fmt.Errorf("parsing server address: %w", err)
	}
	return nil
}

// parseConfigPath get config path from os.Args
func parseConfigPath(args []string) (string, error) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	configArgs := config.GetConfigArgs(args)
	var configPath string

	flags.StringVar(&configPath, "config", "", "")
	flags.StringVar(&configPath, "c", "", "")

	err := flags.Parse(configArgs)
	if err != nil {
		return "", err
	}

	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	return configPath, nil
}
