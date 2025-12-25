package agent

import (
	"log"
	"net/http"
	"net/http/httptest"
)

var testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`OK`))
}))
var testReporter = &Reporter{
	serverAddr: testServer.URL,
	logger:     log.Default(),
}
