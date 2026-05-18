package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
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

			svc.Init(memory.NewService())

			req := httptest.NewRequest(http.MethodPost, "/update/"+tt.args.metricType+"/"+tt.args.name+"/"+tt.args.value, nil)
			req.SetPathValue("type", tt.args.metricType)
			req.SetPathValue("name", tt.args.name)
			req.SetPathValue("value", tt.args.value)

			w := httptest.NewRecorder()
			h := newTestHandler(t)
			h.UpdateHandler(w, req)

			res := w.Result()
			defer res.Body.Close()

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
			svc.Init(memory.NewService())
			payload, err := json.Marshal(tt.args)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}
			reqBody := bytes.NewReader(payload)
			req := httptest.NewRequest(http.MethodPost, "/update", reqBody)
			w := httptest.NewRecorder()
			h := newTestHandler(t)
			h.UpdateHandlerJSON(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("UpdateHandlerJSON() status = %v, want %v", res.StatusCode, tt.wantStatus)
			}
			if tt.args.MType == common.Gauge {
				gauge, err := svc.Gauges().Get(context.Background(), tt.args.ID)
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
				counter, err := svc.Counters().Get(context.Background(), tt.args.ID)
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

func TestUpdatesMetricsHandlerJSON(t *testing.T) {
	svc.Init(memory.NewService())
	h := newTestHandler(t)

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString("not-json"))
		rec := httptest.NewRecorder()
		h.UpdatesMetricsHandlerJSON(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", rec.Code)
		}
	})

	t.Run("success batch", func(t *testing.T) {
		v := 1.5
		d := int64(2)
		body := []byte(`[{"id":"bg1","type":"gauge","value":1.5},{"id":"bc1","type":"counter","delta":2}]`)
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		h.UpdatesMetricsHandlerJSON(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}
		g, err := svc.Gauges().Get(context.Background(), "bg1")
		if err != nil {
			t.Fatal(err)
		}
		if g.Value == nil || *g.Value != v {
			t.Fatalf("gauge value = %v want %v", g.Value, v)
		}
		c, err := svc.Counters().Get(context.Background(), "bc1")
		if err != nil {
			t.Fatal(err)
		}
		if c.Delta == nil || *c.Delta != d {
			t.Fatalf("counter delta = %v want %v", c.Delta, d)
		}
	})
}
