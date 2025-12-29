package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
	"testing"
)

func TestValueHandler(t *testing.T) {
	type args struct {
		metricType string
		name       string
	}
	tests := []struct {
		name       string
		args       args
		wantValue  string
		wantStatus int
	}{
		{
			name: "gauge success",
			args: args{
				metricType: metrics.MTypeGauge,
				name:       "Alloc",
			},
			wantValue:  "123.456",
			wantStatus: http.StatusOK,
		},
		{
			name: "counter success",
			args: args{
				metricType: metrics.MTypeCounter,
				name:       "PollCount",
			},
			wantValue:  "42",
			wantStatus: http.StatusOK,
		},
		{
			name: "unknown metric type",
			args: args{
				metricType: "unset",
				name:       "Alloc",
			},
			wantValue:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "metric not found",
			args: args{
				metricType: metrics.MTypeGauge,
				name:       "Unknown",
			},
			wantValue:  "",
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := memstorage.NewMemStorage[string, *metrics.Counter]()
			gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
			svc.Init(counters, gauges)

			gaugeVal := 123.456
			err := gauges.Set("Alloc", &metrics.Gauge{Metrics: metrics.Metrics{ID: "Alloc", MType: metrics.MTypeGauge, Value: &gaugeVal}})
			if err != nil {
				t.Errorf("gauges.Set(Alloc): expected %v, got %v", nil, err)
			}
			counterVal := int64(42)
			err = counters.Set("Pollcount", &metrics.Counter{Metrics: metrics.Metrics{ID: "Pollcount", MType: metrics.MTypeCounter, Delta: &counterVal}})
			if err != nil {
				t.Errorf("counters.Set(Pollcount): expected %v, got %v", nil, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/value", nil)
			req.SetPathValue("type", tt.args.metricType)
			req.SetPathValue("name", tt.args.name)
			w := httptest.NewRecorder()
			ValueHandler(w, req)
			res := w.Result()
			res.Body.Close()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}(res.Body)

			if res.StatusCode != tt.wantStatus {
				t.Errorf("ValueHandler() status = %v, want %v", res.StatusCode, tt.wantStatus)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
			if tt.wantValue != "" && string(body) != tt.wantValue {
				t.Errorf("ValueHandler() body = %q, want %q", string(body), tt.wantValue)
			}
		})
	}
}
