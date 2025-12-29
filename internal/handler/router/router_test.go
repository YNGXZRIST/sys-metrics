package router

import (
	"net/http"
	"net/http/httptest"
	model "sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
	"testing"
)

func TestGetRouter(t *testing.T) {
	router := GetRouter()
	if router == nil {
		t.Fatal("GetRouter() returned nil")
	}
}

func TestRoutes(t *testing.T) {
	counters := memstorage.NewMemStorage[string, *model.Counter]()
	gauges := memstorage.NewMemStorage[string, *model.Gauge]()
	svc.Init(counters, gauges)

	router := GetRouter()

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
