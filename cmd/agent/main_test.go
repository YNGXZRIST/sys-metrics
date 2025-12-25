package main

import (
	"log"
	"sys-metrics/internal/agent"
	config "sys-metrics/internal/config/agent"
	"sys-metrics/internal/config/server"
	"testing"
)

func Test_initAgent(t *testing.T) {
	serverCfg := server.NewConfig(server.SchemeHTTP, server.DefaultHost, server.DefaultPort, log.Default())
	agentCfg := config.NewConfig(2, 10, serverCfg.ServerAddr(), log.Default())
	a := agent.NewAgent(agentCfg)
	if a == nil {
		t.Fatal("init agent failed, is nil")
	}
	if a.Logger == nil {
		t.Fatal("init agent failed, logger is nil")
	}
	if a.Config == nil {
		t.Fatal("init agent failed, config is nil")
	}
}
