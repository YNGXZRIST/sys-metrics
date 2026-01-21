package main

import (
	"flag"
	"fmt"
	"sys-metrics/internal/config"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	ServerAddress string `env:"ADDRESS"`
	Host          string
	Port          string
}

func (opt *Options) SetHostPort(host, port string) {
	opt.Host = host
	opt.Port = port
}

func parseArgs(args []string) (*Options, error) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	opt := new(Options)
	flags.StringVar(&opt.ServerAddress, "a", "localhost:8080", "Address of the server")
	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}
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
	if opt.ServerAddress != "" {
		err = config.ParseAndSetHostPort(opt.ServerAddress, opt)
		if err != nil {
			return err
		}
	}

	return nil
}
func newOption(args []string) (*Options, error) {
	opt, err := parseArgs(args)
	if err != nil {
		return nil, err
	}
	err = opt.parseEnv()
	if err != nil {
		return nil, err
	}
	return opt, nil
}
