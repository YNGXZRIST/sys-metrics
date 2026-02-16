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
	"sys-metrics/internal/errors/labelerrors"
	lgr "sys-metrics/internal/logger"
	"sys-metrics/internal/repository/file"
	"sys-metrics/internal/repository/memory"
	"sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/internal/repository/postgres"
	"sys-metrics/internal/router"
	"sys-metrics/migrations"

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
		return labelerrors.NewLabelError("PARSE OPTIONS", fmt.Errorf("error parsing flags: %w", err))
	}
	ctx := context.Background()
	err = initServer(ctx, opt)
	if err != nil {
		return labelerrors.NewLabelError("INIT SERVER", fmt.Errorf("error initializing server: %w", err))
	}
	return nil
}
func initLogger(mode string) (*zap.Logger, error) {
	logger, err := lgr.Initialize(mode, common.TypeServer)
	if err != nil {
		return nil, labelerrors.NewLabelError("LOGGER", fmt.Errorf("error initializing logger: %w", err))
	}
	return logger, nil
}

func isDSNSet(opt *server.Options) bool {
	return opt.DNS != ""
}

func initDB(opt *server.Options) (*db.DB, error) {
	cfg := db.NewCfg(opt)
	return db.NewConn(cfg)
}

func initBackupConfig(opt *server.Options) (*file.Config, error) {
	return file.NewConfig(opt.Mode, opt.BackupStoragePath, opt.StoreInterval, opt.Restore)
}

func isDatabaseConnected(conn *db.DB) bool {
	return conn != nil && conn.Ping() == nil
}

func createService(opt *server.Options) (metricsiface.ServiceInterface, *db.DB, *file.Config, bool, error) {
	if isDSNSet(opt) {
		err := migrations.Migrate(opt.DNS)
		if err != nil {
			return nil, nil, nil, false, labelerrors.NewLabelError("MIGRATE", fmt.Errorf("error initializing database connection: %w", err))
		}
		conn, err := initDB(opt)
		if err != nil {
			return nil, nil, nil, false, labelerrors.NewLabelError("INIT DB", fmt.Errorf("error initializing database connection: %w", err))
		}
		if isDatabaseConnected(conn) {
			service := postgres.NewMetricStorage(conn)
			return service, conn, nil, true, nil
		}
	}

	backupConfig, err := initBackupConfig(opt)
	if err != nil {
		return nil, nil, nil, false, labelerrors.NewLabelError("INIT BACKUP", fmt.Errorf("error initializing backup config: %w", err))
	}
	if backupConfig.Enabled {
		backupStorage, err := file.NewMetricFileBackupStorage(backupConfig)
		if err != nil {
			_ = backupConfig.Close()
			return nil, nil, nil, false, labelerrors.NewLabelError("INIT BACKUP", fmt.Errorf("error initializing backup storage: %w", err))
		}
		service := file.NewBackupService(backupStorage)
		return service, nil, backupConfig, true, nil
	}
	return memory.NewService(), nil, nil, false, nil
}

func restoreFromBackup(ctx context.Context, service metricsiface.ServiceInterface) error {
	if err := service.ReadBackup(ctx); err != nil {
		return labelerrors.NewLabelError("INIT BACKUP", fmt.Errorf("error reading backup: %w", err))
	}
	return nil
}

func startBackupRoutine(ctx context.Context, service metricsiface.ServiceInterface, logger *zap.Logger) context.CancelFunc {
	routineCtx, cancel := context.WithCancel(ctx)
	go func() {
		if err := service.InitRoutine(routineCtx); err != nil {
			logger.Error("backup routine error", zap.Error(err))
		}
	}()
	return cancel
}

func startHTTPServer(opt *server.Options, logger *zap.Logger, backupConfig *file.Config, conn *db.DB) error {
	cfg := server.NewConfig(server.SchemeHTTP, opt.Host, opt.Port, logger, backupConfig)
	if err := http.ListenAndServe(cfg.InternalAddr(), router.GetRouter(logger, conn)); err != nil {
		return labelerrors.NewLabelError("HTTP", fmt.Errorf("error starting HTTP server: %w", err))
	}
	return nil
}

func initServer(ctx context.Context, opt *server.Options) error {
	logger, err := initLogger(opt.Mode)
	if err != nil {
		return err
	}
	defer logger.Sync()

	serviceInterface, conn, backupConfigForHTTP, needRestore, err := createService(opt)
	if err != nil {
		return err
	}
	defer serviceInterface.Close(ctx)
	if conn != nil {
		defer conn.Close()
	}
	defer backupConfigForHTTP.Close()

	defer backupConfigForHTTP.Close()
	metrics.Init(serviceInterface)
	if needRestore {
		err = restoreFromBackup(ctx, serviceInterface)
		if err != nil {
			return err
		}
	}

	cancel := startBackupRoutine(ctx, serviceInterface, logger)
	defer cancel()

	return startHTTPServer(opt, logger, backupConfigForHTTP, conn)
}
