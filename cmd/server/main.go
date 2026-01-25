package main

import (
	"log"
	"net/http"
	"os"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/handler/router"
	lgr "sys-metrics/internal/logger"
	"sys-metrics/internal/middleware"
	model "sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
)

func main() {
	opt, err := newOption(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	initStorage()
	err = initServer(opt)
	if err != nil {
		log.Fatal(err)
	}
}
func initStorage() {
	counters := memstorage.NewMemStorage[string, *model.Counter]()
	gauges := memstorage.NewMemStorage[string, *model.Gauge]()
	svc.Init(counters, gauges)
}
func initServer(opt *Options) error {
	logger, initialize := lgr.Initialize(opt.Mode)
	if initialize != nil {
		return initialize
	}
	cfg := server.NewConfig(server.SchemeHTTP, opt.Host, opt.Port, logger)
	err := http.ListenAndServe(cfg.InternalAddr(), middleware.WithLogging(logger)(router.GetRouter()))
	if err != nil {
		return err
	}
	return nil
}
