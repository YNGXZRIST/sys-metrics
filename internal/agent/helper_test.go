package agent

import (
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"
)

var testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`OK`))
}))

var testReporter = NewReporter(testServer.URL, zap.NewNop())
