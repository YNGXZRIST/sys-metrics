package agent

import (
	"log"
	"time"
)

type Config struct {
	PoolInterval   time.Duration
	ReportInterval time.Duration
	ServerAddr     string
	Logger         *log.Logger
}

func NewConfig(poolInterval, reportInterval time.Duration, serverAddr string, logger *log.Logger) *Config {
	return &Config{
		PoolInterval:   poolInterval,
		ReportInterval: reportInterval,
		ServerAddr:     serverAddr,
		Logger:         logger,
	}
}
