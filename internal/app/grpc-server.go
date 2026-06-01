package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sys-metrics/internal/Interceptors"
	"sys-metrics/internal/grpchandler"
	pb "sys-metrics/internal/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type shutdownGRPCServer struct {
	server *grpc.Server
}

func (s *shutdownGRPCServer) shutdown(_ context.Context) error {
	if s.server == nil {
		return nil
	}
	s.server.GracefulStop()
	return nil
}

func (s *shutdownGRPCServer) listenAndServe(lis net.Listener, logger *zap.Logger) {
	go func() {
		if err := s.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("grpc server error", zap.Error(err))
		}
	}()
}

// startGRPC registers the metrics service and starts the gRPC listener.
func (a *App) startGRPC() error {
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(Interceptors.UnaryServerXRealIPInterceptor(a.ipNet)),
	)
	pb.RegisterMetricsServer(grpcSrv, grpchandler.NewMetricServer(a.metricsService, a.logger))

	a.serverGRPC.server = grpcSrv

	lis, err := net.Listen("tcp", a.opts.GRPCInternalAddr())
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	a.serverGRPC.listenAndServe(lis, a.logger)

	return nil
}
