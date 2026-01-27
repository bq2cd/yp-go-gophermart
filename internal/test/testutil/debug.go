package testutil

import (
	"log/slog"
	"os"
)

const (
	// LevelDebug matches [slog.LevelDebug].
	LevelDebug = slog.LevelDebug - iota
	// LevelTrace is one level below [slog.LevelDebug].
	LevelTrace
	// LevelTrace2 is two levels below [slog.LevelDebug].
	LevelTrace2
	// LevelTrace3 is three levels below [slog.LevelDebug].
	LevelTrace3
)

// MaybeEnableDebugLogging check for 'DEBUG' environment variable
// and if the variable is present and not empty,
// it will configure [slog.Logger] with the level specified in the variable.
// Value 'trace3' maps to [LevelTrace3] level.
// Value 'trace2' maps to [LevelTrace2] level.
// Value 'trace' maps to [LevelTrace] level.
// Any other value maps to [LevelDebug] level.
func MaybeEnableDebugLogging() {
	var wantLevel slog.Level

	switch os.Getenv("DEBUG") {
	case "":
		return
	case "trace3":
		wantLevel = LevelTrace3
	case "trace2":
		wantLevel = LevelTrace2
	case "trace":
		wantLevel = LevelTrace
	default:
		wantLevel = LevelDebug
	}

	handler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{ //nolint:exhaustruct
			Level: wantLevel,
		},
	)

	logger := slog.New(handler)

	slog.SetDefault(logger)
}
