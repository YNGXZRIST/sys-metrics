package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
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
				metricType: common.Gauge,
				name:       "Alloc",
			},
			wantValue:  "123.456",
			wantStatus: http.StatusOK,
		},
		{
			name: "counter success",
			args: args{
				metricType: common.Counter,
				name:       "PollCount",
			},
			wantValue:  "42",
			wantStatus: http.StatusOK,
		},
		{
			name: "unknown metric type",
			args: args{
				metricType: "unset",
				name:       common.Alloc,
			},
			wantValue:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "metric not found",
			args: args{
				metricType: common.Gauge,
				name:       "Unknown",
			},
			wantValue:  "",
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc.Init(memory.NewService())

			err := svc.Gauges().Set(context.Background(), common.Alloc, &metrics.Gauge{Metrics: metrics.Metrics{ID: common.Alloc, MType: common.Gauge, Value: new(123.456)}})
			if err != nil {
				t.Errorf("gauges.Set(%s): expected %v, got %v", common.Alloc, nil, err)
			}
			err = svc.Counters().Set(context.Background(), common.PollCount, &metrics.Counter{Metrics: metrics.Metrics{ID: common.PollCount, MType: common.Counter, Delta: new(int64(42))}})
			if err != nil {
				t.Errorf("counters.Set(%s): expected %v, got %v", common.PollCount, nil, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/value/"+tt.args.metricType+"/"+tt.args.name, nil)
			req.SetPathValue("type", tt.args.metricType)
			req.SetPathValue("name", tt.args.name)
			w := httptest.NewRecorder()
			h := newTestHandler(t)
			h.ValueHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

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

func TestValueHandlerJSON(t *testing.T) {
	type args struct {
		ID   string `json:"id"`
		Type string `json:"type"`
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
				Type: common.Gauge,
				ID:   "Alloc",
			},
			wantValue:  "123.456",
			wantStatus: http.StatusOK,
		},
		{
			name: "counter success",
			args: args{
				Type: common.Counter,
				ID:   "PollCount",
			},
			wantValue:  "42",
			wantStatus: http.StatusOK,
		},
		{
			name: "unknown metric type",
			args: args{
				Type: "unset",
				ID:   common.Alloc,
			},
			wantValue:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "metric not found",
			args: args{
				Type: common.Gauge,
				ID:   "Unknown",
			},
			wantValue:  "",
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc.Init(memory.NewService())

			err := svc.Gauges().Set(context.Background(), common.Alloc, &metrics.Gauge{Metrics: metrics.Metrics{ID: common.Alloc, MType: common.Gauge, Value: new(123.456)}})
			if err != nil {
				t.Errorf("gauges.Set(%s): expected %v, got %v", common.Alloc, nil, err)
			}
			err = svc.Counters().Set(context.Background(), common.PollCount, &metrics.Counter{Metrics: metrics.Metrics{ID: common.PollCount, MType: common.Counter, Delta: new(int64(42))}})
			if err != nil {
				t.Errorf("counters.Set(%s): expected %v, got %v", common.PollCount, nil, err)
			}
			payload, err := json.Marshal(tt.args)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}
			reqBody := bytes.NewReader(payload)
			req := httptest.NewRequest(http.MethodPost, "/value", reqBody)
			w := httptest.NewRecorder()
			h := newTestHandler(t)
			h.ValueHandlerJSON(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("ValueHandlerJSON() status = %v, want %v", res.StatusCode, tt.wantStatus)
			}

			body, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
			if tt.wantValue != "" && !bytes.Contains(body, []byte(tt.wantValue)) {
				t.Errorf("ValueHandlerJSON() body = %q, want to contain %q", string(body), tt.wantValue)
			}

		})
	}
}

func Test_getMetricFromStorage(t *testing.T) {
	type args struct {
		metricType string
		name       string
		isWantSet  bool
	}
	tests := []struct {
		name    string
		args    args
		want    metrics.Metrics
		wantErr bool
	}{
		{
			name: "gauge found",
			args: args{
				metricType: common.Gauge,
				name:       "Alloc",
				isWantSet:  true,
			},
			want: metrics.Metrics{
				ID:    "Alloc",
				MType: common.Gauge,
				Value: func() *float64 { return new(123.456) }(),
			},
			wantErr: false,
		},
		{
			name: "counter found",
			args: args{
				metricType: common.Counter,
				name:       "PollCount",
				isWantSet:  true,
			},
			want: metrics.Metrics{
				ID:    "PollCount",
				MType: common.Counter,
				Delta: func() *int64 { return new(int64(42)) }(),
			},
			wantErr: false,
		},
		{
			name: "unknown metric type",
			args: args{
				metricType: "unset",
				name:       "Alloc",
				isWantSet:  false,
			},
			want:    metrics.Metrics{},
			wantErr: true,
		},
		{
			name: "metric not found",
			args: args{
				metricType: common.Gauge,
				name:       "Unknown",
				isWantSet:  false,
			},
			want:    metrics.Metrics{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc.Init(memory.NewService())
			if tt.args.isWantSet {
				switch tt.args.metricType {
				case common.Counter:
					err := svc.Counters().Set(context.Background(), tt.args.name, &metrics.Counter{Metrics: metrics.Metrics{ID: tt.args.name, MType: common.Counter, Delta: new(int64(42))}})
					if err != nil {
						t.Errorf("counters.Set(%s): expected %v, got %v", tt.args.name, nil, err)
					}
				case common.Gauge:
					err := svc.Gauges().Set(context.Background(), tt.args.name, &metrics.Gauge{Metrics: metrics.Metrics{ID: tt.args.name, MType: common.Gauge, Value: new(123.456)}})
					if err != nil {
						t.Errorf("gauges.Set(%s): expected %v, got %v", tt.args.name, nil, err)
					}
				}
			}
			got, err := getMetricFromStorage(context.Background(), tt.args.metricType, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMetricFromStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMetricFromStorage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
