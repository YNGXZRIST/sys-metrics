package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sys-metrics/internal/agent"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/common"
	config "sys-metrics/internal/config/agent"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/errors/labelerrors"
	lgr "sys-metrics/internal/logger"
	"sys-metrics/internal/secure"
	"sys-metrics/internal/utils"
	"syscall"

	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if err := run(); err != nil {
		log.Printf("fatal: %v", err)
		return
	}
}
func run() error {
	utils.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opt, err := config.NewOption(os.Args[1:])
	if err != nil {
		return fmt.Errorf("new option: %w", err)
	}
	a, err := initAgent(opt, ctx)
	if err != nil {
		return labelerrors.NewLabelError("INIT AGENT", err)
	}
	defer a.Logger.Sync()
	a.Logger.Info("Agent initialized.", zap.String("server url", a.ServerAddr))
	go a.StartReport(ctx)
	go a.StartPoll(ctx)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.Logger.Info("Shutting down agent...")
	cancel()
	return nil
}
func initAgent(opt *config.Options, ctx context.Context) (*agent.Agent, error) {
	logger, err := lgr.Initialize(opt.Mode, common.TypeAgent)
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}
	serverCfg := server.NewConfig(server.SchemeHTTP, opt.Host, opt.Port, logger, nil)
	validator := authenticate.NewSha256(&opt.HashKey)
	encryptor, err := secure.NewRequestEncryptor(opt.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error initializing encryptor: %w", err)
	}
	agentCfg := config.NewConfig(opt.PollInterval, opt.ReportInterval, serverCfg.ServerAddr(), logger, validator, encryptor, opt.RateLimit)
	a := agent.NewAgent(agentCfg, ctx)
	return a, nil
}
