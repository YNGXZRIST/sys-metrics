package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	agent "sys-metrics/internal/agent"
	config "sys-metrics/internal/config/agent"
	"sys-metrics/internal/config/server"
	"syscall"
)

func main() {

	serverCfg := server.NewConfig(server.SchemeHTTP, server.DefaultHost, server.DefaultPort, log.Default())
	agentCfg := config.NewConfig(2, 10, serverCfg.ServerAddr(), log.Default())
	fmt.Println(agentCfg.ServerAddr)
	a := agent.NewAgent(agentCfg)
	go a.StartReport()
	go a.StartPool()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down agent...")

}
