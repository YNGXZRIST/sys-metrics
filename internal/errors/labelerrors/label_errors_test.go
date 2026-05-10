package labelerrors

import (
	"errors"
	"strings"
	"testing"
)

func TestLabelError_Error(t *testing.T) {
	inner := errors.New("inner")
	err := NewLabelError("DB", inner)
	got := err.Error()
	if !strings.Contains(got, "[DB]") || !strings.Contains(got, "inner") {
		t.Fatalf("Error() = %q, want to contain [DB] and inner", got)
	}
}

func TestNewLabelError_Unwrap(t *testing.T) {
	inner := errors.New("wrapped")
	err := NewLabelError("X", inner)
	if !errors.Is(err, inner) {
		t.Fatalf("errors.Is(err, inner) = false, want true")
	}
	var le LabelError
	if !errors.As(err, &le) {
		t.Fatal("errors.As to LabelError failed")
	}
	if le.Label != "X" || !errors.Is(inner, le.Err) {
		t.Fatalf("LabelError = %#v, want Label X and inner", le)
	}
	if !errors.Is(inner, errors.Unwrap(err)) {
		t.Fatalf("Unwrap = %v, want %v", errors.Unwrap(err), inner)
	}
}
