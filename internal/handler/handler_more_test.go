package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/observer"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
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

func TestGetIPFromRequest(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	if got := h.GetIPFromRequest(req); got != "192.0.2.1" {
		t.Fatalf("GetIPFromRequest = %q", got)
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
	h := NewHandler(nil, nil, nil, zap.NewNop(), map[ObserverKey]observer.Observer{
		ObserverAudit: noopObserver{},
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
	h := NewHandler(nil, nil, nil, zap.NewNop(), nil)
	if h.Conn != nil || h.Auth != nil || h.ReqDecryptor != nil || h.Logger == nil {
		t.Fatalf("unexpected handler state %#v", h)
	}
}
