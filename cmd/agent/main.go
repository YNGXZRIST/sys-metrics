package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sys-metrics/internal/agent"
	"sys-metrics/internal/common"
	config "sys-metrics/internal/config/agent"
	"sys-metrics/internal/config/server"
	lgr "sys-metrics/internal/logger"
	"syscall"

	"go.uber.org/zap"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opt, err := newOption(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	a, err := initAgent(opt)
	if err != nil {
		log.Fatal(err)
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
}
func initAgent(opt *Options) (*agent.Agent, error) {
	logger, err := lgr.Initialize(opt.Mode, common.TypeAgent)
	if err != nil {
		return nil, err
	}
	serverCfg := server.NewConfig(server.SchemeHTTP, opt.Host, opt.Port, logger, nil)
	agentCfg := config.NewConfig(opt.PollInterval, opt.ReportInterval, serverCfg.ServerAddr(), logger)
	a := agent.NewAgent(agentCfg)
	return a, nil
}
