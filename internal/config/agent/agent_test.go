package agent

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewConfig(t *testing.T) {
	logger := zap.NewExample()
	tests := []struct {
		name string
		in   InitProperties
	}{
		{
			name: "embeds_init_properties",
			in: InitProperties{
				PollInterval:   10 * time.Millisecond,
				ReportInterval: 20 * time.Millisecond,
				ServerAddr:     "localhost",
				Logger:         logger,
				RateLimit:      3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig(tt.in)
			if got == nil {
				t.Fatal("NewConfig() returned nil")
			}
			if got.PollInterval != tt.in.PollInterval {
				t.Errorf("PollInterval = %v, want %v", got.PollInterval, tt.in.PollInterval)
			}
			if got.ReportInterval != tt.in.ReportInterval {
				t.Errorf("ReportInterval = %v, want %v", got.ReportInterval, tt.in.ReportInterval)
			}
			if got.ServerAddr != tt.in.ServerAddr {
				t.Errorf("ServerAddr = %v, want %v", got.ServerAddr, tt.in.ServerAddr)
			}
			if got.RateLimit != tt.in.RateLimit {
				t.Errorf("RateLimit = %v, want %v", got.RateLimit, tt.in.RateLimit)
			}
			if got.Logger != tt.in.Logger {
				t.Errorf("Logger = %p, want %p", got.Logger, tt.in.Logger)
			}
			if got.InitProperties != tt.in {
				t.Errorf("InitProperties mismatch: got %+v, want %+v", got.InitProperties, tt.in)
			}
		})
	}
}
