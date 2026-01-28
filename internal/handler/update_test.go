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
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
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
	type args struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{
			name: "success gauge",
			args: args{
				ID:    "sys-metrics",
				Type:  common.Gauge,
				Value: "123.45",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "success counter",
			args: args{
				ID:    "sys-metrics",
				Type:  common.Counter,
				Value: "100",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "unknown type",
			args: args{
				ID:    "sys-metrics",
				Type:  "test",
				Value: "100",
			},
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := memstorage.NewMemStorage[string, *metrics.Counter]()
			gauges := memstorage.NewMemStorage[string, *metrics.Gauge]()
			svc.Init(counters, gauges)
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
				t.Errorf("UpdateHandlerJSON() status = %v, want %v", res.StatusCode, http.StatusOK)
			}
			if tt.args.Type == common.Gauge {
				gauge, err := gauges.Get(tt.args.ID)
				if err != nil {
					t.Fatalf("gauges.Get(%s): expected %v, got %v", tt.args.ID, nil, err)
				}
				expectedValue, err := strconv.ParseFloat(tt.args.Value, 64)
				if err != nil {
					t.Fatalf("strconv.ParseFloat(%s): expected %v, got %v", tt.args.Value, nil, err)
				}
				if *gauge.Value != expectedValue {
					t.Errorf("gauge value = %v, want %v", *gauge.Value, expectedValue)
				}
			}
			if tt.args.Type == common.Counter {
				counter, err := counters.Get(tt.args.ID)
				if err != nil {
					t.Fatalf("counters.Get(%s): expected %v, got %v", tt.args.ID, nil, err)
				}
				expectedValue := tt.args.Value
				if fmt.Sprintf("%d", *counter.Delta) != expectedValue {
					t.Errorf("counter value = %v, want %v", *counter.Delta, expectedValue)
				}
			}

		})
	}
}
