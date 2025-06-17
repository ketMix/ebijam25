package log

import (
	"log/slog"

	"github.com/sytallax/prettylog"
)

// New creates a new slog.Logger with a pretty log handler.
func New(key, value string) *slog.Logger {
	ph := prettylog.NewHandler(&slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(ph).With(key, value)
}
