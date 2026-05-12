package handler

import (
	"testing"

	"go.uber.org/zap"
)

// newTestHandler returns a Handler wired like the in-memory server path (no DB, no auth, no observers).
func newTestHandler(tb testing.TB) *Handler {
	tb.Helper()
	return NewHandler(nil, nil, nil, zap.NewNop(), nil)
}
