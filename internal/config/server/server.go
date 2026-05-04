// Package server holds HTTP metrics server configuration and client-facing addresses.
package server

import (
	"sys-metrics/internal/repository/file"

	"go.uber.org/zap"
)

// Config stores scheme, host, port, logger, and optional file backup settings.
type Config struct {
	logger       *zap.Logger
	BackupConfig *file.Config
	scheme       string
	host         string
	port         string
}

// DefaultPort is the default HTTP listen port.
const DefaultPort = "8080"

// DefaultHost is the default listen host.
const DefaultHost = "localhost"

// SchemeHTTP is the default URL scheme for the server address.
const SchemeHTTP = "http"

// NewConfig builds a Config for listening and optional on-disk backup.
func NewConfig(s, h, p string, l *zap.Logger, b *file.Config) *Config {
	return &Config{l, b, s, h, p}
}

// ServerAddr returns the full server URL (scheme://host:port).
func (c *Config) ServerAddr() string {
	return c.scheme + "://" + c.host + ":" + c.port
}

// InternalAddr returns host:port without the scheme (for Listen).
func (c *Config) InternalAddr() string {
	return c.host + ":" + c.port

}
