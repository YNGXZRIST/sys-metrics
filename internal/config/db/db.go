// Package db wraps database/sql with retries for transient PostgreSQL errors.
package db

import (
	"context"
	"database/sql"
	"fmt"
	db "sys-metrics/internal/config/db/internal"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/errors/pgerrors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const maxRetries = 3

// generate:reset

// Tx wraps sql.Tx with retries on retriable PostgreSQL errors.
type Tx struct {
	*sql.Tx
}

// generate:reset

// DB extends the pgx connection pool with DSN from config.
type DB struct {
	*sql.DB
	*db.Config
}
type Config struct {
	DNS string
}

// NewCfg builds a connection config from server options (DSN).
func NewCfg(opt *Config) *db.Config {
	dns := opt.DNS
	return &db.Config{
		DNS: dns,
	}
}

// NewConn opens a pgx connection; an empty DSN is an error.
func NewConn(cfg *db.Config) (*DB, error) {
	if cfg == nil || cfg.DNS == "" {
		return nil, labelerrors.NewLabelError("DB", fmt.Errorf("database DSN is not set"))
	}
	dsn := cfg.DNS
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, labelerrors.NewLabelError("DB", fmt.Errorf("error connecting to database: %w", err))
	}
	return &DB{DB: conn, Config: cfg}, nil
}

func runWithRetry[T any](ctx context.Context, op func() (T, error)) (T, error) {
	var zero T
	var lastRes T
	var lastErr error
	sleepSeconds := 1
	classifier := pgerrors.NewPostgresErrorClassifier()
	for attempt := 0; attempt < maxRetries; attempt++ {
		lastRes, lastErr = op()
		if lastErr == nil {
			return lastRes, nil
		}
		if classifier.Classify(lastErr) == pgerrors.NonRetriable {
			return zero, labelerrors.NewLabelError("DB", pgerrors.NewPgError(lastErr))
		}
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(time.Duration(sleepSeconds) * time.Second):
			sleepSeconds += 2
		}
	}
	if lastErr != nil {
		lastErr = labelerrors.NewLabelError("DB", pgerrors.NewPgError(lastErr))
	}
	return lastRes, lastErr
}

// ExecContext runs a statement with context and retries.
func (D *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return runWithRetry(ctx, func() (sql.Result, error) {
		return D.DB.ExecContext(ctx, query, args...)
	})
}

// Exec runs a statement without context (background) with retries.
func (D *DB) Exec(query string, args ...any) (sql.Result, error) {
	return runWithRetry(context.Background(), func() (sql.Result, error) {
		return D.DB.Exec(query, args...)
	})
}

// Query runs a query without context with retries.
func (D *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return runWithRetry(context.Background(), func() (*sql.Rows, error) {
		return D.DB.Query(query, args...)
	})
}

// QueryContext runs a query with context and retries.
func (D *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return runWithRetry(ctx, func() (*sql.Rows, error) {
		return D.DB.QueryContext(ctx, query, args...)
	})
}

// BeginTx starts a transaction wrapped in Tx with retries on Exec/Query.
func (D *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	sqlTx, err := D.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: sqlTx}, nil
}

// ExecContext runs a statement inside the transaction with retries.
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return runWithRetry(ctx, func() (sql.Result, error) {
		return tx.Tx.ExecContext(ctx, query, args...)
	})
}

// Exec runs a statement in the transaction without context, with retries.
func (tx *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return runWithRetry(context.Background(), func() (sql.Result, error) {
		return tx.Tx.Exec(query, args...)
	})
}

// Query runs a query in the transaction without context, with retries.
func (tx *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return runWithRetry(context.Background(), func() (*sql.Rows, error) {
		return tx.Tx.Query(query, args...)
	})
}

// QueryContext runs a query in the transaction with context and retries.
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return runWithRetry(ctx, func() (*sql.Rows, error) {
		return tx.Tx.QueryContext(ctx, query, args...)
	})
}
