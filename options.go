package slogx

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Format defines the output format for the logger (Text or JSON).
type Format int

// MaskMap is a map that associates attribute keys with their corresponding masking strategies.
type MaskMap map[string]MaskType

// RemoveMap is a set of attribute keys that should be completely excluded from the logs.
type RemoveMap map[string]struct{}

const (
	// FormatText represents a human-readable key=value output format.
	FormatText Format = iota
	// FormatJSON represents a structured JSON output format.
	FormatJSON
)

// MaskRules provides a fluent interface to build and group masking configurations.
type MaskRules struct {
	rules MaskMap
}

// NewMaskRules creates a new instance of MaskRules builder.
func NewMaskRules() *MaskRules {
	return &MaskRules{rules: make(MaskMap)}
}

// Add associates a specific key with a MaskType.
func (r *MaskRules) Add(key string, mType MaskType) *MaskRules {
	r.rules[key] = mType
	return r
}

// Config represents the atomic logger configuration state.
// It includes level management, formatting, and data sanitization rules.
type Config struct {
	Level       slog.Level
	Format      Format
	Output      io.Writer
	MaskKeys    MaskMap
	RemoveKeys  RemoveMap
	LevelNames  LevelNames
	Masker      Masker
	ContextKeys []string
}

// Clone creates a deep copy of the Config to ensure thread-safe updates.
func (c *Config) Clone() *Config {
	newCfg := *c

	newCfg.MaskKeys = make(MaskMap, len(c.MaskKeys))
	for k, v := range c.MaskKeys {
		newCfg.MaskKeys[k] = v
	}

	newCfg.RemoveKeys = make(RemoveMap, len(c.RemoveKeys))
	for k, v := range c.RemoveKeys {
		newCfg.RemoveKeys[k] = v
	}

	newCfg.LevelNames = make(LevelNames, len(c.LevelNames))
	for k, v := range c.LevelNames {
		newCfg.LevelNames[k] = v
	}

	newCfg.ContextKeys = make([]string, len(c.ContextKeys))
	copy(newCfg.ContextKeys, c.ContextKeys)

	return &newCfg
}

// options internal helper for logger initialization.
type options struct {
	initialConfig *Config
}

// Option is a functional configuration parameter for logger initialization.
type Option func(*options)

// WithOutput sets the output destination (e.g., os.Stdout, a file, or a buffer).
func WithOutput(w io.Writer) Option {
	return func(o *options) {
		if w != nil {
			o.initialConfig.Output = w
		}
	}
}

// WithFormat sets the log output format (Text or JSON).
func WithFormat(f Format) Option {
	return func(o *options) {
		o.initialConfig.Format = f
	}
}

// ParseFormat converts a string representation to a Format type.
// It defaults to FormatText if the string is not recognized.
func ParseFormat(s string) Format {
	if strings.ToLower(s) == "json" {
		return FormatJSON
	}
	return FormatText
}

// WithLevel sets the initial logging threshold.
func WithLevel(l slog.Level) Option {
	return func(o *options) {
		o.initialConfig.Level = l
	}
}

// WithMaskKey associates a single attribute key with a MaskType.
func WithMaskKey(key string, mType MaskType) Option {
	return func(o *options) {
		o.initialConfig.MaskKeys[key] = mType
	}
}

// WithMaskKeys applies a batch of masking rules using a MaskMap.
func WithMaskKeys(keys MaskMap) Option {
	return func(o *options) {
		for k, v := range keys {
			o.initialConfig.MaskKeys[k] = v
		}
	}
}

// WithMaskRules applies masking rules using the MaskRules builder.
func WithMaskRules(r *MaskRules) Option {
	return func(o *options) {
		if r == nil {
			return
		}
		for k, v := range r.rules {
			o.initialConfig.MaskKeys[k] = v
		}
	}
}

// WithMasker replaces the default masking logic with a custom Masker implementation.
func WithMasker(m Masker) Option {
	return func(o *options) {
		if m != nil {
			o.initialConfig.Masker = m
		}
	}
}

// WithRemoval specifies attribute keys that should be omitted from the output.
func WithRemoval(keys ...string) Option {
	return func(o *options) {
		for _, k := range keys {
			o.initialConfig.RemoveKeys[k] = struct{}{}
		}
	}
}

// WithLevelNames customizes the string representation of log levels.
func WithLevelNames(m LevelNames) Option {
	return func(o *options) {
		for k, v := range m {
			o.initialConfig.LevelNames[k] = v
		}
	}
}

// WithContextKeys registers keys to be automatically extracted from context.Context and logged.
func WithContextKeys(keys ...string) Option {
	return func(o *options) {
		o.initialConfig.ContextKeys = append(o.initialConfig.ContextKeys, keys...)
	}
}

// defaultOptions provides the baseline configuration for a new logger.
func defaultOptions() *options {
	ln := make(LevelNames, len(defaultLevelNames))
	for k, v := range defaultLevelNames {
		ln[k.Level()] = v
	}

	return &options{
		initialConfig: &Config{
			Level:      LevelTrace,
			Format:     FormatText,
			Output:     os.Stdout,
			MaskKeys:   make(MaskMap),
			RemoveKeys: make(RemoveMap),
			LevelNames: ln,
			Masker:     &DefaultMasker{},
		},
	}
}
