package server

import (
	"flag"
	"fmt"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config"
	"sys-metrics/internal/errors/labelerrors"
	"time"

	"github.com/caarlos0/env/v11"
)

type Options struct {
	ServerAddress     *string `env:"ADDRESS"`
	Host              string
	Port              string
	Mode              string `env:"MODE"`
	StoreIntervalSec  *int   `env:"STORE_INTERVAL" default:"300"`
	StoreInterval     time.Duration
	BackupStoragePath string `env:"STORE_FILE" envDefault:"./backups"`
	Restore           bool   `env:"RESTORE" envDefault:"true"`
	DNS               string `env:"DATABASE_DSN"`
}

func (opt *Options) SetHostPort(host, port string) {
	opt.Host = host
	opt.Port = port
}

func parseArgs(args []string) (*Options, error) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	opt := new(Options)
	var intervalSec int
	var serverAddr string
	flags.StringVar(&serverAddr, "a", "localhost:8080", "Address of the server")
	opt.ServerAddress = &serverAddr
	flags.StringVar(&opt.Mode, "m", common.TypeModeDefault, "Server mode. Possible values: production, development")
	flags.IntVar(&intervalSec, "i", 300, "Storage interval in seconds")
	flags.StringVar(&opt.DNS, "d", "", "Database DSN for backup storage")
	opt.StoreInterval = time.Duration(intervalSec) * time.Second
	flags.StringVar(&opt.BackupStoragePath, "f", "./backups", "Backup storage path")
	flags.BoolVar(&opt.Restore, "r", true, "Restore backups")
	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}
	err = config.ParseAndSetHostPort(serverAddr, opt)
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
	if opt.StoreIntervalSec != nil {
		opt.StoreInterval = time.Duration(*opt.StoreIntervalSec) * time.Second
	}
	if opt.ServerAddress != nil {
		err = config.ParseAndSetHostPort(*opt.ServerAddress, opt)
		if err != nil {
			return fmt.Errorf("error parsing server address: %w", err)
		}
	}

	return nil
}
func NewOption(args []string) (*Options, error) {
	opt, err := parseArgs(args)
	if err != nil {
		return nil, labelerrors.NewLabelError("ARGS", fmt.Errorf("error parsing args: %w", err))
	}
	err = opt.parseEnv()
	if err != nil {
		return nil, labelerrors.NewLabelError("ENV", fmt.Errorf("error parsing env: %w", err))
	}
	err = config.ValidateMode(opt.Mode)
	if err != nil {
		return nil, labelerrors.NewLabelError("MODE", fmt.Errorf("error validation mode: %w", err))
	}
	return opt, nil
}
