package main

import (
	"sys-metrics/internal/handler/router"
	"sys-metrics/pkg/server"
)

func main() {
	e := server.NewServer(server.DefaultHost, server.DefaultPort, router.SetRouter())
	if e != nil {
		panic(e)
	}
}
