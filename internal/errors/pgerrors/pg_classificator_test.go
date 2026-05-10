package pgerrors

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresErrorClassifier_Classify_nil(t *testing.T) {
	c := NewPostgresErrorClassifier()
	if got := c.Classify(nil); got != NonRetriable {
		t.Fatalf("Classify(nil) = %v, want NonRetriable", got)
	}
}

func TestPostgresErrorClassifier_Classify_nonPg(t *testing.T) {
	c := NewPostgresErrorClassifier()
	if got := c.Classify(errors.New("plain")); got != NonRetriable {
		t.Fatalf("Classify(plain) = %v, want NonRetriable", got)
	}
}

func TestPostgresErrorClassifier_Classify_retriable(t *testing.T) {
	c := NewPostgresErrorClassifier()
	cases := []struct {
		code string
		name string
	}{
		{pgerrcode.ConnectionFailure, "connection failure"},
		{pgerrcode.SerializationFailure, "serialization"},
		{pgerrcode.DeadlockDetected, "deadlock"},
		{pgerrcode.CannotConnectNow, "cannot connect now"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := &pgconn.PgError{Code: tc.code, Message: tc.name}
			if got := c.Classify(err); got != Retriable {
				t.Fatalf("Classify(%s) = %v, want Retriable", tc.code, got)
			}
		})
	}
}

func TestPostgresErrorClassifier_Classify_nonRetriable(t *testing.T) {
	c := NewPostgresErrorClassifier()
	cases := []struct {
		code string
		name string
	}{
		{pgerrcode.UniqueViolation, "unique"},
		{pgerrcode.UndefinedTable, "undefined table"},
		{pgerrcode.SyntaxError, "syntax"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := &pgconn.PgError{Code: tc.code, Message: tc.name}
			if got := c.Classify(err); got != NonRetriable {
				t.Fatalf("Classify(%s) = %v, want NonRetriable", tc.code, got)
			}
		})
	}
}

func TestСlassifyPgError_unknownCode(t *testing.T) {
	err := &pgconn.PgError{Code: "99999", Message: "unknown"}
	if got := СlassifyPgError(err); got != NonRetriable {
		t.Fatalf("ClassifyPgError = %v, want NonRetriable", got)
	}
}

func TestСlassifyPgError_allMappedCodes(t *testing.T) {
	retriable := []string{
		pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure,
		pgerrcode.TransactionRollback,
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected,
		pgerrcode.CannotConnectNow,
	}
	nonRetriable := []string{
		pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException,
		pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation,
		pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction,
	}
	for _, code := range retriable {
		if got := СlassifyPgError(&pgconn.PgError{Code: code}); got != Retriable {
			t.Fatalf("code %q: got %v, want Retriable", code, got)
		}
	}
	for _, code := range nonRetriable {
		if got := СlassifyPgError(&pgconn.PgError{Code: code}); got != NonRetriable {
			t.Fatalf("code %q: got %v, want NonRetriable", code, got)
		}
	}
}
