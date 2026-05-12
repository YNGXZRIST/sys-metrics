package agent

import (
	"flag"
	"fmt"
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
	ServerAddress  string `env:"ADDRESS"`
	Host           string
	Port           string
	Mode           string `env:"MODE"`
	HashKey        string `env:"KEY"`
	PollInterval   time.Duration
	ReportInterval time.Duration
	PollSec        int    `env:"POLL_INTERVAL"`
	ReportSec      int    `env:"REPORT_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKeyPath  string `env:"CRYPTO_KEY"`
}

// SetHostPort implements config.HostPortSetter.
func (opt *Options) SetHostPort(host, port string) {
	opt.Host = host
	opt.Port = port
}

func parseArgs(args []string) (*Options, error) {
	var hashKey string
	flags := flag.NewFlagSet("agent", flag.ContinueOnError)
	opt := new(Options)
	flags.StringVar(&opt.ServerAddress, "a", fmt.Sprintf("%v:%v", server.DefaultHost, server.DefaultPort), "Address of agent server")
	flags.IntVar(&opt.ReportSec, "r", 10, "Reporting interval in seconds")
	flags.IntVar(&opt.PollSec, "p", 2, "Poll interval in seconds")
	flags.StringVar(&hashKey, "k", "", "Server Hash key")
	flags.StringVar(&opt.Mode, "m", common.TypeModeDefault, "Agent mode. Possible values: production, development")
	flags.IntVar(&opt.RateLimit, "l", 1, "agent rate limit")
	flags.StringVar(&opt.CryptoKeyPath, "crypto-key", "", "crypto key for encoding request")
	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}

	if hashKey != "" {
		opt.HashKey = hashKey
	}
	opt.PollInterval = time.Duration(opt.PollSec) * time.Second
	opt.ReportInterval = time.Duration(opt.ReportSec) * time.Second
	err = config.ParseAndSetHostPort(opt.ServerAddress, opt)
	if err != nil {
		return nil, err
	}
	return opt, nil
}

func (opt *Options) parseEnv() error {
	err := env.Parse(opt)
	if err != nil {
		return fmt.Errorf("error parsing env: %w", err)
	}
	if opt.PollSec > 0 {
		opt.PollInterval = time.Duration(opt.PollSec) * time.Second
	}
	if opt.ReportSec > 0 {
		opt.ReportInterval = time.Duration(opt.ReportSec) * time.Second
	}
	if opt.ServerAddress != "" {
		err = config.ParseAndSetHostPort(opt.ServerAddress, opt)
		if err != nil {
			return fmt.Errorf("error parsing server address: %w", err)
		}
	}

	return nil
}

// NewOption parses the agent argv and environment and validates logging mode.
func NewOption(args []string) (*Options, error) {
	opt, err := parseArgs(args)
	if err != nil {
		return nil, labelerrors.NewLabelError("PARSE ARGS", err)
	}
	err = opt.parseEnv()
	if err != nil {
		return nil, labelerrors.NewLabelError("PARSE ENV", err)
	}
	err = config.ValidateMode(opt.Mode)
	if err != nil {
		return nil, labelerrors.NewLabelError("MODE", err)
	}
	return opt, nil
}
