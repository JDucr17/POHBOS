package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Sentinel errors describing categories of database failure or request
// lifecycle outcomes.
var (
	ErrConfig      = errors.New("postgres: invalid config")
	ErrUnavailable = errors.New("postgres: unavailable")
	ErrTransient   = errors.New("postgres: transient failure")
	ErrForbidden   = errors.New("postgres: access denied")

	ErrCanceled = errors.New("postgres: request canceled")
	ErrTimeout  = errors.New("postgres: request timeout")

	ErrUniqueViolation     = errors.New("postgres: unique violation")
	ErrForeignKeyViolation = errors.New("postgres: foreign key violation")
	ErrNotNullViolation    = errors.New("postgres: not null violation")
	ErrCheckViolation      = errors.New("postgres: check violation")
	ErrIntegrityViolation  = errors.New("postgres: integrity violation")

	ErrInvalidFormat = errors.New("postgres: invalid format")
	ErrQueryFailed   = errors.New("postgres: query failed")
)

// Error wraps a pgx error with a category sentinel. Exposes the underlying
// *pgconn.PgError for diagnostic detail and supports errors.Is against
// the sentinel via the Is method
type Error struct {
	Sentinel error
	PgError  *pgconn.PgError
	cause    error
}

func (e *Error) Error() string {
	if e.PgError != nil {
		return e.Sentinel.Error() + ": " + e.PgError.Message
	}
	return e.Sentinel.Error() + ": " + e.cause.Error()
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.Sentinel, target)
}

func (e *Error) Unwrap() error {
	return e.cause
}

// SQLSTATE codes per https://www.postgresql.org/docs/current/errcodes-appendix.html
var sqlstateMap = map[string]error{
	// Connection / resource issues — service genuinely unavailable
	"08000": ErrUnavailable, // connection_exception
	"08006": ErrUnavailable, // connection_failure
	"25006": ErrUnavailable, // read_only_sql_transaction
	"53100": ErrUnavailable, // disk_full
	"53200": ErrUnavailable, // out_of_memory
	"53300": ErrUnavailable, // too_many_connections
	"57P03": ErrUnavailable, // cannot_connect_now

	// Transient — application should retry
	"40001": ErrTransient, // serialization_failure
	"40P01": ErrTransient, // deadlock_detected
	"55P03": ErrTransient, // lock_not_available

	// Permission
	"42501": ErrForbidden, // insufficient_privilege

	// Data format / range
	"22001": ErrInvalidFormat, // string_data_right_truncation
	"22003": ErrInvalidFormat, // numeric_value_out_of_range
	"22007": ErrInvalidFormat, // invalid_datetime_format
	"22008": ErrInvalidFormat, // datetime_field_overflow
	"22P02": ErrInvalidFormat, // invalid_text_representation

	// Integrity violations
	"23502": ErrNotNullViolation,    // not_null_violation
	"23503": ErrForeignKeyViolation, // foreign_key_violation
	"23505": ErrUniqueViolation,     // unique_violation
	"23514": ErrCheckViolation,      // check_violation
}

// classify wraps a raw pgx error in a category-tagged *Error. Returns nil
// when err is nil. Distinguishes request-lifecycle outcomes (cancellation,
// timeout) from database failures (server-side errors with SQLSTATE) from
// connection failures (no SQLSTATE present)
func Classify(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) {
		return &Error{Sentinel: ErrCanceled, cause: err}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &Error{Sentinel: ErrTimeout, cause: err}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if sentinel, ok := sqlstateMap[pgErr.Code]; ok {
			return &Error{Sentinel: sentinel, PgError: pgErr, cause: err}
		}
		if len(pgErr.Code) >= 2 && pgErr.Code[:2] == "23" {
			return &Error{Sentinel: ErrIntegrityViolation, PgError: pgErr, cause: err}
		}
		return &Error{Sentinel: ErrQueryFailed, PgError: pgErr, cause: err}
	}

	return &Error{Sentinel: ErrUnavailable, cause: err}
}