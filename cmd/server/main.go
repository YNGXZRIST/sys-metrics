package main

import (
	"sys-metrics/internal/handler/router"
	model "sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
	"sys-metrics/pkg/server"
)

func main() {
	counters := memstorage.NewMemStorage[string, *model.Counter]()
	gauges := memstorage.NewMemStorage[string, *model.Gauge]()
	svc.Init(counters, gauges)
	e := server.NewServer(server.DefaultHost, server.DefaultPort, router.GetRouter())
	if e != nil {
		panic(e)
	}
}
