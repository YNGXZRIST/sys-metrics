package metrics

import (
	"context"
	"sys-metrics/internal/middleware"
	"sys-metrics/internal/observer"
	"time"
)

type AuditNotifier interface {
	Notify(ctx context.Context, data any)
}

type MetricService struct {
	AuditNotifier
}

func NewService(audit AuditNotifier) *MetricService {
	return &MetricService{AuditNotifier: audit}
}

func (ms *MetricService) notifyUpdated(ctx context.Context, ids []string) {
	if ms == nil || ms.AuditNotifier == nil {
		return
	}
	event := observer.MetricsEvent{
		TS:      time.Now().UTC().Unix(),
		Metrics: ids,
	}
	if ip, ok := ctx.Value(middleware.CtxClientIPKey).(string); ok {
		event.IP = ip
	}
	ms.Notify(ctx, event)
}
