package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/handler"
	lgr "sys-metrics/internal/logger"
	"sys-metrics/internal/observer"
	"sys-metrics/internal/repository/file"
	"sys-metrics/internal/repository/memory"
	"sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/internal/repository/postgres"
	"sys-metrics/internal/router"
	"sys-metrics/internal/secure"
	"sys-metrics/internal/utils"
	"sys-metrics/migrations"

	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	utils.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	err := run(os.Args[1:])
	if err != nil {
		fmt.Printf("fatal error: %v\n", err)
	}
}
func run(args []string) error {
	o, err := server.NewOption(args)
	if err != nil {
		return labelerrors.NewLabelError("PARSE OPTIONS", fmt.Errorf("error parsing flags: %w", err))
	}
	ctx := context.Background()
	err = initServer(ctx, o)
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

func isDSNSet(o *server.Options) bool {
	return o.DNS != ""
}

func initDB(o *server.Options) (*db.DB, error) {
	cfg := db.NewCfg(o)
	return db.NewConn(cfg)
}

func initBackupConfig(o *server.Options) (*file.Config, error) {
	fmt.Printf("%+v\n", o)
	return file.NewConfig(o.Mode, o.BackupStoragePath, o.StoreInterval, o.Restore)
}

func isDatabaseConnected(conn *db.DB) bool {
	return conn != nil && conn.Ping() == nil
}

func createService(o *server.Options) (metricsiface.ServiceInterface, *db.DB, *file.Config, bool, error) {
	if isDSNSet(o) {
		err := migrations.Migrate(o.DNS)
		if err != nil {
			return nil, nil, nil, false, labelerrors.NewLabelError("MIGRATE", fmt.Errorf("error initializing database connection: %w", err))
		}
		conn, err := initDB(o)
		if err != nil {
			return nil, nil, nil, false, labelerrors.NewLabelError("INIT DB", fmt.Errorf("error initializing database connection: %w", err))
		}
		if isDatabaseConnected(conn) {
			service := postgres.NewMetricStorage(conn)
			return service, conn, nil, true, nil
		}
	}

	backupConfig, err := initBackupConfig(o)
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

func startHTTPServer(o *server.Options, h *handler.Handler, backupConfig *file.Config) error {
	cfg := server.NewConfig(server.SchemeHTTP, o.Host, o.Port, h.Logger, backupConfig)
	if err := http.ListenAndServe(cfg.InternalAddr(), router.GetRouter(h)); err != nil {
		return labelerrors.NewLabelError("HTTP", fmt.Errorf("error starting HTTP server: %w", err))
	}
	return nil
}

func initServer(ctx context.Context, o *server.Options) error {
	logger, err := initLogger(o.Mode)
	if err != nil {
		return err
	}
	defer logger.Sync()

	serviceInterface, conn, backupConfigForHTTP, needRestore, err := createService(o)
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
	metricsObserver, err := initMetricsObserver(ctx, o)
	if err != nil {
		return err
	}
	observersMap := make(map[handler.ObserverKey]observer.Observer)
	observersMap[handler.ObserverAudit] = metricsObserver
	authenticator := initAuthenticator(o)
	reqDecryptor, err := secure.NewRequestDecryptor(o.CryptoKeyPath)
	if err != nil {
		return fmt.Errorf("error initializing request decryptor: %w", err)
	}
	h := initHandler(conn, logger, authenticator, reqDecryptor, observersMap)
	return startHTTPServer(o, h, backupConfigForHTTP)
}
func initMetricsObserver(ctx context.Context, o *server.Options) (*observer.MetricsObserver, error) {
	cfg := observer.MetricObserverConfig{
		FilePath:  o.AuditFile,
		URL:       o.AuditURL,
		Mode:      o.Mode,
		RateLimit: 1,
	}
	obs, err := observer.NewMetricsObserver(cfg)
	if err != nil {
		return nil, err
	}
	err = obs.Register(ctx)
	if err != nil {
		return nil, err
	}
	return obs, nil
}
func initHandler(c *db.DB, l *zap.Logger, a authenticate.Authenticator, d *secure.RequestDecryptor, o map[handler.ObserverKey]observer.Observer) *handler.Handler {
	newHandler := handler.NewHandler(c, a, d, l, o)
	return newHandler
}
func initAuthenticator(o *server.Options) authenticate.Authenticator {
	sha := authenticate.NewSha256(o.HashKey)
	var a authenticate.Authenticator
	if sha != nil {
		a = sha
	}
	return a
}
