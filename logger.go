package slogx

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
)

// Logger is a wrapper around slog.Logger that supports atomic configuration updates.
// It allows changing log levels, formats, and sanitization rules at runtime without restarts.
type Logger struct {
	*slog.Logger
	cfgPtr *atomic.Pointer[Config]
}

// New creates a new Logger instance with the provided options.
// It initializes a DynamicHandler linked to an atomic configuration pointer.
func New(opts ...Option) *Logger {
	o := defaultOptions()
	for _, fn := range opts {
		if fn != nil {
			fn(o)
		}
	}

	// Initialize atomic pointer for thread-safe configuration management
	ptr := &atomic.Pointer[Config]{}
	ptr.Store(o.initialConfig)

	// Create a dynamic handler that reacts to config changes in real-time
	handler := &DynamicHandler{
		cfg: ptr,
	}

	return &Logger{
		Logger: slog.New(handler),
		cfgPtr: ptr,
	}
}

// With returns a derived slogx.Logger while preserving shared config.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		cfgPtr: l.cfgPtr,
	}
}

// WithGroup returns a grouped slogx.Logger that keeps the same config pointer.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: l.Logger.WithGroup(name),
		cfgPtr: l.cfgPtr,
	}
}

// UpdateConfig allows thread-safe, atomic updates to the logger's configuration.
// It uses a copy-on-write strategy by cloning the current config and applying the provided function.
func (l *Logger) UpdateConfig(fn func(*Config)) {
	oldCfg := l.cfgPtr.Load()
	newCfg := oldCfg.Clone()
	fn(newCfg)
	l.cfgPtr.Store(newCfg)
}

// SetLevel is a convenience method to quickly update the logging threshold.
func (l *Logger) SetLevel(lvl slog.Level) {
	l.UpdateConfig(
		func(c *Config) {
			c.Level = lvl
		},
	)
}

// TraceContext logs a message at the LevelTrace level with the given context.
func (l *Logger) TraceContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelTrace, msg, args...)
}

// FatalContext logs a message at the LevelFatal level and immediately terminates the process with exit code 1.
func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelFatal, msg, args...)
	os.Exit(1)
}

// SetupDefault initializes a new Logger and sets it as the global default logger for the slog package.
func SetupDefault(opts ...Option) {
	l := New(opts...)
	slog.SetDefault(l.Logger)
}

// Err is a helper function that creates a structured slog.Attr for an error.
// It ensures that error reporting remains consistent across all logs.
func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{Key: "error", Value: slog.StringValue("nil")}
	}
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
