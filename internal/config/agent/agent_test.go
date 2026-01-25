package agent

import (
	"reflect"
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
		want *Config
	}{
		{
			name: "default",
			args: args{
				pollInterval:   10 * time.Millisecond,
				reportInterval: 10 * time.Millisecond,
				serverAddr:     "localhost",
				logger:         zap.NewExample(),
			},
			want: &Config{
				PollInterval:   10 * time.Millisecond,
				ReportInterval: 10 * time.Millisecond,
				ServerAddr:     "localhost",
				Logger:         zap.NewExample(),
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
