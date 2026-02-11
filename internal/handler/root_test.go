package handler

import (
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	svm "sys-metrics/internal/repository/metrics"
	"sys-metrics/pkg/storage"
	"testing"

	"github.com/go-resty/resty/v2"
)

func TestIndexHandler(t *testing.T) {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	svm.Init(memory.NewService(counters, gauges))
	h := http.NewServeMux()
	h.HandleFunc("/", IndexHandler)
	srv := httptest.NewServer(h)
	defer srv.Close()
	req := resty.New().R()
	req.URL = srv.URL
	resp, err := req.Send()
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("content type not expected: %s", resp.Header().Get("Content-Type"))
	}
}
