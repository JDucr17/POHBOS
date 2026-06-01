package httpapi

import "strconv"

const (
	MaxEventBodyBytes = 32 * 1024

	DefaultListLimit = 100
	MaxListLimit     = 1000
)

func ParseLimit(raw string) (int, *Error) {
	if raw == "" {
		return DefaultListLimit, nil
	}

	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, BadRequest("limit must be an integer")
	}

	if limit <= 0 {
		return 0, BadRequest("limit must be greater than zero")
	}

	if limit > MaxListLimit {
		return 0, BadRequest("limit must be less than or equal to 1000")
	}

	return limit, nil
}