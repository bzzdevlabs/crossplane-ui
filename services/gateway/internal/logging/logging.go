// Package logging builds a slog.Logger honouring the service-wide conventions:
// snake_case keys, structured attributes only, static messages.
package logging

import (
	"log/slog"
	"os"
)

// New builds a slog.Logger from the configured level and format. Unknown
// values fall back to "info" and "json" so that logging is always available.
func New(level, format string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: lvl == slog.LevelDebug,
	}

	var handler slog.Handler
	switch format {
	case "text":
		handler = slog.NewTextHandler(os.Stderr, opts)
	default:
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}

	return slog.New(handler).With(slog.String("service", "gateway"))
}
