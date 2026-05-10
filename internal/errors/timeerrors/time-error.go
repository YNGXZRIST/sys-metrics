// Package timeerrors annotates errors with a timestamp for log messages.
package timeerrors

import (
	"fmt"
	"time"
)

// generate:reset

// TimeError pairs an error with the time it was created.
type TimeError struct {
	Time time.Time
	Err  error
}

func (te *TimeError) Error() string {
	return fmt.Sprintf("%v %v", te.Time.Format("2006/01/02 15:04:05"), te.Err)
}

// NewTimeError records the current time and wraps err.
func NewTimeError(err error) error {
	return &TimeError{
		Time: time.Now(),
		Err:  err,
	}
}
func (te *TimeError) Unwrap() error { return te.Err }
