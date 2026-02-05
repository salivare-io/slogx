package slogx

import (
	"context"
	"log/slog"
)

// ctxKey is an unexported type for context keys to avoid collisions with other packages.
type ctxKey struct{}

// FromContext extracts the Logger from the provided context.
// If no Logger is found, it returns a new Logger instance wrapping the slog.Default() logger.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return l
	}
	return &Logger{Logger: slog.Default()}
}

// ToContext injects the Logger into the provided context and returns the resulting context.
// This allows the logger to be passed down through the call stack using the context.
func ToContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}
