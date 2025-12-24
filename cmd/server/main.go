package main

import (
	"log"
	"net/http"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/handler/router"
	model "sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
)

func main() {
	counters := memstorage.NewMemStorage[string, *model.Counter]()
	gauges := memstorage.NewMemStorage[string, *model.Gauge]()
	svc.Init(counters, gauges)
	cfg := server.NewConfig(server.SchemeHTTP, server.DefaultHost, server.DefaultPort, log.Default())
	err := http.ListenAndServe(cfg.InternalAddr(), router.GetRouter())
	if err != nil {
		panic(err)
	}
}
