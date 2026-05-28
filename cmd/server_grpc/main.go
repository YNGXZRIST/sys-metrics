package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sys-metrics/internal/app"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/grpchandler"
	pb "sys-metrics/internal/proto"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type AppGRPC struct {
	o    *server.Options
	app  *app.App
	grpc app.GRPCServer
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	ag, err := run(ctx, os.Args[1:])
	if err != nil {
		log.Fatalf("fatal error: %v", err)
	}

	<-ctx.Done()

	log.Println("shutting down application...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if errShutdown := ag.app.Close(ctxShutdown); errShutdown != nil {
		log.Printf("shutdown error: %v", errShutdown)
	}

	log.Println("application stopped")
}

func run(ctx context.Context, args []string) (*AppGRPC, error) {
	o, err := server.NewOption(common.ServerGRPC, args)
	if err != nil {
		return nil, labelerrors.NewLabelError("PARSE", fmt.Errorf("error parsing flags: %w", err))
	}

	ag, errI := initGRPCApp(ctx, o)
	if errI != nil {
		return nil, errI
	}

	cfg := server.NewConfig(server.SchemeHTTP, o.Host, o.Port, ag.app.Logger, ag.app.BackupConfig)
	lis, errL := net.Listen("tcp", cfg.InternalAddr())
	if errL != nil {
		return nil, labelerrors.NewLabelError("LISTEN", fmt.Errorf("failed to listen: %w", errL))
	}

	go func() {
		if errServe := ag.grpc.ListenAndServe(lis); errServe != nil &&
			!errors.Is(errServe, grpc.ErrServerStopped) {
			ag.app.Logger.Error("grpc server error", zap.Error(errServe))
		}
	}()

	return ag, nil
}

func initGRPCApp(ctx context.Context, o *server.Options) (*AppGRPC, error) {
	baseApp, err := app.Bootstrap(ctx, app.OptionFromServer(o))
	if err != nil {
		return nil, labelerrors.NewLabelError("BOOTSTRAP", fmt.Errorf("error bootstrapping: %w", err))
	}

	srv := grpc.NewServer()
	metricServer := &grpchandler.MetricServer{}
	pb.RegisterMetricsServer(srv, metricServer)

	shutdownGRPC := &app.ShutdownGRPCServer{Server: srv}
	baseApp.Server = shutdownGRPC

	return &AppGRPC{
		o:    o,
		app:  baseApp,
		grpc: shutdownGRPC,
	}, nil
}
