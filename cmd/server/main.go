package main

import (
	"context"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sys-metrics/internal/app"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/utils"
	"syscall"
	"time"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

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

	if errShutdown := a.Close(ctxShutdown); errShutdown != nil {
		log.Printf("shutdown error: %v", errShutdown)
	}

	log.Println("application stopped")
}

func run(ctx context.Context, args []string) (*app.App, error) {
	o, err := server.NewOption(args)
	if err != nil {
		return nil, labelerrors.NewLabelError("PARSE OPTIONS", fmt.Errorf("error parsing flags: %w", err))
	}

	a, errInit := app.Bootstrap(ctx, o)
	if errInit != nil {
		return nil, labelerrors.NewLabelError("BOOTSTRAP", fmt.Errorf("error bootstrapping: %w", errInit))
	}

	if errServe := a.StartServers(); errServe != nil {
		return nil, labelerrors.NewLabelError("INIT SERVERS", fmt.Errorf("error starting servers: %w", errServe))
	}

	return a, nil
}
