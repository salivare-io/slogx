package slogx

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_UpdateConfig(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(
		WithOutput(buf),
		WithFormat(FormatJSON),
		WithLevel(slog.LevelInfo),
	)

	l.Info("message 1")
	assert.Contains(t, buf.String(), "message 1")
	buf.Reset()

	l.UpdateConfig(
		func(c *Config) {
			c.Level = slog.LevelError
		},
	)

	l.Info("invisible")
	assert.Empty(t, buf.String(), "INFO лог не должен был напечататься")

	l.Error("visible error")
	assert.Contains(t, buf.String(), "visible error")
}

func TestLogger_FormatSwitch(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(WithOutput(buf), WithFormat(FormatText))

	l.Info("text mode")
	assert.Contains(t, buf.String(), "level=INFO")
	assert.NotContains(t, buf.String(), "{") // Не JSON
	buf.Reset()

	l.UpdateConfig(
		func(c *Config) {
			c.Format = FormatJSON
		},
	)

	l.Info("json mode")
	assert.Contains(t, buf.String(), `"level":"INFO"`)
	assert.True(t, json.Valid(buf.Bytes()))
}
