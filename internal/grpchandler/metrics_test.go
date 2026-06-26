package grpchandler

import (
	"context"
	"sys-metrics/internal/common"
	pb "sys-metrics/internal/proto"
	"sys-metrics/internal/repository/memory"
	repoMetrics "sys-metrics/internal/repository/metrics"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testMetricService(t *testing.T) *serviceMetrics.MetricService {
	t.Helper()
	repoMetrics.Init(memory.NewService())
	return serviceMetrics.NewService(nil)
}

func TestMetricServer_UpdateMetrics_empty(t *testing.T) {
	svc := testMetricService(t)
	s := NewMetricServer(svc, zap.NewNop())

	resp, err := s.UpdateMetrics(context.Background(), pb.UpdateMetricsRequest_builder{}.Build())
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
}

func TestNewMetricServer(t *testing.T) {
	svc := testMetricService(t)
	s := NewMetricServer(svc, zap.NewNop())
	if s.MetricService != svc || s.Logger == nil {
		t.Fatal("unexpected server fields")
	}
}

func TestMetricServer_UpdateMetrics_success(t *testing.T) {
	svc := testMetricService(t)
	s := NewMetricServer(svc, zap.NewNop())

	resp, err := s.UpdateMetrics(context.Background(), pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: "HeapAlloc", Type: pb.Metric_GAUGE, Value: 42}.Build(),
			pb.Metric_builder{Id: "PollCount", Type: pb.Metric_COUNTER, Delta: 1}.Build(),
		},
	}.Build())
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
}

func TestMetricServer_UpdateMetrics_invalidType(t *testing.T) {
	svc := testMetricService(t)
	s := NewMetricServer(svc, zap.NewNop())

	_, err := s.UpdateMetrics(context.Background(), pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: "bad", Type: pb.Metric_MType(99)}.Build(),
		},
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("err = %v, want InvalidArgument", err)
	}
}

func TestProtoToModels_unknownType(t *testing.T) {
	_, err := protoToModels([]*pb.Metric{
		pb.Metric_builder{Id: "x", Type: pb.Metric_MType(99)}.Build(),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProtoToModels_nilMetric(t *testing.T) {
	got, err := protoToModels([]*pb.Metric{nil})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d", len(got))
	}
}

func TestProtoToModels_counterAndGauge(t *testing.T) {
	got, err := protoToModels([]*pb.Metric{
		pb.Metric_builder{Id: common.Alloc, Type: pb.Metric_GAUGE, Value: 3.14}.Build(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].MType != common.Gauge {
		t.Fatalf("type = %q", got[0].MType)
	}
}
