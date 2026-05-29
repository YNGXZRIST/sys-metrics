package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sys-metrics/internal/agent"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/common"
	config "sys-metrics/internal/config/agent"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/errors/labelerrors"
	lgr "sys-metrics/internal/logger"
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

func main() {
	if err := run(); err != nil {
		log.Printf("fatal: %v", err)
		return
	}
}
func run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()
	return runWithContext(ctx, os.Args[1:])
}

func runWithContext(ctx context.Context, args []string) error {
	utils.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	opt, err := config.NewOption(args)
	if err != nil {
		return fmt.Errorf("new option: %w", err)
	}
	a, err := initAgent(opt, ctx)
	if err != nil {
		return labelerrors.NewLabelError("INIT AGENT", err)
	}
	defer a.Logger.Sync()
	a.Logger.Info("Agent initialized.",
		zap.String("server url", a.ServerAddr),
		zap.String("endpoint", opt.Endpoint()),
		zap.String("report transport", opt.ReportTransport),
	)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		a.StartReport(ctx)
	}()
	go func() {
		defer wg.Done()
		a.StartPoll(ctx)
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.Close(shutdownCtx); err != nil {
		a.Logger.Error("shutdown error", zap.Error(err))
	}
	wg.Wait()
	a.Logger.Info("Shutting down agent...")
	return nil
}
func initAgent(opt *config.Options, ctx context.Context) (*agent.Agent, error) {
	logger, err := lgr.Initialize(opt.Mode, common.TypeAgent)
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}
	serverCfg := server.NewConfig(server.SchemeHTTP, opt.Host, opt.Port, logger, nil)
	validator := authenticate.NewSha256(&opt.HashKey)
	encryptor, err := secure.NewRequestEncryptor(opt.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error initializing encryptor: %w", err)
	}
	localIpV4, err := getAgentLocalIpV4()
	if err != nil {
		return nil, fmt.Errorf("error initializing local address: %w", err)
	}
	initProp := config.InitProperties{
		PollInterval:     opt.PollInterval,
		ReportInterval:   opt.ReportInterval,
		ServerAddr:       serverCfg.ServerAddr(),
		ServerEndpoint:   serverCfg.InternalAddr(),
		ReportTransport:  opt.ReportTransport,
		Logger:           logger,
		Authenticator:    validator,
		RequestEncryptor: encryptor,
		RateLimit:        opt.RateLimit,
		LocalIpV4:        localIpV4,
	}
	agentCfg := config.NewConfig(initProp)
	a, err := agent.NewAgent(agentCfg, ctx)
	return a, nil
}
func getAgentLocalIpV4() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", labelerrors.NewLabelError("INIT AGENT", fmt.Errorf("no local addresses found"))
}
