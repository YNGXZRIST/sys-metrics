package grpchandler

import (
	"context"
	pb "sys-metrics/internal/proto"
)

type MetricServer struct {
	pb.UnimplementedMetricsServer
}

func (ms *MetricServer) UpdateMetrics(context.Context, *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	return &pb.UpdateMetricsResponse{}, nil
}
