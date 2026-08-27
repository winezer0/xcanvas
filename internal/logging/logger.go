package logging

import (
	"io"
	"log/slog"
)

// Nop returns a logger that discards all records.
func Nop() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Normalize returns logger or a no-op logger when logger is nil.
func Normalize(logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return Nop()
	}
	return logger
}
