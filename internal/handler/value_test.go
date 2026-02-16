package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

			gaugeVal := 123.456
			err := svc.Gauges().Set(context.Background(), common.Alloc, &metrics.Gauge{Metrics: metrics.Metrics{ID: common.Alloc, MType: common.Gauge, Value: &gaugeVal}})
			if err != nil {
				t.Errorf("gauges.Set(%s): expected %v, got %v", common.Alloc, nil, err)
			}
			counterVal := int64(42)
			err = svc.Counters().Set(context.Background(), common.PollCount, &metrics.Counter{Metrics: metrics.Metrics{ID: common.PollCount, MType: common.Counter, Delta: &counterVal}})
			if err != nil {
				t.Errorf("counters.Set(%s): expected %v, got %v", common.PollCount, nil, err)
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

			gaugeVal := 123.456
			err := svc.Gauges().Set(context.Background(), common.Alloc, &metrics.Gauge{Metrics: metrics.Metrics{ID: common.Alloc, MType: common.Gauge, Value: &gaugeVal}})
			if err != nil {
				t.Errorf("gauges.Set(%s): expected %v, got %v", common.Alloc, nil, err)
			}
			counterVal := int64(42)
			err = svc.Counters().Set(context.Background(), common.PollCount, &metrics.Counter{Metrics: metrics.Metrics{ID: common.PollCount, MType: common.Counter, Delta: &counterVal}})
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
			ValueHandlerJSON(w, req)
			res := w.Result()
			res.Body.Close()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}(res.Body)

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
				Value: func() *float64 { v := 123.456; return &v }(),
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
				Delta: func() *int64 { v := int64(42); return &v }(),
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
					counterVal := int64(42)
					err := svc.Counters().Set(context.Background(), tt.args.name, &metrics.Counter{Metrics: metrics.Metrics{ID: tt.args.name, MType: common.Counter, Delta: &counterVal}})
					if err != nil {
						t.Errorf("counters.Set(%s): expected %v, got %v", tt.args.name, nil, err)
					}
				case common.Gauge:
					gaugeVal := 123.456
					err := svc.Gauges().Set(context.Background(), tt.args.name, &metrics.Gauge{Metrics: metrics.Metrics{ID: tt.args.name, MType: common.Gauge, Value: &gaugeVal}})
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
