package logger

import (
	"log/slog"
	"os"
)

// Init configures the global slog logger based on APP_ENV.
// JSON output is used when APP_ENV=production; human-readable text otherwise.
func Init() {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	var handler slog.Handler
	if isProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}

func isProduction() bool {
	env := os.Getenv("APP_ENV")
	return env == "production" || env == "prod"
}
