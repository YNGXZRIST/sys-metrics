// Package agent defines the metrics agent configuration: intervals, server address, authentication.
package agent

import (
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/secure"
	"time"

	"go.uber.org/zap"
)

// generate:reset

// Config sets poll/report intervals, metrics ingest address, and parallelism limit.
type Config struct {
	InitProperties
}

// InitProperties configuration of agent config
type InitProperties struct {
	Authenticator    authenticate.Authenticator
	Logger           *zap.Logger
	RequestEncryptor *secure.RequestEncryptor
	ServerAddr       string
	PollInterval     time.Duration
	ReportInterval   time.Duration
	RateLimit        int
	LocalIpV4        string
}

// NewConfig constructs an agent Config.
func NewConfig(prop InitProperties) *Config {
	return &Config{
		InitProperties: prop,
	}
}
