package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

const StatusClientClosedRequest = 499

// ErrInvalidPayload is returned when a request body fails to decode or
// validate. Mapped to 400 by ErrorFor.
var ErrInvalidPayload = errors.New("invalid payload")

// failure value returned by a Handler.
type Error struct {
	Err     error
	Message string
	Status  int
}

// Handler is the handler contract for this package. Unlike http.Handler,
// implementations return *Error on failure instead of writing the response
// directly, letting ServeHTTP centralize error logging and status mapping.
//
// Handler still satisfies http.Handler because this type defines ServeHTTP.
// On success, handlers write directly to w and return nil; on failure, they
// must not write to w and should return a non-nil *Error.
type Handler func(w http.ResponseWriter, r *http.Request) *Error

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e := fn(w, r)
	if e == nil {
		return
	}

	logRequestError(r, e)
	writeError(w, e)
}

// ErrorFor maps application/storage errors to HTTP responses.
func ErrorFor(r *http.Request, err error) *Error {
	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.As(err, &maxBytesErr):
		return PayloadTooLarge(err, "request body too large")

	case errors.Is(err, ErrInvalidPayload):
		return BadRequestErr(err, err.Error())
	}

	if e := postgresError(r, err); e != nil {
		return e
	}


	return Internal(err)
}

func BadRequest(message string) *Error {
	return httpResp(errors.New(message), http.StatusBadRequest, message)
}

func BadRequestErr(err error, message string) *Error {
	return httpResp(err, http.StatusBadRequest, message)
}

func NotFound(message string) *Error {
	return httpResp(errors.New(message), http.StatusNotFound, message)
}

func Conflict(err error, message string) *Error {
	return httpResp(err, http.StatusConflict, message)
}

func PayloadTooLarge(err error, message string) *Error {
	return httpResp(err, http.StatusRequestEntityTooLarge, message)
}

func ServiceUnavailable(err error) *Error {
	return httpResp(err, http.StatusServiceUnavailable, "service unavailable")
}

func Internal(err error) *Error {
	return httpResp(err, http.StatusInternalServerError, "internal error")
}

// postgresError maps postgres sentinels to HTTP responses.
func postgresError(r *http.Request, err error) *Error {
	switch {
	case errors.Is(err, postgres.ErrCanceled):
		return httpResp(err, StatusClientClosedRequest, "request canceled")

	case errors.Is(err, postgres.ErrTimeout):
		return httpResp(err, http.StatusGatewayTimeout, "request timeout")

	case errors.Is(err, postgres.ErrUnavailable),
		errors.Is(err, postgres.ErrTransient):
		return ServiceUnavailable(err)

	case errors.Is(err, postgres.ErrForbidden):
		return httpResp(err, http.StatusForbidden, "access denied")

	case errors.Is(err, postgres.ErrUniqueViolation):
		return Conflict(err, "resource already exists")

	case errors.Is(err, postgres.ErrForeignKeyViolation):
		if r.Method == http.MethodDelete {
			return Conflict(err, "resource has dependencies")
		}
		return httpResp(err, http.StatusUnprocessableEntity, "referenced resource does not exist")

	case errors.Is(err, postgres.ErrNotNullViolation),
		errors.Is(err, postgres.ErrCheckViolation),
		errors.Is(err, postgres.ErrInvalidFormat),
		errors.Is(err, postgres.ErrIntegrityViolation):
		return BadRequestErr(err, "invalid request")

	case errors.Is(err, postgres.ErrQueryFailed):
		return Internal(err)
	}

	return nil
}

// httpResp is the constructor shorthand for *Error.
func httpResp(err error, status int, message string) *Error {
	return &Error{
		Err:     err,
		Message: message,
		Status:  status,
	}
}

func logRequestError(r *http.Request, e *Error) {
	if e.Status >= 500 {
		slog.ErrorContext(r.Context(), "request failed",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.Int("status", e.Status),
			slog.Any("error", e.Err),
		)
		return
	}

	slog.WarnContext(r.Context(), "request rejected",
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.Int("status", e.Status),
		slog.Any("error", e.Err),
	)
}

func writeError(w http.ResponseWriter, e *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)

	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: e.Message,
	})
}