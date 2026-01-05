package main

import (
	"log"
	"net/http"
	"os"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/handler/router"
	model "sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
)

func main() {
	opt, err := parseArgs(os.Args[1:])
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
	cfg := server.NewConfig(server.SchemeHTTP, opt.host, opt.port, log.Default())
	err := http.ListenAndServe(cfg.InternalAddr(), router.GetRouter())
	if err != nil {
		return err
	}
	return nil
}
