package app

import (
	"log/slog"
	"os"
)

const (
	logLevelDefault = slog.LevelInfo
)

func setupLogger(debug bool) {
	level := logLevelDefault
	if debug {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{ //nolint:exhaustruct
		Level: level,
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)
}
