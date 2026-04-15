// Package agent defines the metrics agent configuration: intervals, server address, authentication.
package agent

import (
	"sys-metrics/internal/authenticate"
	"time"

	"go.uber.org/zap"
)

// Config sets poll/report intervals, metrics ingest address, and parallelism limit.
type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddr     string
	Logger         *zap.Logger
	Authenticator  authenticate.Authenticator
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
