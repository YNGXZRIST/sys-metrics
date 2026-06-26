package db

import (
	"strings"
	internal "sys-metrics/internal/config/db/internal"
	"testing"
)

func TestNewCfg(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "DSN from options",
			cfg:  &Config{DNS: "postgres://localhost/db"},
			want: "postgres://localhost/db",
		},
		{
			name: "empty DSN",
			cfg:  &Config{DNS: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCfg(tt.cfg)
			if got == nil {
				t.Fatal("NewCfg() returned nil")
			}
			if got.DNS != tt.want {
				t.Errorf("NewCfg().DNS = %q, want %q", got.DNS, tt.want)
			}
		})
	}
}

func TestNewConn(t *testing.T) {
	tests := []struct {
		cfg     *internal.Config
		name    string
		errText string
		wantErr bool
	}{
		{
			name:    "nil config returns error",
			cfg:     nil,
			wantErr: true,
			errText: "database DSN is not set",
		},
		{
			name:    "empty DSN returns error",
			cfg:     &internal.Config{DNS: ""},
			wantErr: true,
			errText: "database DSN is not set",
		},
		{
			name:    "valid DSN returns connection",
			cfg:     &internal.Config{DNS: "postgres://user:pass@localhost:5432/testdb?sslmode=disable"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := NewConn(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != nil && tt.errText != "" && !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("NewConn() error = %q, want containing %q", err.Error(), tt.errText)
				}
				return
			}
			if conn == nil {
				t.Fatal("NewConn() returned nil connection")
			}
			if conn.Config != tt.cfg {
				t.Error("NewConn().Config != cfg")
			}
			_ = conn.Close()
		})
	}
}
