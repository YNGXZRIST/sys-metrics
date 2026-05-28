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
	"time"

	"go.uber.org/zap"
)

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}
type dbCloser interface {
	Close() error
}

var _ dbCloser = (*db.DB)(nil)

type App struct {
	Server         Shutdowner
	DB             *db.DB
	Service        metricsiface.ServiceInterface
	Logger         *zap.Logger
	BackupConfig   *file.Config
	MetricsService *mServ.MetricService
	IpNet          *net.IPNet
	option         *Option
	needRestore    bool
}

type Option struct {
	AuditFilePath     string
	AuditURL          string
	Mode              string
	TrustedSubnetMask string
	DNS               string
	BackupStoragePath string
	StoreInterval     time.Duration
	Restore           bool
}

func Bootstrap(ctx context.Context, o *Option) (*App, error) {
	if o == nil {
		return nil, fmt.Errorf("bootstrap: nil option")
	}

	a := &App{option: o}

	logger, err := a.initLogger()
	if err != nil {
		return nil, err
	}
	a.Logger = logger

	if errStorage := a.initStorage(); errStorage != nil {
		return nil, errStorage
	}

	ipNet, errNet := a.parseTrustedSubnet()
	if errNet != nil {
		return nil, errNet
	}
	a.IpNet = ipNet

	metrics.Init(a.Service)

	if a.needRestore {
		if err = a.restoreFromBackup(ctx); err != nil {
			return nil, err
		}
	}

	startBackupRoutine(ctx, a.Service, logger)

	metricsObserver, err := a.initMetricsObserver(ctx)
	if err != nil {
		return nil, err
	}
	a.MetricsService = mServ.NewService(metricsObserver)

	return a, nil
}

func (a *App) Close(ctx context.Context) error {
	var errs []error
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown server: %w", err))
		}
	}

	if a.Service != nil {
		if err := a.Service.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("close service: %w", err))
		}
	}

	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close database: %w", err))
		}
	}

	if a.BackupConfig != nil {
		if err := a.BackupConfig.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close backup config: %w", err))
		}
	}

	if a.Logger != nil {
		if err := a.Logger.Sync(); err != nil {
			errs = append(errs, fmt.Errorf("sync logger: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (a *App) initMetricsObserver(ctx context.Context) (*observer.MetricsObserver, error) {
	cfg := observer.MetricObserverConfig{
		FilePath:  a.option.AuditFilePath,
		URL:       a.option.AuditURL,
		Mode:      a.option.Mode,
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
	if a.option.TrustedSubnetMask == "" {
		return nil, nil
	}
	_, ipNet, err := net.ParseCIDR(a.option.TrustedSubnetMask)
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
	if a.Service == nil {
		return nil
	}
	if err := a.Service.ReadBackup(ctx); err != nil {
		return labelerrors.NewLabelError(
			"INIT BACKUP",
			fmt.Errorf("error reading backup: %w", err),
		)
	}

	return nil
}

func (a *App) initStorage() error {
	a.Service = nil
	a.DB = nil
	a.BackupConfig = nil
	a.needRestore = false

	if a.isDSNSet() {
		if err := migrations.Migrate(a.option.DNS); err != nil {
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
			a.Service = postgres.NewMetricStorage(conn)
			a.DB = conn
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

		a.Service = file.NewBackupService(backupStorage)
		a.BackupConfig = backupConfig
		a.needRestore = true
		return nil
	}

	a.Service = memory.NewService()
	return nil
}

func (a *App) initLogger() (*zap.Logger, error) {
	logger, err := lgr.Initialize(a.option.Mode, common.TypeServer)
	if err != nil {
		return nil, labelerrors.NewLabelError("LOGGER", fmt.Errorf("error initializing logger: %w", err))
	}

	return logger, nil
}

func (a *App) isDSNSet() bool {
	return a.option.DNS != ""
}

func (a *App) initDB() (*db.DB, error) {
	cfg := db.NewCfg(&db.Config{DNS: a.option.DNS})

	return db.NewConn(cfg)
}

func (a *App) initBackupConfig() (*file.Config, error) {
	return file.NewConfig(
		a.option.Mode,
		a.option.BackupStoragePath,
		a.option.StoreInterval,
		a.option.Restore,
	)
}

func isDatabaseConnected(conn *db.DB) bool {
	return conn != nil && conn.Ping() == nil
}
func OptionFromServer(o *server.Options) *Option {
	return &Option{
		AuditFilePath:     o.AuditFile,
		AuditURL:          o.AuditURL,
		Mode:              o.Mode,
		TrustedSubnetMask: o.TrustedSubnetMask,
		DNS:               o.DNS,
		BackupStoragePath: o.BackupStoragePath,
		StoreInterval:     o.StoreInterval,
		Restore:           o.Restore,
	}
}
