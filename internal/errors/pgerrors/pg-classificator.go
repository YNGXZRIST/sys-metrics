package pgerrors

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// PGErrorClassification classifies PostgreSQL errors as retriable or not.
type PGErrorClassification int

const (
	// NonRetriable means the operation should not be retried.
	NonRetriable PGErrorClassification = iota

	// Retriable means the operation may be retried (e.g. network blips, deadlock).
	Retriable
)

// PostgresErrorClassifier decides whether a driver/server error is worth retrying.
type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify maps err to PGErrorClassification.
func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}

	// Convert to pgconn.PgError when possible
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return СlassifyPgError(pgErr)
	}

	// Default: non-retriable
	return NonRetriable
}

func СlassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	// PostgreSQL error codes: https://www.postgresql.org/docs/current/errcodes-appendix.html

	switch pgErr.Code {
	// Class 08 - connection errors
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable

	// Class 40 - transaction rollback
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retriable

	// Class 57 - operator intervention
	case pgerrcode.CannotConnectNow: // 57P03
		return Retriable
	}

	// More specific checks can be added using pgerrcode constants
	switch pgErr.Code {
	// Class 22 - data errors
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable

	// Class 23 - integrity constraint violations
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable

	// Class 42 - syntax and access errors
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}

	// Default: non-retriable
	return NonRetriable
}
