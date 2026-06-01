package logging

import (
	"log/slog"
	"os"
	"strings"
)

const (
	envLogLevel  = "LOG_LEVEL"
	envLogFormat = "LOG_FORMAT"
)

// ConfigureFromEnv sets the default slog handler based on LOG_LEVEL and
// LOG_FORMAT env vars. Levels: debug, info (default), warn, error.
// Format: text (default) or json
func ConfigureFromEnv() {
	level := slog.LevelInfo

	switch strings.ToLower(strings.TrimSpace(os.Getenv(envLogLevel))) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if strings.ToLower(strings.TrimSpace(os.Getenv(envLogFormat))) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}