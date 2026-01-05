package main

import (
	"flag"
	"sys-metrics/internal/config"
)

type Options struct {
	serverAddress string
	host          string
	port          string
}

func parseArgs(args []string) (*Options, error) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	opt := new(Options)
	flags.StringVar(&opt.serverAddress, "a", "localhost:8080", "Address of the server")
	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}
	addr, err2 := config.ParseServerAddress(opt.serverAddress)
	if err2 != nil {
		return nil, err2
	}
	opt.host = addr.Host
	opt.port = addr.Port
	return opt, nil
}
