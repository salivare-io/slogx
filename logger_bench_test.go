package slogx

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

// BenchmarkSimpleInfo measures the normal speed of Log Info to Void (io.Discard)
func BenchmarkSimpleInfo(b *testing.B) {
	l := New(
		WithOutput(io.Discard),
		WithFormat(FormatJSON),
		WithLevel(slog.LevelInfo),
	)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.InfoContext(ctx, "test message", "key", "value")
	}
}

// BenchmarkMasking measures the speed of the masking operation
func BenchmarkMasking(b *testing.B) {
	l := New(
		WithOutput(io.Discard),
		WithFormat(FormatJSON),
		WithMaskKey("email", MaskEmail),
	)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.InfoContext(ctx, "user login", "email", "expensive_benchmark@gmail.com")
	}
}

// BenchmarkUpdateConfig Calculates the overhead costs for atomic replacements under load
func BenchmarkUpdateConfig(b *testing.B) {
	l := New(WithOutput(io.Discard))
	ctx := context.Background()

	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				// Read log
				l.InfoContext(ctx, "processing")
				// And change level in parallel
				l.SetLevel(slog.LevelDebug)
			}
		},
	)
}

// BenchmarkLevels Measure level filtering speed (when the log should not be written)
func BenchmarkLevels(b *testing.B) {
	l := New(
		WithOutput(io.Discard),
		WithLevel(slog.LevelError), // Level ERROR, and write INFO
	)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.InfoContext(ctx, "this should be skipped")
	}
}
