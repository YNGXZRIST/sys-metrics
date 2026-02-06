package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	lgr "sys-metrics/internal/logger"
	model "sys-metrics/internal/model/metrics"
	repo "sys-metrics/internal/repository"
	"sys-metrics/internal/router"
	"sys-metrics/pkg/storage"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opt, err := server.NewOption(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	err = initServer(opt, ctx)
	if err != nil {
		log.Fatal(err)
	}
}
func initStorage(backupConfig *repo.Config) {
	if backupConfig.Enabled {
		backupStorage, err := repo.NewMetricBackupStorage(backupConfig)
		if err != nil {
			log.Fatal(err)
		}
		service := repo.InitBackup(backupStorage)

		if err := service.ReadBackup(); err != nil {
			log.Printf("warning: failed to restore from backup: %v", err)
		}
	} else {
		counters := storage.NewMemStorage[string, *model.Counter]()
		gauges := storage.NewMemStorage[string, *model.Gauge]()
		repo.Init(counters, gauges)
	}
}
func initServer(opt *server.Options, ctx context.Context) error {
	logger, err := lgr.Initialize(opt.Mode, common.TypeServer)
	if err != nil {
		return fmt.Errorf("error initializing logger: %w", err)
	}
	defer logger.Sync()
	backupConfig, err := repo.NewConfig(opt.Mode, opt.BackupStoragePath, opt.StoreInterval, opt.Restore)
	if err != nil {
		log.Fatal(err)
	}
	defer backupConfig.Close()

	initStorage(backupConfig)

	go func() {
		if err := backupConfig.InitBackupRoutine(ctx); err != nil {
			logger.Error("backup routine error", zap.Error(err))
		}
	}()

	cfg := server.NewConfig(server.SchemeHTTP, opt.Host, opt.Port, logger, backupConfig)
	err = http.ListenAndServe(cfg.InternalAddr(), router.GetRouter(logger, cfg))
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
