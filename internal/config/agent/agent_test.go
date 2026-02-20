package agent

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewConfig(t *testing.T) {
	type args struct {
		pollInterval   time.Duration
		reportInterval time.Duration
		serverAddr     string
		logger         *zap.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				pollInterval:   10 * time.Millisecond,
				reportInterval: 10 * time.Millisecond,
				serverAddr:     "localhost",
				logger:         zap.NewExample(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig(tt.args.pollInterval, tt.args.reportInterval, tt.args.serverAddr, tt.args.logger, nil)
			if got.PollInterval != tt.args.pollInterval {
				t.Errorf("NewConfig().PollInterval = %v, want %v", got.PollInterval, tt.args.pollInterval)
			}
			if got.ReportInterval != tt.args.reportInterval {
				t.Errorf("NewConfig().ReportInterval = %v, want %v", got.ReportInterval, tt.args.reportInterval)
			}
			if got.ServerAddr != tt.args.serverAddr {
				t.Errorf("NewConfig().ServerAddr = %v, want %v", got.ServerAddr, tt.args.serverAddr)
			}
			if got.Logger == nil {
				t.Errorf("NewConfig().Logger is nil, want non-nil logger")
			}
		})
	}
}
