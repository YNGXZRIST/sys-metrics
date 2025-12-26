package handlers

import (
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
	"testing"
)

func Test_writeBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	writeBadRequest(w)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("writeBadRequest error, want %v got %v", http.StatusBadRequest, res.StatusCode)
	}
}
func Test_writeSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	writeSuccess(w)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("writeSuccess error, want %v got %v", http.StatusOK, res.StatusCode)
	}
}

func TestUpdateHandler(t *testing.T) {
	type args struct {
		metricType string
		name       string
		value      string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "success",
			args: args{
				metricType: "gauge",
				name:       "sys-metrics",
				value:      "1",
			},
			want: http.StatusOK,
		},
		{
			name: "error",
			args: args{
				metricType: "test",
				name:       "sys-metrics",
				value:      "1",
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := memstorage.NewMemStorage[string, *metrics.Counter]()
			gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
			svc.Init(counters, gauges)

			req := httptest.NewRequest(http.MethodGet, "/update", nil)
			req.SetPathValue("type", tt.args.metricType)
			req.SetPathValue("name", tt.args.name)
			req.SetPathValue("value", tt.args.value)

			w := httptest.NewRecorder()
			UpdateHandler(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want {
				t.Errorf("UpdateHandler() = %v, want %v", res.StatusCode, tt.want)
			}
		})
	}
}
