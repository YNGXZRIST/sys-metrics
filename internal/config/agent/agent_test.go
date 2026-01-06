package agent

import (
	"log"
	"reflect"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	type args struct {
		pollInterval   time.Duration
		reportInterval time.Duration
		serverAddr     string
		logger         *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		{
			name: "default",
			args: args{
				pollInterval:   10 * time.Millisecond,
				reportInterval: 10 * time.Millisecond,
				serverAddr:     "localhost",
				logger:         log.Default(),
			},
			want: &Config{
				PollInterval:   10 * time.Millisecond,
				ReportInterval: 10 * time.Millisecond,
				ServerAddr:     "localhost",
				Logger:         log.Default(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(tt.args.pollInterval, tt.args.reportInterval, tt.args.serverAddr, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
