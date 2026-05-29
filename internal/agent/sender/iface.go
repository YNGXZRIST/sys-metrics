package sender

import (
	"context"
	models "sys-metrics/internal/model/metrics"

	"go.uber.org/zap"
)

type MetricsSender interface {
	SendBatch(ctx context.Context, metrics []*models.Metrics) error
	Close() error
}
type senderDeps struct {
	logger    *zap.Logger
	localIPv4 string
}
