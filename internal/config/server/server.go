package server

import (
	"sys-metrics/internal/backup"

	"go.uber.org/zap"
)

type Config struct {
	scheme       string
	host         string
	port         string
	logger       *zap.Logger
	BackupConfig *backup.BackupConfig
}

const DefaultPort = "8080"
const DefaultHost = "localhost"
const SchemeHTTP = "http"

func NewConfig(s, h, p string, logger *zap.Logger, backupConfig *backup.BackupConfig) *Config {
	return &Config{s, h, p, logger, backupConfig}
}
func (c *Config) ServerAddr() string {
	return c.scheme + "://" + c.host + ":" + c.port
}
func (c *Config) InternalAddr() string {
	return c.host + ":" + c.port

}
