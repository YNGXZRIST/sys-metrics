package db

import (
	"strings"
	internal "sys-metrics/internal/config/db/internal"
	"sys-metrics/internal/config/server"
	"testing"
)

func TestNewCfg(t *testing.T) {
	tests := []struct {
		name string
		opt  *server.Options
		want string
	}{
		{
			name: "DSN from options",
			opt:  &server.Options{DNS: "postgres://localhost/db"},
			want: "postgres://localhost/db",
		},
		{
			name: "empty DSN",
			opt:  &server.Options{DNS: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCfg(tt.opt)
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
		name    string
		cfg     *internal.Config
		wantErr bool
		errText string
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
