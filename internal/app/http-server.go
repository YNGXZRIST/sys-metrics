package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/handler"
	"sys-metrics/internal/router"
	"sys-metrics/internal/secure"

	"go.uber.org/zap"
)

type serverHTTP struct {
	server *http.Server
}

func (s *serverHTTP) shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *serverHTTP) listenAndServe(logger *zap.Logger) {
	go func() {
		if err := s.server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", zap.Error(err))
		}
	}()
}

// startHTTP builds the handler, starts the HTTP server, and stores it on App for shutdown.
func (a *App) startHTTP() error {
	h, err := a.newHTTPHandler()
	if err != nil {
		return fmt.Errorf("init http handler: %w", err)
	}

	cfg := server.NewConfig(
		server.SchemeHTTP,
		a.opts.Host,
		a.opts.Port,
		h.Logger,
		a.backupConfig,
	)

	a.serverHTTP.server = &http.Server{
		Addr:    cfg.InternalAddr(),
		Handler: router.GetRouter(h),
	}
	a.serverHTTP.listenAndServe(a.logger)

	return nil
}

func (a *App) newHTTPHandler() (*handler.Handler, error) {
	reqDecryptor, err := secure.NewRequestDecryptor(a.opts.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("request decryptor: %w", err)
	}

	return handler.NewHandler(
		handler.InitProperties{
			Logger:           a.logger,
			Conn:             a.db,
			Authenticator:    initAuthenticator(a.opts),
			RequestDecryptor: reqDecryptor,
			MetricService:    a.metricsService,
			IpNet:            a.ipNet,
		},
	), nil
}

func initAuthenticator(o *server.Options) authenticate.Authenticator {
	sha := authenticate.NewSha256(o.HashKey)
	if sha != nil {
		return sha
	}
	return nil
}
