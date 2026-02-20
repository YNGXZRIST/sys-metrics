package agent

import (
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"testing"

	"go.uber.org/zap"
)

func TestNewReporter(t *testing.T) {
	reporter := NewReporter(testServer.URL, zap.NewExample(), nil)
	if reporter == nil {
		t.Fatal("NewReporter() returned nil")
	}
	if reporter.serverAddr != testServer.URL {
		t.Errorf("NewReporter() serverAddr = %v, want %v", reporter.serverAddr, testServer.URL)
	}

}

func TestReporter_BuildUpdateURL(t *testing.T) {

	tests := []struct {
		name string
		want string
	}{
		{
			name: "empty",
			want: testServer.URL + "/update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testReporter.BuildUpdateURL(); got != tt.want {
				t.Errorf("BuildUpdateURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReporter_ConvertMetricValue(t *testing.T) {
	type args struct {
		m string
		v float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{},
			want: "0",
		},
		{
			name: common.Gauge,
			args: args{
				m: common.Gauge,
				v: 10.43,
			},
			want: "10.43",
		},
		{
			name: common.Counter,
			args: args{
				m: common.Counter,
				v: 12.43,
			},
			want: "12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testReporter.ConvertMetricValue(tt.args.m, tt.args.v); got != tt.want {
				t.Errorf("ConvertMetricValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReporter_Send(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Collector
		wantErr bool
	}{
		{
			name: "success",
			setup: func() *Collector {
				c := NewCollector()
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
			name: "error",
			setup: func() *Collector {
				c := NewCollector()
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
			if err := testReporter.sendMetricsToServer(collector); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReporter_sendMetricToServer(t *testing.T) {
	tests := []struct {
		name    string
		metric  *models.Metrics
		wantErr bool
	}{
		{
			name:    "empty",
			metric:  &models.Metrics{},
			wantErr: false,
		},
		{
			name: "not empty",
			metric: &models.Metrics{
				ID:    common.Gauge,
				MType: common.Gauge,
				Value: func() *float64 { v := 10.43; return &v }(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testReporter.sendMetricToServer(*tt.metric); (err != nil) != tt.wantErr {
				t.Errorf("sendMetricToServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
