package db

import (
	"database/sql"
	"fmt"
	db "sys-metrics/internal/config/db/internal"
	"sys-metrics/internal/config/server"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
	*db.Config
}

func NewCfg(opt *server.Options) *db.Config {
	dns := opt.DNS
	return &db.Config{
		DNS: dns,
	}
}
func NewConn(cfg *db.Config) (*DB, error) {
	conn, err := sql.Open("pgx", cfg.DNS)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	return &DB{DB: conn, Config: cfg}, nil
}
