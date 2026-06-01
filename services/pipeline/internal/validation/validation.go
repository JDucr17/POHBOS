package validation

import (
	"errors"
	"fmt"
)

// Validator collects validation errors and joins them via errors.Join
type Validator struct {
	errs []error
}

// Check appends an error with the given message if cond is false
func (v *Validator) Check(cond bool, msg string) {
	if !cond {
		v.errs = append(v.errs, errors.New(msg))
	}
}

// Checkf appends a formatted error if cond is false. Mirrors fmt.Errorf
func (v *Validator) Checkf(cond bool, format string, args ...any) {
	if !cond {
		v.errs = append(v.errs, fmt.Errorf(format, args...))
	}
}

// Err returns the joined errors, or nil if none were collected
func (v Validator) Err() error {
	return errors.Join(v.errs...)
}
