// Package grpchandler implements the gRPC Metrics service (UpdateMetrics) on top of MetricService.
package grpchandler

import (
	"context"
	"fmt"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	pb "sys-metrics/internal/proto"
	"sys-metrics/internal/service/metrics"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricServer implements gRPC Metrics service.
type MetricServer struct {
	pb.UnimplementedMetricsServer
	MetricService *metrics.MetricService
	Logger        *zap.Logger
}

// NewMetricServer creates a gRPC metrics handler with the shared metrics service.
func NewMetricServer(svc *metrics.MetricService, logger *zap.Logger) *MetricServer {
	return &MetricServer{
		MetricService: svc,
		Logger:        logger,
	}
}

func (s *MetricServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	req, err := protoToModels(in.GetMetrics())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.MetricService.BatchUpdateMetrics(ctx, req); err != nil {
		s.Logger.Warn("UpdateMetrics failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to save metrics")
	}

	return &pb.UpdateMetricsResponse{}, nil
}

func protoToModels(in []*pb.Metric) ([]models.Metrics, error) {
	out := make([]models.Metrics, 0, len(in))
	for _, m := range in {
		if m == nil {
			continue
		}

		item := models.Metrics{
			ID: m.GetId(),
		}

		switch m.GetType() {
		case pb.Metric_COUNTER:
			item.MType = common.Counter
			item.Delta = new(m.GetDelta())
		case pb.Metric_GAUGE:
			item.MType = common.Gauge
			item.Value = new(m.GetValue())
		default:
			return nil, fmt.Errorf("unknown metric type for %q", m.GetId())
		}

		out = append(out, item)
	}
	return out, nil
}
