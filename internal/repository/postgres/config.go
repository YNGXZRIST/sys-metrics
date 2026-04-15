package postgres

import (
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/repository/metricsiface"
)

// Config binds the DB pool to metricsiface.Handler and sync readiness.
type Config struct {
	handler     metricsiface.Handler
	conn        *db.DB
	initialized bool
}

// NewConfig builds config with a SQL handler for the metrics table.
func NewConfig(dbConn *db.DB) *Config {
	handler := NewHandler(dbConn)
	return &Config{conn: dbConn, handler: handler}
}

// NeedSync reports whether storage is initialized (used by BackupStorage).
func (c *Config) NeedSync() bool {
	return c.initialized
}
