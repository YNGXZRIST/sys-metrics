package main

import (
	"context"
	"sys-metrics/internal/common"
	config "sys-metrics/internal/config/agent"
	"testing"
	"time"
)

func TestInitAgent(t *testing.T) {
	tests := []struct {
		name    string
		opt     *config.Options
		wantErr bool
	}{
		{
			name: "valid options returns agent",
			opt: &config.Options{
				Host:           "localhost",
				Port:           "8080",
				Mode:           common.TypeModeTest,
				PollInterval:   time.Second,
				ReportInterval: 2 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "development mode returns agent",
			opt: &config.Options{
				Host:           "127.0.0.1",
				Port:           "9090",
				Mode:           common.TypeModeDevelopment,
				PollInterval:   time.Second,
				ReportInterval: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid mode returns error",
			opt: &config.Options{
				Host:           "localhost",
				Port:           "8080",
				Mode:           "invalid",
				PollInterval:   time.Second,
				ReportInterval: 2 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := initAgent(tt.opt, context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("initAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if a == nil {
					t.Error("initAgent() returned nil agent")
					return
				}
				if a.Logger == nil {
					t.Error("initAgent() agent.Logger is nil")
				}
				if a.ServerAddr == "" {
					t.Error("initAgent() agent.ServerAddr is empty")
				}
				if want := "http://" + tt.opt.Host + ":" + tt.opt.Port; a.ServerAddr != want {
					t.Errorf("initAgent() agent.ServerAddr = %q, want %q", a.ServerAddr, want)
				}
			}
		})
	}
}
