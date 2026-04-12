// Package logger configures structured logging for the PhysicsCopilot server
// and provides security-audit helpers. Call Init() once at startup before
// any handler runs; use SecurityLog and HashIP for security-event logging.
package logger

import (
	"log/slog"
	"os"
)

// Init configures the global slog logger based on APP_ENV and LOG_LEVEL.
// JSON output is used when APP_ENV=production; human-readable text otherwise.
// LOG_LEVEL overrides the default level: debug, info, warn, error (case-insensitive).
func Init() {
	opts := &slog.HandlerOptions{Level: resolveLevel()}

	var handler slog.Handler
	if isProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}

// resolveLevel returns the slog.Level from the LOG_LEVEL env var, falling back
// to DEBUG in dev and INFO in production.
func resolveLevel() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "warn", "WARN":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	default:
		if isProduction() {
			return slog.LevelInfo
		}
		return slog.LevelDebug
	}
}

func isProduction() bool {
	env := os.Getenv("APP_ENV")
	return env == "production" || env == "prod"
}
