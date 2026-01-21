package main

import (
	"flag"
	"fmt"
	"sys-metrics/internal/config"
	"sys-metrics/internal/config/server"
	"time"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	ServerAddress  string `env:"ADDRESS"`
	Host           string
	Port           string
	PollInterval   time.Duration
	ReportInterval time.Duration
	PollSec        int `env:"POLL_INTERVAL"`
	ReportSec      int `env:"REPORT_INTERVAL"`
}

func parseArgs(args []string) (*Options, error) {
	flags := flag.NewFlagSet("agent", flag.ContinueOnError)
	opt := new(Options)
	flags.StringVar(&opt.ServerAddress, "a", fmt.Sprintf("%v:%v", server.DefaultHost, server.DefaultPort), "Address of agent server")
	flags.IntVar(&opt.ReportSec, "r", 10, "Reporting interval in seconds")
	flags.IntVar(&opt.PollSec, "p", 2, "Poll interval in seconds")
	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}

	// Устанавливаем Duration из секунд, полученных из флагов
	opt.PollInterval = time.Duration(opt.PollSec) * time.Second
	opt.ReportInterval = time.Duration(opt.ReportSec) * time.Second

	err = opt.rebuildHostAndPort()
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
		err = opt.rebuildHostAndPort()
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
func (opt *Options) rebuildHostAndPort() error {
	addr, err := config.ParseServerAddress(opt.ServerAddress)
	if err != nil {
		return err
	}
	opt.Host = addr.Host
	opt.Port = addr.Port
	return nil
}
