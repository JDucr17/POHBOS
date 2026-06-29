package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// RequiredEnv returns a trimmed environment variable or an error if missing.
func RequiredEnv(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

// EnvOrDefault returns a trimmed env var or fallback if it is empty.
func EnvOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

// SplitCSV returns trimmed, non-empty comma-separated values.
func SplitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}

	return values
}

func ParseDurationEnv(name string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}

	return value, nil
}

func ParseIntEnv(name string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}

	return value, nil
}

// ParseIntAtLeast reads an optional int env var and requires it to be at least min.
func ParseIntAtLeast(name string, fallback, min int) (int, error) {
	value, err := ParseIntEnv(name, fallback)
	if err != nil {
		return 0, err
	}
	if value < min {
		return 0, fmt.Errorf("%s must be at least %d, got %d", name, min, value)
	}
	return value, nil
}

// RequiredVars collects missing required env vars so loaders can report all
// config errors at once.
type RequiredVars struct {
	errs []error
}

// Get reads a required env var.
func (r *RequiredVars) Get(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		r.errs = append(r.errs, fmt.Errorf("%s is required", name))
	}
	return value
}

// CSV reads a required comma-separated env var.
func (r *RequiredVars) CSV(name string) []string {
	values := SplitCSV(os.Getenv(name))
	if len(values) == 0 {
		r.errs = append(r.errs, fmt.Errorf("%s is required", name))
	}
	return values
}

// Err returns all missing required env vars collected so far.
func (r *RequiredVars) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	return fmt.Errorf("missing required environment variables: %w", errors.Join(r.errs...))
}

// OptionalVars collects invalid optional env vars so loaders can report all
// config errors at once. Invalid values return their fallback.
type OptionalVars struct {
	errs []error
}

// IntAtLeast reads an optional int env var and requires it to be at least min.
// Missing vars use fallback, invalid vars are recorded and also use fallback.
func (o *OptionalVars) IntAtLeast(name string, fallback, min int) int {
	value, err := ParseIntAtLeast(name, fallback, min)
	if err != nil {
		o.errs = append(o.errs, err)
		return fallback
	}
	return value
}

// Err returns all optional env var errors collected so far.
func (o *OptionalVars) Err() error {
	if len(o.errs) == 0 {
		return nil
	}
	return fmt.Errorf("invalid optional environment variables: %w", errors.Join(o.errs...))
}