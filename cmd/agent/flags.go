package main

import (
	"flag"
	"fmt"
	"sys-metrics/internal/config"
	"sys-metrics/internal/config/server"
	"time"
)

type Options struct {
	serverAddress  string
	host           string
	port           string
	poolInterval   time.Duration
	reportInterval time.Duration
}

func parseArgs(args []string) (*Options, error) {
	flags := flag.NewFlagSet("agent", flag.ContinueOnError)
	opt := new(Options)
	var poolSec int
	var reportSec int
	flags.StringVar(&opt.serverAddress, "a", fmt.Sprintf("%v:%v", server.DefaultHost, server.DefaultPort), "Address of agent server")
	flags.IntVar(&reportSec, "r", 10, "Reporting interval in seconds")
	flags.IntVar(&poolSec, "p", 2, "Pool interval in seconds")
	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}

	opt.poolInterval = time.Duration(poolSec) * time.Second
	opt.reportInterval = time.Duration(reportSec) * time.Second
	addr, err := config.ParseServerAddress(opt.serverAddress)
	if err != nil {
		return nil, err
	}
	opt.host = addr.Host
	opt.port = addr.Port
	return opt, nil
}
