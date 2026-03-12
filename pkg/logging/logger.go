package logging

import (
	"log/slog"
	"os"
)

// New returns a JSON logger suitable for service processes and local tooling.
func New(serviceName string, env string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLevel(env),
	})

	return slog.New(handler).With(
		slog.String("service", serviceName),
		slog.String("env", env),
	)
}

func defaultLevel(env string) slog.Level {
	if env == "local" || env == "dev" {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}
