package timeerrors

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewTimeError_ErrorAndUnwrap(t *testing.T) {
	inner := errors.New("oops")
	before := time.Now()
	err := NewTimeError(inner)
	after := time.Now()
	if err == nil {
		t.Fatal("NewTimeError returned nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "oops") {
		t.Fatalf("Error() = %q, want substring oops", msg)
	}
	// format from Time.Format("2006/01/02 15:04:05")
	if !strings.Contains(msg, "/") {
		t.Fatalf("Error() = %q, want date-like prefix", msg)
	}
	var te *TimeError
	if !errors.As(err, &te) {
		t.Fatal("errors.As to *TimeError failed")
	}
	if te.Err != inner {
		t.Fatalf("TimeError.Err = %v, want %v", te.Err, inner)
	}
	if te.Time.Before(before) || te.Time.After(after) {
		t.Fatalf("TimeError.Time = %v, want between %v and %v", te.Time, before, after)
	}
	if !errors.Is(err, inner) {
		t.Fatalf("errors.Is(err, inner) = false, want true")
	}
}
