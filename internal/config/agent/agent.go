package agent

import (
	"log"
	"time"
)

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddr     string
	Logger         *log.Logger
}

func NewConfig(pollInterval, reportInterval time.Duration, serverAddr string, logger *log.Logger) *Config {
	return &Config{
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		ServerAddr:     serverAddr,
		Logger:         logger,
	}
}
