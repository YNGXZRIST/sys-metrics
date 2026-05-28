package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sys-metrics/internal/app"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/handler"
	"sys-metrics/internal/router"
	"sys-metrics/internal/secure"
	"sys-metrics/internal/utils"
	"syscall"
	"time"

	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

type AppHTTP struct {
	o   *server.Options
	h   *handler.Handler
	App *app.App
}

func main() {
	utils.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	a, err := run(ctx, os.Args[1:])
	if err != nil {
		log.Fatalf("fatal error: %v", err)
	}

	<-ctx.Done()

	log.Println("shutting down application...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if errShutdown := a.App.Close(ctxShutdown); errShutdown != nil {
		log.Printf("shutdown error: %v", errShutdown)
	}

	log.Println("application stopped")
}

func run(ctx context.Context, args []string) (*AppHTTP, error) {
	o, err := server.NewOption(common.ServerGRPC, args)
	if err != nil {
		return nil, labelerrors.NewLabelError("PARSE OPTIONS", fmt.Errorf("error parsing flags: %w", err))
	}
	a := &AppHTTP{
		o: o,
	}
	baseApp, err := app.Bootstrap(ctx, app.OptionFromServer(a.o))
	if err != nil {
		return nil, labelerrors.NewLabelError("BOOTSTRAP", fmt.Errorf("error bootstrapping: %w", err))
	}
	a.App = baseApp

	err = a.initServer()
	if err != nil {
		return nil, labelerrors.NewLabelError("INIT SERVER", fmt.Errorf("error initializing server: %w", err))
	}

	return a, nil
}

func (a *AppHTTP) initServer() error {

	err := a.initHandler()
	if err != nil {
		return labelerrors.NewLabelError("HANDLER", fmt.Errorf("error initializing handler: %w", err))
	}
	return a.startHTTPServer()
}
func (a *AppHTTP) initHandler() error {
	authenticator := initAuthenticator(a.o)

	reqDecryptor, err := secure.NewRequestDecryptor(a.o.CryptoKeyPath)
	if err != nil {
		return fmt.Errorf("error initializing request decryptor: %w", err)
	}
	h := handler.NewHandler(
		handler.InitProperties{
			Logger:           a.App.Logger,
			Conn:             a.App.DB,
			Authenticator:    authenticator,
			RequestDecryptor: reqDecryptor,
			MetricService:    a.App.MetricsService,
			IpNet:            a.App.IpNet,
		},
	)
	a.h = h
	return nil
}

func (a *AppHTTP) startHTTPServer() error {
	cfg := server.NewConfig(
		server.SchemeHTTP,
		a.o.Host,
		a.o.Port,
		a.h.Logger,
		a.App.BackupConfig,
	)

	srv := &http.Server{
		Addr:    cfg.InternalAddr(),
		Handler: router.GetRouter(a.h),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			a.h.Logger.Error("http server error", zap.Error(err))
		}
	}()
	a.App.Server = srv
	return nil
}

func initAuthenticator(o *server.Options) authenticate.Authenticator {
	sha := authenticate.NewSha256(o.HashKey)
	var a authenticate.Authenticator
	if sha != nil {
		a = sha
	}
	return a
}
