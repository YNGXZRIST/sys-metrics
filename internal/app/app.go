// Package app wires shared server startup and shutdown: storage, DB, backup, audit, and transport.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/errors/labelerrors"
	lgr "sys-metrics/internal/logger"
	"sys-metrics/internal/observer"
	"sys-metrics/internal/repository/file"
	"sys-metrics/internal/repository/memory"
	"sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/internal/repository/postgres"
	mServ "sys-metrics/internal/service/metrics"
	"sys-metrics/migrations"

	"go.uber.org/zap"
)

// Shutdowner stops the network transport gracefully using ctx as a deadline.
type shutdowner interface {
	shutdown(ctx context.Context) error
}

type dbCloser interface {
	Close() error
}

var _ dbCloser = (*db.DB)(nil)

// App holds runtime dependencies shared by HTTP and gRPC servers.
type App struct {
	serverHTTP
	serverGRPC     shutdownGRPCServer
	db             *db.DB
	service        metricsiface.ServiceInterface
	logger         *zap.Logger
	backupConfig   *file.Config
	metricsService *mServ.MetricService
	ipNet          *net.IPNet
	opts           *server.Options
	needRestore    bool
}

// Logger returns the application logger (for tests and diagnostics).
func (a *App) Logger() *zap.Logger {
	return a.logger
}

// Bootstrap initializes logger, metrics storage, optional restore, backup routine, and MetricService.
func Bootstrap(ctx context.Context, opts *server.Options) (*App, error) {
	if opts == nil {
		return nil, fmt.Errorf("bootstrap: nil options")
	}

	a := &App{opts: opts}

	logger, err := a.initLogger()
	if err != nil {
		return nil, err
	}
	a.logger = logger

	if errStorage := a.initStorage(); errStorage != nil {
		return nil, errStorage
	}

	ipNet, errNet := a.parseTrustedSubnet()
	if errNet != nil {
		return nil, errNet
	}
	a.ipNet = ipNet

	metrics.Init(a.service)

	if a.needRestore {
		if err := a.restoreFromBackup(ctx); err != nil {
			return nil, err
		}
	}

	startBackupRoutine(ctx, a.service, logger)

	metricsObserver, err := a.initMetricsObserver(ctx)
	if err != nil {
		return nil, err
	}
	a.metricsService = mServ.NewService(metricsObserver)

	return a, nil
}

// StartServers starts HTTP and gRPC listeners.
func (a *App) StartServers() error {
	if err := a.startHTTP(); err != nil {
		return err
	}
	if err := a.startGRPC(); err != nil {
		return err
	}

	a.logger.Info("servers started",
		zap.String(common.ReportTransportHTTP, net.JoinHostPort(a.opts.Host, a.opts.Port)),
		zap.String(common.ReportTransportGRPC, a.opts.GRPCInternalAddr()),
	)

	return nil
}

// Close shuts down transport, storage, DB, backup config, and syncs the logger.
func (a *App) Close(ctx context.Context) error {
	var errs []error
	if err := a.serverHTTP.shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("shutdown http server: %w", err))
	}
	if err := a.serverGRPC.shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("shutdown grpc server: %w", err))
	}

	if a.service != nil {
		if err := a.service.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("close service: %w", err))
		}
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close database: %w", err))
		}
	}

	if a.backupConfig != nil {
		if err := a.backupConfig.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close backup config: %w", err))
		}
	}

	if a.logger != nil {
		if err := a.logger.Sync(); err != nil {
			errs = append(errs, fmt.Errorf("sync logger: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (a *App) initMetricsObserver(ctx context.Context) (*observer.MetricsObserver, error) {
	cfg := observer.MetricObserverConfig{
		FilePath:  a.opts.AuditFile,
		URL:       a.opts.AuditURL,
		Mode:      a.opts.Mode,
		RateLimit: 1,
	}

	obs, err := observer.NewMetricsObserver(cfg)
	if err != nil {
		return nil, err
	}

	if err = obs.Register(ctx); err != nil {
		return nil, err
	}

	return obs, nil
}

func (a *App) parseTrustedSubnet() (*net.IPNet, error) {
	if a.opts.TrustedSubnetMask == "" {
		return nil, nil
	}
	_, ipNet, err := net.ParseCIDR(a.opts.TrustedSubnetMask)
	return ipNet, err
}

func startBackupRoutine(
	ctx context.Context,
	service metricsiface.ServiceInterface,
	logger *zap.Logger,
) {
	go func() {
		if err := service.InitRoutine(ctx); err != nil {
			logger.Error(
				"backup routine error",
				zap.Error(err),
			)
		}
	}()
}

func (a *App) restoreFromBackup(ctx context.Context) error {
	if a.service == nil {
		return nil
	}
	if err := a.service.ReadBackup(ctx); err != nil {
		return labelerrors.NewLabelError(
			"INIT BACKUP",
			fmt.Errorf("error reading backup: %w", err),
		)
	}

	return nil
}

func (a *App) initStorage() error {
	a.service = nil
	a.db = nil
	a.backupConfig = nil
	a.needRestore = false

	if a.isDSNSet() {
		if err := migrations.Migrate(a.opts.DNS); err != nil {
			return labelerrors.NewLabelError(
				"MIGRATE",
				fmt.Errorf("error initializing database connection: %w", err),
			)
		}

		conn, err := a.initDB()
		if err != nil {
			return labelerrors.NewLabelError(
				"INIT DB",
				fmt.Errorf("error initializing database connection: %w", err),
			)
		}

		if isDatabaseConnected(conn) {
			a.service = postgres.NewMetricStorage(conn)
			a.db = conn
			a.needRestore = true
			return nil
		}
	}

	backupConfig, err := a.initBackupConfig()
	if err != nil {
		return labelerrors.NewLabelError(
			"INIT BACKUP",
			fmt.Errorf("error initializing backup config: %w", err),
		)
	}

	if backupConfig.Enabled {
		backupStorage, err := file.NewMetricFileBackupStorage(backupConfig)
		if err != nil {
			_ = backupConfig.Close()
			return labelerrors.NewLabelError(
				"INIT BACKUP",
				fmt.Errorf("error initializing backup storage: %w", err),
			)
		}

		a.service = file.NewBackupService(backupStorage)
		a.backupConfig = backupConfig
		a.needRestore = true
		return nil
	}

	a.service = memory.NewService()
	return nil
}

func (a *App) initLogger() (*zap.Logger, error) {
	logger, err := lgr.Initialize(a.opts.Mode, common.TypeServer)
	if err != nil {
		return nil, labelerrors.NewLabelError("LOGGER", fmt.Errorf("error initializing logger: %w", err))
	}

	return logger, nil
}

func (a *App) isDSNSet() bool {
	return a.opts.DNS != ""
}

func (a *App) initDB() (*db.DB, error) {
	cfg := db.NewCfg(&db.Config{DNS: a.opts.DNS})

	return db.NewConn(cfg)
}

func (a *App) initBackupConfig() (*file.Config, error) {
	return file.NewConfig(
		a.opts.Mode,
		a.opts.BackupStoragePath,
		a.opts.StoreInterval,
		a.opts.Restore,
	)
}

func isDatabaseConnected(conn *db.DB) bool {
	return conn != nil && conn.Ping() == nil
}
