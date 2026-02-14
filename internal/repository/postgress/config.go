package postgress

import (
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/repository/metricsiface"
)

type Config struct {
	handler     metricsiface.Handler
	conn        *db.DB
	initialized bool
}

func NewConfig(dbConn *db.DB) *Config {
	handler := NewHandler(dbConn)
	return &Config{conn: dbConn, handler: handler}
}

func (c *Config) NeedSync() bool {
	return c.initialized
}
