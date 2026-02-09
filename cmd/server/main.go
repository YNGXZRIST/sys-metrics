package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/config/server"
	lgr "sys-metrics/internal/logger"
	model "sys-metrics/internal/model/metrics"
	repo "sys-metrics/internal/repository"
	"sys-metrics/internal/router"
	"sys-metrics/pkg/storage"

	"go.uber.org/zap"
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
func run(args []string) error {
	opt, err := server.NewOption(args)
	if err != nil {
		return fmt.Errorf("error parsing options: %w", err)
	}
	ctx := context.Background()
	err = initServer(ctx, opt)
	if err != nil {
		return fmt.Errorf("error initializing server: %w", err)
	}
	return nil
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
func initServer(ctx context.Context, opt *server.Options) error {
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

	routineCtx, cancel := context.WithCancel(ctx)
	go func() {
		if err := backupConfig.InitBackupRoutine(routineCtx); err != nil {
			logger.Error("backup routine error", zap.Error(err))
		}
	}()
	defer cancel()

	conn, err := initDB(opt)
	if err != nil {
		return err
	}
	defer conn.Close()

	cfg := server.NewConfig(server.SchemeHTTP, opt.Host, opt.Port, logger, backupConfig)
	if err := http.ListenAndServe(cfg.InternalAddr(), router.GetRouter(logger, conn)); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func initDB(opt *server.Options) (*db.DB, error) {
	dbConfig := db.NewCfg(opt)
	return db.NewConn(dbConfig)
}
