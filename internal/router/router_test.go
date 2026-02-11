package router

import (
	"net/http"
	"net/http/httptest"
	model "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
	"sys-metrics/pkg/storage"
	"testing"

	"go.uber.org/zap"
)

func TestGetRouter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	router := GetRouter(logger, nil)
	if router == nil {
		t.Fatal("GetRouter() returned nil")
	}
}

func TestRoutes(t *testing.T) {
	counters := storage.NewMemStorage[string, *model.Counter]()
	gauges := storage.NewMemStorage[string, *model.Gauge]()
	svc.Init(memory.NewService(counters, gauges))

	logger, _ := zap.NewDevelopment()
	router := GetRouter(logger, nil)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "update gauge success",
			method:     http.MethodPost,
			path:       "/update/gauge/test/123.45",
			wantStatus: http.StatusOK,
		},
		{
			name:       "update counter success",
			method:     http.MethodPost,
			path:       "/update/counter/test/100",
			wantStatus: http.StatusOK,
		},
		{
			name:       "update invalid type",
			method:     http.MethodPost,
			path:       "/update/invalid/test/100",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update invalid counter value",
			method:     http.MethodPost,
			path:       "/update/counter/test/not_a_number",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
