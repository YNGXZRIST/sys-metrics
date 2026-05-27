package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/observer"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"testing"

	"go.uber.org/zap"
)

func TestGetObserverByType_missing(t *testing.T) {
	h := newTestHandler(t)
	_, err := h.GetObserverByType(ObserverAudit)
	if err == nil {
		t.Fatal("expected error")
	}
}

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
		Logger: zap.NewNop(),
		Observers: map[ObserverKey]observer.Observer{
			ObserverAudit: noopObserver{},
		},
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

func TestNewHandler_fields(t *testing.T) {
	h := NewHandler(InitProperties{
		Logger:        zap.NewNop(),
		MetricService: serviceMetrics.NewService(nil),
	})
	if h.Conn != nil || h.Authenticator != nil || h.RequestDecryptor != nil || h.Logger == nil {
		t.Fatalf("unexpected handler state %#v", h)
	}
}
