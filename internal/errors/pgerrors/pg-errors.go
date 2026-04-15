// Package pgerrors represents PostgreSQL errors in a unified form and wraps pgconn.PgError.
package pgerrors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// PgErrors holds SQLSTATE, server message, and the original error.
type PgErrors struct {
	Code    string // SQLSTATE, e.g. "23505", "42P01"
	Message string
	Err     error
}

func (e PgErrors) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%s %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%s %s", e.Code, e.Message)
}

func (e PgErrors) Unwrap() error { return e.Err }

// NewPgError extracts code and message from pgconn.PgError or stores a generic message.
func NewPgError(err error) error {
	if err == nil {
		return nil
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return &PgErrors{Code: pgErr.Code, Message: pgErr.Message, Err: err}
	}
	return &PgErrors{Message: err.Error(), Err: err}
}
