package agent

import (
	"time"

	"go.uber.org/zap"
)

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddr     string
	Logger         *zap.Logger
}

func NewConfig(pollInterval, reportInterval time.Duration, serverAddr string, logger *zap.Logger) *Config {
	return &Config{
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		ServerAddr:     serverAddr,
		Logger:         logger,
	}
}
