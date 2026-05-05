// Package agent defines the metrics agent configuration: intervals, server address, authentication.
package agent

import (
	"sys-metrics/internal/authenticate"
	"time"

	"go.uber.org/zap"
)

// generate:reset

// Config sets poll/report intervals, metrics ingest address, and parallelism limit.
type Config struct {
	Authenticator  authenticate.Authenticator
	Logger         *zap.Logger
	ServerAddr     string
	PollInterval   time.Duration
	ReportInterval time.Duration
	RateLimit      int
}

// NewConfig constructs an agent Config.
func NewConfig(pollInterval, reportInterval time.Duration, serverAddr string, logger *zap.Logger, authenticator authenticate.Authenticator, rateLimit int) *Config {
	return &Config{
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		ServerAddr:     serverAddr,
		Logger:         logger,
		Authenticator:  authenticator,
		RateLimit:      rateLimit,
	}
}
