package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/observer"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"testing"

	"go.uber.org/zap"
)

type noopObserver struct{}

func (noopObserver) Notify(ctx context.Context, data any) {}
func (noopObserver) Register(any) error                   { return nil }
func newBenchmarkHandler(tb testing.TB) *Handler {
	tb.Helper()
	return NewHandler(nil, nil, nil, zap.NewNop(), map[ObserverKey]observer.Observer{
		ObserverAudit: noopObserver{},
	}, serviceMetrics.NewService(noopObserver{}))
}

type gaugeDataset struct {
	names          []string
	valStrs        []string
	jsonBodies     [][]byte
	valueReqBodies [][]byte
}

func makeGaugeDataset(k int) gaugeDataset {
	r := rand.New(rand.NewSource(1))
	ds := gaugeDataset{
		names:          make([]string, k),
		valStrs:        make([]string, k),
		jsonBodies:     make([][]byte, k),
		valueReqBodies: make([][]byte, k),
	}
	for i := 0; i < k; i++ {
		name := fmt.Sprintf("m_%d_%x", i, r.Uint64())
		val := r.Float64() * 1e6
		valStr := strconv.FormatFloat(val, 'f', -1, 64)

		body, _ := json.Marshal(metrics.Metrics{
			ID:    name,
			MType: common.Gauge,
			Value: new(val),
		})
		valueReq, _ := json.Marshal(struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}{
			ID:   name,
			Type: common.Gauge,
		})

		ds.names[i] = name
		ds.valStrs[i] = valStr
		ds.jsonBodies[i] = body
		ds.valueReqBodies[i] = valueReq
	}
	return ds
}

type counterDataset struct {
	names          []string
	valStrs        []string
	jsonBodies     [][]byte
	valueReqBodies [][]byte
}

func makeCounterDataset(k int) counterDataset {
	r := rand.New(rand.NewSource(2))
	ds := counterDataset{
		names:          make([]string, k),
		valStrs:        make([]string, k),
		jsonBodies:     make([][]byte, k),
		valueReqBodies: make([][]byte, k),
	}
	for i := 0; i < k; i++ {
		name := fmt.Sprintf("c_%d_%x", i, r.Uint64())
		val := int64(r.Intn(1_000_000))
		valStr := strconv.FormatInt(val, 10)

		body, _ := json.Marshal(metrics.Metrics{
			ID:    name,
			MType: common.Counter,
			Delta: new(val),
		})
		valueReq, _ := json.Marshal(struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}{
			ID:   name,
			Type: common.Counter,
		})

		ds.names[i] = name
		ds.valStrs[i] = valStr
		ds.jsonBodies[i] = body
		ds.valueReqBodies[i] = valueReq
	}
	return ds
}

func BenchmarkHandler_UpdateHandler_Gauge(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)
	ds := makeGaugeDataset(4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j := i % len(ds.names)
		name := ds.names[j]
		value := ds.valStrs[j]

		req := httptest.NewRequest(http.MethodPost, "/update/"+common.Gauge+"/"+name+"/"+value, nil)
		req.SetPathValue("type", common.Gauge)
		req.SetPathValue("name", name)
		req.SetPathValue("value", value)

		w := httptest.NewRecorder()
		h.UpdateHandler(w, req)
		_ = w.Result().Body.Close()
	}
}

func BenchmarkHandler_UpdateHandlerJSON_Gauge(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)
	ds := makeGaugeDataset(4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		body := ds.jsonBodies[i%len(ds.jsonBodies)]
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.UpdateHandlerJSON(w, req)
		_ = w.Result().Body.Close()
	}
}

