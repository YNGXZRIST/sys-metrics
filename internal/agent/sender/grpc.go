package sender

import (
	"context"
	"fmt"
	"sys-metrics/internal/Interceptors"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	pb "sys-metrics/internal/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcSenderConfig struct {
	senderDeps
	endpoint string
}

type grpcSender struct {
	grpcSenderConfig
	conn   *grpc.ClientConn
	client pb.MetricsClient
}

var _ MetricsSender = (*grpcSender)(nil)

func newGrpcSender(cfg grpcSenderConfig) (*grpcSender, error) {
	s := &grpcSender{grpcSenderConfig: cfg}
	if err := s.initClient(); err != nil {
		return nil, err
	}
	return s, nil
}

func (g *grpcSender) initClient() error {
	conn, err := grpc.NewClient(
		g.endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(Interceptors.UnaryClientXRealIPInterceptor(g.localIPv4)),
	)
	if err != nil {
		return err
	}
	g.client = pb.NewMetricsClient(conn)
	g.conn = conn
	return nil
}

func (g *grpcSender) SendBatch(ctx context.Context, batch []*metrics.Metrics) error {
	if len(batch) == 0 {
		return fmt.Errorf("empty batch")
	}
	resp, err := g.client.UpdateMetrics(ctx, pb.UpdateMetricsRequest_builder{
		Metrics: g.getProtoMetrics(batch),
	}.Build())
	g.logger.Info("grpc send batch", zap.Any("batch", batch))
	if err != nil {
		return fmt.Errorf("grpc send batch: %w", err)
	}
	g.logger.Info("grpc send batch success", zap.Any("response", resp.String()))
	return nil
}

func (g *grpcSender) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}
func (g *grpcSender) getProtoMetrics(batch []*metrics.Metrics) []*pb.Metric {
	protoMetrics := make([]*pb.Metric, 0, len(batch))
	for _, m := range batch {
		if m == nil {
			continue
		}
		b := pb.Metric_builder{Id: m.ID}
		switch m.MType {
		case common.Counter:
			b.Type = pb.Metric_COUNTER
			if m.Delta != nil {
				b.Delta = *m.Delta
			}
		default:
			b.Type = pb.Metric_GAUGE
			if m.Value != nil {
				b.Value = *m.Value
			}
		}
		protoMetrics = append(protoMetrics, b.Build())
	}
	return protoMetrics
}
