package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sys-metrics/internal/backup"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	lgr "sys-metrics/internal/logger"
	model "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/router"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opt, err := newOption(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	initStorage()
	err = initServer(opt, ctx)
	if err != nil {
		log.Fatal(err)
	}
}
func initStorage() {
	counters := memstorage.NewMemStorage[string, *model.Counter]()
	gauges := memstorage.NewMemStorage[string, *model.Gauge]()
	svc.Init(counters, gauges)
}
func initServer(opt *Options, ctx context.Context) error {
	logger, err := lgr.Initialize(opt.Mode, common.TypeServer)
	if err != nil {
		return err
	}
	defer logger.Sync()
	backupConfig, err := backup.NewBackupConfig(opt.Mode, opt.BackupStoragePath, opt.StoreInterval, opt.Restore)
	if err != nil {
		log.Fatal(err)
	}
	defer backupConfig.Close()
	go func() {
		if err := backupConfig.InitBackupRoutine(ctx); err != nil {
			logger.Error("backup routine error", zap.Error(err))
		}
	}()

	cfg := server.NewConfig(server.SchemeHTTP, opt.Host, opt.Port, logger, backupConfig)
	err = http.ListenAndServe(cfg.InternalAddr(), router.GetRouter(logger, cfg))
	if err != nil {
		return err
	}
	return nil
}
