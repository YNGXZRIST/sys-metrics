package agent

import (
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/agent/sender"
	"sys-metrics/internal/common"

	"go.uber.org/zap"
)

var testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`OK`))
}))

var testSender = mustNewSender(sender.SenderConfig{
	Transport: common.ReportTransportHTTP,
	ServerURL: testServer.URL,
	Logger:    zap.NewNop(),
})

func mustNewSender(cfg sender.SenderConfig) sender.MetricsSender {
	s, err := sender.NewMetricsSender(cfg)
	if err != nil {
		panic(err)
	}
	return s
}
