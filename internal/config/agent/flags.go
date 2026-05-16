package agent

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/errors/labelerrors"
	"time"

	"github.com/caarlos0/env/v11"
)

// generate:reset

// Options holds agent CLI flags and env: server address, intervals, mode, key, rate limit.
type Options struct {
	ServerAddress  string `json:"address" env:"ADDRESS"`
	Host           string
	Port           string
	Mode           string `env:"MODE"`
	HashKey        string `env:"KEY"`
	CryptoKeyPath  string `json:"crypto_key" env:"CRYPTO_KEY"`
	ConfigFilePath string
	PollInterval   time.Duration
	ReportInterval time.Duration
	PollSec        int `env:"POLL_INTERVAL"`
	ReportSec      int `env:"REPORT_INTERVAL"`
	RateLimit      int `env:"RATE_LIMIT"`
}

// NewOption parses the agent argv, environment and config. Validates logging mode.
func NewOption(args []string) (*Options, error) {
	configPath, err := parseConfigPath(args)
	if err != nil {
		return nil, fmt.Errorf("error getting config path: %w", err)
	}

	opt := new(Options)
	if configPath != "" {
		err = opt.ParseConfig(configPath)
		if err != nil {
			return nil, labelerrors.NewLabelError("CONFIG", err)
		}
	}
	err = opt.parseEnv()
	if err != nil {
		return nil, labelerrors.NewLabelError("PARSE ENV", err)
	}
	err = opt.parseArgs(args)
	if err != nil {
		return nil, labelerrors.NewLabelError("PARSE ARGS", err)
	}
	err = applyDefaults(opt)
	if err != nil {
		return nil, labelerrors.NewLabelError("APPLY DEFAULTS", err)
	}
	err = config.ValidateMode(opt.Mode)
	if err != nil {
		return nil, labelerrors.NewLabelError("MODE", err)
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

	var cfg struct {
		ServerAddress  string `json:"address"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKeyPath  string `json:"crypto_key"`
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&cfg)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	if opt.ServerAddress == "" && cfg.ServerAddress != "" {
		opt.ServerAddress = cfg.ServerAddress
		err = config.ParseAndSetHostPort(opt.ServerAddress, opt)
		if err != nil {
			return fmt.Errorf("error parsing server address: %w", err)
		}
	}

	if opt.CryptoKeyPath == "" && cfg.CryptoKeyPath != "" {
		opt.CryptoKeyPath = cfg.CryptoKeyPath
	}

	if opt.ReportSec == 0 && cfg.ReportInterval != "" {

		duration, err := time.ParseDuration(cfg.ReportInterval)
		if err != nil {
			return fmt.Errorf("error parsing report interval: %w", err)
		}
		opt.ReportInterval = duration
		opt.ReportSec = int(duration.Seconds())
	}

	if opt.PollSec == 0 && cfg.PollInterval != "" {
		duration, err := time.ParseDuration(
			cfg.PollInterval,
		)
		if err != nil {
			return fmt.Errorf("error parsing poll interval: %w", err)
		}

		opt.PollInterval = duration
		opt.PollSec = int(duration.Seconds())
	}

	return nil
}

// parseEnv parsing os.ENV. medium priority
func (opt *Options) parseEnv() error {
	cfg := new(Options)

	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("error parsing env: %w", err)
	}

	if cfg.PollSec > 0 {
		cfg.PollInterval =
			time.Duration(cfg.PollSec) * time.Second
	}

	if cfg.ReportSec > 0 {
		cfg.ReportInterval =
			time.Duration(cfg.ReportSec) * time.Second
	}

	if cfg.ServerAddress != "" {
		err = config.ParseAndSetHostPort(
			cfg.ServerAddress,
			cfg,
		)
		if err != nil {
			return fmt.Errorf("error parsing server address: %w", err)
		}
	}

	mergeOptions(opt, cfg)

	return nil
}

// parseArgs parsiong os.Args. Highest priority
func (opt *Options) parseArgs(args []string) error {
	flags := flag.NewFlagSet("agent", flag.ContinueOnError)

	var (
		serverAddress string
		reportSec     int
		pollSec       int
		hashKey       string
		mode          string
		rateLimit     int
		cryptoKey     string
	)

	flags.StringVar(
		&serverAddress, "a", "", "Address of agent server")

	flags.IntVar(&reportSec, "r", 0, "Reporting interval in seconds")

	flags.IntVar(&pollSec, "p", 0, "Poll interval in seconds")

	flags.StringVar(&hashKey, "k", "", "Server Hash key")

	flags.StringVar(&mode, "m", "", "Agent mode. Possible values: production, development")

	flags.IntVar(&rateLimit, "l", 0, "agent rate limit")

	flags.StringVar(&cryptoKey, "crypto-key", "", "crypto key for encoding request")
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
		opt.ServerAddress = serverAddress

		err = config.ParseAndSetHostPort(
			opt.ServerAddress,
			opt,
		)
		if err != nil {
			return err
		}
	}

	if visited["r"] {
		opt.ReportSec = reportSec
		opt.ReportInterval =
			time.Duration(reportSec) * time.Second
	}

	if visited["p"] {
		opt.PollSec = pollSec
		opt.PollInterval =
			time.Duration(pollSec) * time.Second
	}

	if visited["k"] {
		opt.HashKey = hashKey
	}

	if visited["m"] {
		opt.Mode = mode
	}

	if visited["l"] {
		opt.RateLimit = rateLimit
	}

	if visited["crypto-key"] {
		opt.CryptoKeyPath = cryptoKey
	}

	return nil
}

// mergeOptions merge new and source options. Not overriding existed options
func mergeOptions(dst, src *Options) {
	if dst.ServerAddress == "" && src.ServerAddress != "" {
		dst.ServerAddress = src.ServerAddress
	}

	if dst.Host == "" && src.Host != "" {

		dst.Host = src.Host
	}

	if dst.Port == "" && src.Port != "" {
		dst.Port = src.Port
	}

	if dst.Mode == "" && src.Mode != "" {
		dst.Mode = src.Mode
	}

	if dst.HashKey == "" && src.HashKey != "" {
		dst.HashKey = src.HashKey
	}

	if dst.PollSec == 0 && src.PollSec > 0 {
		dst.PollSec = src.PollSec
		dst.PollInterval = src.PollInterval
	}

	if dst.ReportSec == 0 && src.ReportSec > 0 {
		dst.ReportSec = src.ReportSec
		dst.ReportInterval = src.ReportInterval
	}

	if dst.RateLimit == 0 && src.RateLimit > 0 {
		dst.RateLimit = src.RateLimit
	}

	if dst.CryptoKeyPath == "" && src.CryptoKeyPath != "" {
		dst.CryptoKeyPath = src.CryptoKeyPath
	}
}

// applyDefaults set default fields if not exist
func applyDefaults(opt *Options) error {
	if opt.ServerAddress == "" {
		opt.ServerAddress = fmt.Sprintf("%v:%v", server.DefaultHost, server.DefaultPort)
	}

	if opt.ReportSec == 0 {
		opt.ReportSec = 10
		opt.ReportInterval = 10 * time.Second
	}

	if opt.PollSec == 0 {
		opt.PollSec = 2
		opt.PollInterval = 2 * time.Second
	}

	if opt.Mode == "" {
		opt.Mode = common.TypeModeDefault
	}

	if opt.RateLimit == 0 {
		opt.RateLimit = 1
	}

	err := config.ParseAndSetHostPort(opt.ServerAddress, opt)
	if err != nil {
		return fmt.Errorf("error parsing server address: %w", err)
	}
	return nil
}

// parseConfigPath get config path from os.Args

func parseConfigPath(args []string) (string, error) {
	flags := flag.NewFlagSet("agent", flag.ContinueOnError)
	configArgs := config.GetConfigArgs(args)
	var configPath string
	flags.StringVar(&configPath, "config", "", "")

	flags.StringVar(&configPath, "c", "", "")

	err := flags.Parse(configArgs)
	if err != nil {
		return "", err
	}

	if configPath == "" {
		path, ok := os.LookupEnv("CONFIG")
		if ok {
			configPath = path
		}
	}

	return configPath, nil
}

// SetHostPort implements config.HostPortSetter.
func (opt *Options) SetHostPort(host, port string) {
	opt.Host = host
	opt.Port = port
}
