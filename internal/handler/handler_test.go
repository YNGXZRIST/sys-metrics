package handler

import (
	serviceMetrics "sys-metrics/internal/service/metrics"
	"testing"

	"go.uber.org/zap"
)

// newTestHandler returns a Handler wired like the in-memory server path (no DB, no auth, no observers).
func newTestHandler(tb testing.TB) *Handler {
	tb.Helper()
	return NewHandler(InitProperties{
		Logger:        zap.NewNop(),
		MetricService: serviceMetrics.NewService(nil),
	})
}