func BenchmarkHandler_ValueHandler_Gauge(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)
	ds := makeGaugeDataset(4096)

	ctx := context.Background()
	for i := 0; i < len(ds.names); i++ {
		v, err := strconv.ParseFloat(ds.valStrs[i], 64)
		if err != nil {
			b.Fatal(err)
		}
		if err := svc.Gauges().Set(ctx, ds.names[i], &metrics.Gauge{
			Metrics: metrics.Metrics{ID: ds.names[i], MType: common.Gauge, Value: &v},
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		name := ds.names[i%len(ds.names)]
		req := httptest.NewRequest(http.MethodGet, "/value/"+common.Gauge+"/"+name, nil)
		req.SetPathValue("type", common.Gauge)
		req.SetPathValue("name", name)

		w := httptest.NewRecorder()
		h.ValueHandler(w, req)
		_ = w.Result().Body.Close()
	}
}

func BenchmarkHandler_ValueHandlerJSON_Gauge(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)
	ds := makeGaugeDataset(4096)

	ctx := context.Background()
	for i := 0; i < len(ds.names); i++ {
		v, err := strconv.ParseFloat(ds.valStrs[i], 64)
		if err != nil {
			b.Fatal(err)
		}
		if err := svc.Gauges().Set(ctx, ds.names[i], &metrics.Gauge{
			Metrics: metrics.Metrics{ID: ds.names[i], MType: common.Gauge, Value: &v},
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reqBody := ds.valueReqBodies[i%len(ds.valueReqBodies)]
		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()
		h.ValueHandlerJSON(w, req)
		_ = w.Result().Body.Close()
	}
}

func BenchmarkHandler_UpdateHandler_Counter(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)
	ds := makeCounterDataset(4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j := i % len(ds.names)
		name := ds.names[j]
		value := ds.valStrs[j]

		req := httptest.NewRequest(http.MethodPost, "/update/"+common.Counter+"/"+name+"/"+value, nil)
		req.SetPathValue("type", common.Counter)
		req.SetPathValue("name", name)
		req.SetPathValue("value", value)

		w := httptest.NewRecorder()
		h.UpdateHandler(w, req)
		_ = w.Result().Body.Close()
	}
}

func BenchmarkHandler_UpdateHandlerJSON_Counter(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)
	ds := makeCounterDataset(4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		body := ds.jsonBodies[i%len(ds.jsonBodies)]
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.UpdateHandlerJSON(w, req)
		_ = w.Result().Body.Close()
	}
}

func BenchmarkHandler_ValueHandler_Counter(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)
	ds := makeCounterDataset(4096)

	ctx := context.Background()
	for i := 0; i < len(ds.names); i++ {
		v, err := strconv.ParseInt(ds.valStrs[i], 10, 64)
		if err != nil {
			b.Fatal(err)
		}
		if err := svc.Counters().Set(ctx, ds.names[i], &metrics.Counter{
			Metrics: metrics.Metrics{ID: ds.names[i], MType: common.Counter, Delta: &v},
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		name := ds.names[i%len(ds.names)]
		req := httptest.NewRequest(http.MethodGet, "/value/"+common.Counter+"/"+name, nil)
		req.SetPathValue("type", common.Counter)
		req.SetPathValue("name", name)

		w := httptest.NewRecorder()
		h.ValueHandler(w, req)
		_ = w.Result().Body.Close()
	}
}

func BenchmarkHandler_ValueHandlerJSON_Counter(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)
	ds := makeCounterDataset(4096)

	ctx := context.Background()
	for i := 0; i < len(ds.names); i++ {
		v, err := strconv.ParseInt(ds.valStrs[i], 10, 64)
		if err != nil {
			b.Fatal(err)
		}
		if err := svc.Counters().Set(ctx, ds.names[i], &metrics.Counter{
			Metrics: metrics.Metrics{ID: ds.names[i], MType: common.Counter, Delta: &v},
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reqBody := ds.valueReqBodies[i%len(ds.valueReqBodies)]
		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()
		h.ValueHandlerJSON(w, req)
		_ = w.Result().Body.Close()
	}
}

func BenchmarkHandler_UpdatesMetricsHandlerJSON_Batch(b *testing.B) {
	svc.Init(memory.NewService())
	h := newBenchmarkHandler(b)

	r := rand.New(rand.NewSource(3))
	const batchSize = 256
	req := make([]metrics.Metrics, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		if i%2 == 0 {
			name := fmt.Sprintf("bg_%d_%x", i, r.Uint64())
			val := r.Float64() * 1e6
			v := val
			req = append(req, metrics.Metrics{ID: name, MType: common.Gauge, Value: &v})
		} else {
			name := fmt.Sprintf("bc_%d_%x", i, r.Uint64())
			val := int64(r.Intn(1_000_000))
			v := val
			req = append(req, metrics.Metrics{ID: name, MType: common.Counter, Delta: &v})
		}
	}
	body, err := json.Marshal(req)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		httpReq := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.UpdatesMetricsHandlerJSON(w, httpReq)
		_ = w.Result().Body.Close()
	}
}
