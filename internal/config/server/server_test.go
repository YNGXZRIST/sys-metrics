package server

import (
	file "sys-metrics/internal/repository/file"
	"testing"

	"go.uber.org/zap"
)

func TestConfig_InternalAddr(t *testing.T) {
	type fields struct {
		logger *zap.Logger
		scheme string
		host   string
		port   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "default",
			fields: fields{
				scheme: "http",
				host:   "localhost",
				port:   "8080",
				logger: zap.NewExample(),
			},
			want: "localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				scheme: tt.fields.scheme,
				host:   tt.fields.host,
				port:   tt.fields.port,
				logger: tt.fields.logger,
			}
			if got := c.InternalAddr(); got != tt.want {
				t.Errorf("InternalAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_ServerAddr(t *testing.T) {
	type fields struct {
		logger *zap.Logger
		scheme string
		host   string
		port   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "default",
			fields: fields{
				scheme: SchemeHTTP,
				host:   DefaultHost,
				port:   DefaultPort,
				logger: zap.NewNop(),
			},
			want: "http://localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				scheme: tt.fields.scheme,
				host:   tt.fields.host,
				port:   tt.fields.port,
				logger: tt.fields.logger,
			}
			if got := c.ServerAddr(); got != tt.want {
				t.Errorf("ServerAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	type args struct {
		logger       *zap.Logger
		backupConfig *file.Config
		s            string
		h            string
		p            string
	}
	tests := []struct {
		args args
		name string
	}{
		{
			name: "default",
			args: args{
				s:            SchemeHTTP,
				h:            DefaultHost,
				p:            DefaultPort,
				logger:       zap.NewExample(),
				backupConfig: &file.Config{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig(tt.args.s, tt.args.h, tt.args.p, tt.args.logger, tt.args.backupConfig)
			if got.scheme != tt.args.s {
				t.Errorf("NewConfig().scheme = %v, want %v", got.scheme, tt.args.s)
			}
			if got.host != tt.args.h {
				t.Errorf("NewConfig().host = %v, want %v", got.host, tt.args.h)
			}
			if got.port != tt.args.p {
				t.Errorf("NewConfig().port = %v, want %v", got.port, tt.args.p)
			}
			if got.logger == nil {
				t.Errorf("NewConfig().logger is nil, want non-nil logger")
			}
		})
	}
}
