package agent

import (
	"log"
	"sys-metrics/internal/common"
	"testing"
)

func TestNewReporter(t *testing.T) {
	reporter := NewReporter(testServer.URL, log.Default())
	if reporter == nil {
		t.Fatal("NewReporter() returned nil")
	}
	if reporter.serverAddr != testServer.URL {
		t.Errorf("NewReporter() serverAddr = %v, want %v", reporter.serverAddr, testServer.URL)
	}

}

func TestReporter_BuildUpdateURL(t *testing.T) {

	type args struct {
		m string
		n string
		v string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{},
			want: testServer.URL + "/update///",
		},
		{
			name: "not empty",
			args: args{
				m: common.Gauge,
				n: "random",
				v: "10",
			},
			want: testServer.URL + "/update/" + common.Gauge + "/random/10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testReporter.BuildUpdateURL(tt.args.m, tt.args.n, tt.args.v); got != tt.want {
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
	type args struct {
		c Collector
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				c: Collector{
					map[string]map[string]float64{
						common.Gauge: {
							"random": 12.43,
						},
						common.Counter: {
							"random": 12.43,
						},
					},
				},
			},
		},
		{
			name: "error",
			args: args{
				c: Collector{
					map[string]map[string]float64{
						common.Gauge: {
							"": 0,
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := testReporter.Send(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReporter_sendMetricToServer(t *testing.T) {
	type args struct {
		metric string
		name   string
		value  float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "empty",
			args:    args{},
			wantErr: true,
		},
		{
			name: "not empty",
			args: args{
				metric: common.Gauge,
				name:   common.Gauge,
				value:  10.43,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testReporter.sendMetricToServer(tt.args.metric, tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("sendMetricToServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
