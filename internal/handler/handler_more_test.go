package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"testing"

	"go.uber.org/zap"
)

func TestPingHandler_noDB(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	h.PingHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatesMetricsHandlerJSON_withObserver(t *testing.T) {
	svc.Init(memory.NewService())
	h := NewHandler(InitProperties{
		Logger:        zap.NewNop(),
		MetricService: serviceMetrics.NewService(noopObserver{}),
	})
	body := []byte(`[{"id":"obs_g","type":"gauge","value":2}]`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	h.UpdatesMetricsHandlerJSON(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateHandlerJSON_invalidBody(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader([]byte("not-json")))
	h.UpdateHandlerJSON(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatesMetricsHandlerJSON_invalidBody(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader([]byte("[")))
	h.UpdatesMetricsHandlerJSON(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestNewHandler_fields(t *testing.T) {
	h := NewHandler(InitProperties{
		Logger:        zap.NewNop(),
		MetricService: serviceMetrics.NewService(nil),
	})
	if h.Conn != nil || h.Authenticator != nil || h.RequestDecryptor != nil || h.Logger == nil {
		t.Fatalf("unexpected handler state %#v", h)
	}
}
