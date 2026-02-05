package slogx

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// DynamicHandler is a middleware handler that supports hot-swapping configuration
// at runtime. It manages dynamic log leveling, formatting, masking, and attribute removal.
type DynamicHandler struct {
	cfg    *atomic.Pointer[Config]
	attrs  []slog.Attr
	groups []string
}

// Enabled reports whether the handler handles records at the given level.
// It fetches the current threshold from the atomic configuration.
func (h *DynamicHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.cfg.Load().Level
}

// Handle processes the log record. It performs the following steps:
// 1. Extracts registered keys from the context.
// 2. Applies masking and removal rules via ReplaceAttr.
// 3. Selects the base handler (JSON or Text) based on current config.
// 4. Passes accumulated attributes and groups to the underlying handler.
func (h *DynamicHandler) Handle(ctx context.Context, r slog.Record) error {
	cfg := h.cfg.Load()

	// Automatically extract and log registered keys from context
	for _, key := range cfg.ContextKeys {
		if val := ctx.Value(key); val != nil {
			r.AddAttrs(slog.Any(key, val))
		}
	}

	// Prepare options for the underlying handler based on current config
	hOpts := &slog.HandlerOptions{
		Level:       cfg.Level,
		ReplaceAttr: h.getReplaceAttr(cfg),
	}

	// Initialize the base handler according to the current format
	var base slog.Handler
	if cfg.Format == FormatJSON {
		base = slog.NewJSONHandler(cfg.Output, hOpts)
	} else {
		base = slog.NewTextHandler(cfg.Output, hOpts)
	}

	// Forward accumulated attributes to the base handler
	if len(h.attrs) > 0 {
		base = base.WithAttrs(h.attrs)
	}

	// Forward accumulated groups to the base handler
	for _, g := range h.groups {
		base = base.WithGroup(g)
	}

	return base.Handle(ctx, r)
}

// WithAttrs returns a new DynamicHandler with the additional attributes.
// This ensures that Logger.With() calls work correctly.
func (h *DynamicHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &DynamicHandler{
		cfg:    h.cfg,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

// WithGroup returns a new DynamicHandler with the given group name.
// This ensures that Logger.WithGroup() calls work correctly.
func (h *DynamicHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &DynamicHandler{
		cfg:    h.cfg,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// getReplaceAttr returns a closure for attribute transformation.
// It handles field removal, data masking, and level name customization.
func (h *DynamicHandler) getReplaceAttr(cfg *Config) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		// Handle field removal
		if _, shouldRemove := cfg.RemoveKeys[a.Key]; shouldRemove {
			return slog.Attr{}
		}

		// Handle data masking
		if mType, ok := cfg.MaskKeys[a.Key]; ok {
			a.Value = slog.AnyValue(cfg.Masker.Mask(a.Value.Any(), mType))
			return a
		}

		// Handle level name customization
		if a.Key == slog.LevelKey {
			if lvl, ok := a.Value.Any().(slog.Level); ok {
				a.Value = slog.StringValue(getLevelName(lvl, cfg.LevelNames))
			}
		}

		return a
	}
}
