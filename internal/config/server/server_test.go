package server

import (
	"log"
	"reflect"
	"testing"
)

func TestConfig_InternalAddr(t *testing.T) {
	type fields struct {
		scheme string
		host   string
		port   string
		logger *log.Logger
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
				logger: log.Default(),
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
		scheme string
		host   string
		port   string
		logger *log.Logger
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
				logger: log.Default(),
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
		s      string
		h      string
		p      string
		logger *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		{
			name: "default",
			args: args{
				s:      SchemeHTTP,
				h:      DefaultHost,
				p:      DefaultPort,
				logger: log.Default(),
			},
			want: &Config{
				scheme: SchemeHTTP,
				host:   DefaultHost,
				port:   DefaultPort,
				logger: log.Default(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(tt.args.s, tt.args.h, tt.args.p, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
