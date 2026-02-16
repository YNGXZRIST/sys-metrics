package middleware

import (
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/common"
	"testing"

	"go.uber.org/zap"
)

func TestWithLoggerContext(t *testing.T) {
	logger := zap.NewNop()

	var got interface{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Context().Value(common.ContextLoggerKey)
	})

	handler := WithLoggerContext(logger)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got != logger {
		t.Errorf("context logger = %v, want same logger instance", got)
	}
}
