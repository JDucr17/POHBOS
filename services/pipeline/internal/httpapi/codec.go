package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Validatable is the contract for types decoded from a request body. Decoded
// values are validated immediately by decodeAndValidate
type Validatable interface {
	Validate() error
}

// decodeAndValidate reads a JSON request body into T, enforces a max body
// size, rejects unknown fields, and runs T.Validate. Returns *Error on any
// failure with the appropriate status mapping
func DecodeAndValidate[T Validatable](w http.ResponseWriter, r *http.Request, maxBytes int64) (T, *Error) {
	var payload T

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		return payload, ErrorFor(r,fmt.Errorf("%w: %v", ErrInvalidPayload, err,))
	}

	if err := payload.Validate(); err != nil {
		return payload, ErrorFor(r,fmt.Errorf("%w: %v", ErrInvalidPayload, err))
	}

	return payload, nil
}

// encodeJSON writes input "v" as a JSON response with the given status code. Encoding
// happens to a buffer first so encoding errors can return a clean 500 before
// any bytes are committed to the response writer
func EncodeJSON(w http.ResponseWriter, status int, v any) *Error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		return &Error{
			Err:     fmt.Errorf("encode response: %v", err),
			Message: "internal error",
			Status:  http.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
	return nil
}
