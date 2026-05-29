package agent

import (
	"context"
	"sys-metrics/internal/agent/sender"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"testing"

	"go.uber.org/zap"
)

func TestNewMetricsSender(t *testing.T) {
	s, err := sender.NewMetricsSender(sender.Config{
		Transport: common.ReportTransportHTTP,
		ServerURL: testServer.URL,
		Logger:    zap.NewExample(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("NewMetricsSender() returned nil")
	}
}

func TestSender_SendBatch(t *testing.T) {
	tests := []struct {
		setup   func() *Collector
		name    string
		wantErr bool
	}{
		{
			name: "success",
			setup: func() *Collector {
				c := NewCollector(context.Background(), 1)
				g := models.NewGauge("random")
				g.SetValue(12.43)
				c.Gauges["random"] = g
				counter := models.NewCounter("random")
				counter.SetValue(12)
				c.Counters["random"] = counter
				return c
			},
		},
		{
			name: "empty id gauge",
			setup: func() *Collector {
				c := NewCollector(context.Background(), 1)
				g := models.NewGauge("")
				g.SetValue(0)
				c.Gauges[""] = g
				return c
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := tt.setup()
			err := testSender.SendBatch(context.Background(), metricsFromCollector(collector))
			if (err != nil) != tt.wantErr {
				t.Errorf("SendBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
