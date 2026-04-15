// Package labelerrors wraps errors with a short label for logging and errors.Unwrap.
package labelerrors

import "fmt"

// LabelError attaches a label to an underlying error.
type LabelError struct {
	Label string
	Err   error
}

func (e LabelError) Error() string {
	return fmt.Sprintf("[%s]: %s", e.Label, e.Err)
}

// NewLabelError returns a LabelError with the given label and inner error (supports errors.Unwrap).
func NewLabelError(label string, err error) error {
	return LabelError{Label: label, Err: err}
}
func (e LabelError) Unwrap() error { return e.Err }
