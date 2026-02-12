package handler

import (
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/repository/memory"
	svm "sys-metrics/internal/repository/metrics"
	"testing"

	"github.com/go-resty/resty/v2"
)

func TestIndexHandler(t *testing.T) {
	svm.Init(memory.NewService())
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
