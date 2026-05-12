package router_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"sys-metrics/internal/common"
	"sys-metrics/internal/handler"
	"sys-metrics/internal/repository/memory"
	svc "sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/router"

	"go.uber.org/zap"
)

// newExampleRouter builds a chi router with an empty in-memory metrics store (no DB, no auth).
func newExampleRouter() http.Handler {
	svc.Init(memory.NewService())
	h := handler.NewHandler(nil, nil, nil, zap.NewNop(), nil)
	return router.GetRouter(h)
}

// Plain-text API: POST /update/{type}/{name}/{value}
func Example_plainTextUpdate_gauge() {
	h := newExampleRouter()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/heap_alloc/12345.5", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}

func Example_plainTextUpdate_counter() {
	h := newExampleRouter()
	req := httptest.NewRequest(http.MethodPost, "/update/counter/poll_count/7", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}

func Example_plainTextValue_gauge() {
	h := newExampleRouter()
	post := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/36.6", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, post)

	get := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)
	out := httptest.NewRecorder()
	h.ServeHTTP(out, get)
	body, _ := io.ReadAll(out.Body)
	fmt.Println(out.Code, strings.TrimSpace(string(body)))
	// Output:
	// 200 36.6
}

func Example_plainTextValue_counter() {
	h := newExampleRouter()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/update/counter/requests/10", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/value/counter/requests", nil))
	body, _ := io.ReadAll(rec.Body)
	fmt.Println(rec.Code, strings.TrimSpace(string(body)))
	// Output:
	// 200 10
}

// JSON API: POST /update with Content-Type: application/json
func Example_jsonUpdateMetric() {
	h := newExampleRouter()
	body := `{"id":"m1","type":"gauge","value":99.5}`
	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(body))
	req.Header.Set(common.ContentTypeHeader, common.ApplicationJSON)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}

// JSON API: POST /value with a JSON body describing which metric to read.
// Metric names are normalized (see agent.GetMetricType); use the same id the server stores.
func Example_jsonGetValue() {
	h := newExampleRouter()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/12345", nil))

	body := `{"id":"Alloc","type":"gauge"}`
	req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(body))
	req.Header.Set(common.ContentTypeHeader, common.ApplicationJSON)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}

// JSON API: POST /updates with a batch of metrics
func Example_jsonBatchUpdates() {
	h := newExampleRouter()
	body := `[{"id":"c1","type":"counter","delta":3},{"id":"g1","type":"gauge","value":2.5}]`
	req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(body))
	req.Header.Set(common.ContentTypeHeader, common.ApplicationJSON)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}

// GET /ping — liveness; with no database configured it always returns 200
func Example_ping() {
	h := newExampleRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}
