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
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := initAgent()
	a.Logger.Printf("Agent initialized. server url: %v", a.ServerAddr)
	go a.StartReport(ctx)
	go a.StartPool(ctx)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.Logger.Println("Shutting down agent...")
	cancel()
}
func initAgent() *agent.Agent {
	serverCfg := server.NewConfig(server.SchemeHTTP, server.DefaultHost, server.DefaultPort, log.Default())
	agentCfg := config.NewConfig(2*time.Second, 10*time.Second, serverCfg.ServerAddr(), log.Default())
	a := agent.NewAgent(agentCfg)
	return a
}
