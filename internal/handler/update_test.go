package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
	"sys-metrics/pkg/storage"
	"testing"
)

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
				metricType: common.Gauge,
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
			counters := storage.NewMemStorage[string, *metrics.Counter]()
			gauges := storage.NewMemStorage[string, *metrics.Gauge]()
			svc.Init(memory.NewService(counters, gauges))

			req := httptest.NewRequest(http.MethodGet, "/update", nil)
			req.SetPathValue("type", tt.args.metricType)
			req.SetPathValue("name", tt.args.name)
			req.SetPathValue("value", tt.args.value)

			w := httptest.NewRecorder()
			UpdateHandler(w, req)

			res := w.Result()
			res.Body.Close()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}(res.Body)

			if res.StatusCode != tt.want {
				t.Errorf("UpdateHandler() = %v, want %v", res.StatusCode, tt.want)
			}
		})
	}
}

func TestUpdateHandlerJSON(t *testing.T) {
	tests := []struct {
		name       string
		args       metrics.Metrics
		wantStatus int
	}{
		{
			name: "success gauge",
			args: metrics.Metrics{
				ID:    "sys-metrics",
				MType: common.Gauge,
				Value: func() *float64 { v, _ := strconv.ParseFloat("123.45", 64); return &v }(),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "success counter",
			args: metrics.Metrics{
				ID:    "sys-metrics",
				MType: common.Counter,
				Delta: func() *int64 { v, _ := strconv.ParseInt("100", 10, 64); return &v }(),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "unknown type",
			args: metrics.Metrics{
				ID:    "sys-metrics",
				MType: "test",
				Delta: func() *int64 { v, _ := strconv.ParseInt("100", 10, 64); return &v }(),
			},
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := storage.NewMemStorage[string, *metrics.Counter]()
			gauges := storage.NewMemStorage[string, *metrics.Gauge]()
			svc.Init(memory.NewService(counters, gauges))
			payload, err := json.Marshal(tt.args)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}
			reqBody := bytes.NewReader(payload)
			req := httptest.NewRequest(http.MethodPost, "/value", reqBody)
			w := httptest.NewRecorder()
			UpdateHandlerJSON(w, req)
			res := w.Result()
			res.Body.Close()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}(res.Body)

			if res.StatusCode != tt.wantStatus {
				t.Errorf("UpdateHandlerJSON() status = %v, want %v", res.StatusCode, tt.wantStatus)
			}
			if tt.args.MType == common.Gauge {
				gauge, err := gauges.Get(tt.args.ID)
				if err != nil {
					t.Fatalf("gauges.Get(%s): expected %v, got %v", tt.args.ID, nil, err)
				}
				if gauge.Value == nil {
					t.Fatalf("gauge.Value is nil")
				}
				if *gauge.Value != *tt.args.Value {
					t.Errorf("gauge value = %v, want %v", *gauge.Value, *tt.args.Value)
				}
			}
			if tt.args.MType == common.Counter {
				counter, err := counters.Get(tt.args.ID)
				if err != nil {
					t.Fatalf("counters.Get(%s): expected %v, got %v", tt.args.ID, nil, err)
				}
				if counter.Delta == nil {
					t.Fatalf("counter.Delta is nil")
				}
				if *counter.Delta != *tt.args.Delta {
					t.Errorf("counter value = %v, want %v", *counter.Delta, *tt.args.Delta)
				}
			}
		})
	}
}
