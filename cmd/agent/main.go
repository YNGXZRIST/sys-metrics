package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sys-metrics/internal/agent"
	config "sys-metrics/internal/config/agent"
	"sys-metrics/internal/config/server"
	"syscall"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opt, err := parseArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	a := initAgent(opt)
	a.Logger.Printf("Agent initialized. server url: %v", a.ServerAddr)
	go a.StartReport(ctx)
	go a.StartPool(ctx)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.Logger.Println("Shutting down agent...")
	cancel()
}
func initAgent(opt *Options) *agent.Agent {
	serverCfg := server.NewConfig(server.SchemeHTTP, opt.host, opt.port, log.Default())
	agentCfg := config.NewConfig(opt.poolInterval, opt.reportInterval, serverCfg.ServerAddr(), log.Default())
	a := agent.NewAgent(agentCfg)
	return a
}
