package pgerrors

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestNewPgError_nil(t *testing.T) {
	if NewPgError(nil) != nil {
		t.Fatal("NewPgError(nil) != nil")
	}
}

func TestNewPgError_generic(t *testing.T) {
	inner := errors.New("dial tcp")
	err := NewPgError(inner)
	if err == nil {
		t.Fatal("NewPgError returned nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "dial tcp") {
		t.Fatalf("Error() = %q", msg)
	}
	var pe *PgErrors
	if !errors.As(err, &pe) {
		t.Fatal("errors.As to *PgErrors failed")
	}
	if pe.Message != "dial tcp" || !errors.Is(inner, pe.Err) {
		t.Fatalf("PgErrors = %#v", pe)
	}
	if !errors.Is(inner, errors.Unwrap(err)) {
		t.Fatalf("Unwrap = %v, want %v", errors.Unwrap(err), inner)
	}
}

func TestNewPgError_pgconn(t *testing.T) {
	inner := &pgconn.PgError{
		Code:    pgerrcode.UniqueViolation,
		Message: "duplicate key",
	}
	err := NewPgError(inner)
	var pe *PgErrors
	if !errors.As(err, &pe) {
		t.Fatal("errors.As to *PgErrors failed")
	}
	if pe.Code != pgerrcode.UniqueViolation {
		t.Fatalf("Code = %q, want %q", pe.Code, pgerrcode.UniqueViolation)
	}
	if pe.Message != "duplicate key" {
		t.Fatalf("Message = %q", pe.Message)
	}
	if !errors.Is(inner, pe.Err) {
		t.Fatalf("Err not preserved")
	}
	msg := err.Error()
	if !strings.Contains(msg, pgerrcode.UniqueViolation) || !strings.Contains(msg, "duplicate key") {
		t.Fatalf("Error() = %q", msg)
	}
}

func TestPgErrors_Error_withoutWrappedErr(t *testing.T) {
	e := PgErrors{Code: "42P01", Message: "undefined_table"}
	got := e.Error()
	if got != "code=42P01 undefined_table" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestPgErrors_Error_withWrappedErr(t *testing.T) {
	inner := errors.New("cause")
	e := PgErrors{Code: "23505", Message: "dup", Err: inner}
	got := e.Error()
	if !strings.Contains(got, "23505") || !strings.Contains(got, "dup") || !strings.Contains(got, "cause") {
		t.Fatalf("Error() = %q", got)
	}
}
