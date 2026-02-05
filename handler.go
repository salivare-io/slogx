package slogx

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// DynamicHandler is a middleware-style slog.Handler implementation that supports
// dynamic reconfiguration at runtime. It allows hot-swapping log level, output format,
// masking rules, attribute removal rules, and level name customization.
//
// To reduce per-log-call overhead, the handler caches the static handler chain
// (base handler + WithAttrs + WithGroup). Only context-derived attributes are applied
// dynamically on each call, because they depend on the incoming context.
//
// Attribute priority order remains:
//  1. Context-derived attributes (dynamic, highest priority)
//  2. Logger.With(...) attributes (cached)
//  3. Attributes added directly in the log call (slog.Record)
type DynamicHandler struct {
	cfg *atomic.Pointer[Config]

	attrs  []slog.Attr
	groups []string

	// cachedHandler stores the fully constructed static handler chain:
	//   baseHandler -> WithAttrs(attrs) -> WithGroup(groups)
	// It is rebuilt only when configuration, attrs, or groups change.
	cachedHandler atomic.Pointer[slog.Handler]

	// cachedConfigVersion tracks the last configuration pointer used to build the cache.
	// If cfg.Load() returns a different pointer, the cache is invalidated.
	cachedConfigVersion atomic.Pointer[Config]
}

// Enabled reports whether the record should be logged based on the current
// dynamic log level stored in the atomic configuration.
func (h *DynamicHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.cfg.Load().Level
}

// Handle processes a log record using a cached static handler chain.
// Only context-derived attributes are applied dynamically.
func (h *DynamicHandler) Handle(ctx context.Context, r slog.Record) error {
	cfg := h.cfg.Load()

	// Step 1: Collect context-derived attributes (highest priority)
	// This allows middleware to inject IDs into context that automatically appear in logs.
	var ctxAttrs []slog.Attr
	for _, key := range cfg.ContextKeys {
		if val := ctx.Value(key); val != nil {
			ctxAttrs = append(ctxAttrs, slog.Any(key, val))
		}
	}

	// Step 2: Get or rebuild the cached static handler chain
	base := h.getOrBuildCachedHandler(cfg)

	// Step 3: Apply context attributes (highest priority)
	if len(ctxAttrs) > 0 {
		base = base.WithAttrs(ctxAttrs)
	}

	// Step 4: Forward the record to the underlying handler
	return base.Handle(ctx, r)
}

// getOrBuildCachedHandler returns the cached handler chain if valid,
// otherwise rebuilds it and updates the cache.
//
// Cached chain includes:
//   - JSON/Text handler
//   - ReplaceAttr
//   - WithAttrs(attrs)
//   - WithGroup(groups)
//
// Context attributes are NOT cached.
func (h *DynamicHandler) getOrBuildCachedHandler(cfg *Config) slog.Handler {
	// Fast path: if config pointer matches and cache exists — return it
	if h.cachedConfigVersion.Load() == cfg {
		if cached := h.cachedHandler.Load(); cached != nil {
			return *cached
		}
	}

	// Slow path: rebuild the handler chain
	hOpts := &slog.HandlerOptions{
		Level:       cfg.Level,
		ReplaceAttr: h.getReplaceAttr(cfg),
	}

	var base slog.Handler
	if cfg.Format == FormatJSON {
		base = slog.NewJSONHandler(cfg.Output, hOpts)
	} else {
		base = slog.NewTextHandler(cfg.Output, hOpts)
	}

	// Apply WithAttrs (Logger.With(...) attributes)
	if len(h.attrs) > 0 {
		base = base.WithAttrs(h.attrs)
	}

	// Apply WithGroup (Logger.WithGroup(...) groups)
	for _, g := range h.groups {
		base = base.WithGroup(g)
	}

	// Store in cache
	h.cachedHandler.Store(&base)
	h.cachedConfigVersion.Store(cfg)

	return base
}

// WithAttrs returns a new DynamicHandler with additional attributes appended.
// Cache is invalidated because the handler chain changes.
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

// WithGroup returns a new DynamicHandler with an additional attribute group.
// Cache is invalidated because the handler chain changes.
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

// getReplaceAttr returns a transformation function used by slog.HandlerOptions.
// It performs:
//
//	Attribute removal (RemoveKeys)
//	Attribute masking (MaskKeys)
//	Level name customization (LevelNames)
func (h *DynamicHandler) getReplaceAttr(cfg *Config) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {

		// Attribute removal: Check if the key is in the removal set
		if _, shouldRemove := cfg.RemoveKeys[a.Key]; shouldRemove {
			return slog.Attr{}
		}

		// Attribute masking: Apply data redaction rules
		if mType, ok := cfg.MaskKeys[a.Key]; ok {
			a.Value = slog.AnyValue(cfg.Masker.Mask(a.Value.Any(), mType))
			return a
		}

		// Level name customization: Transform log level values to custom strings
		if a.Key == slog.LevelKey {
			if lvl, ok := a.Value.Any().(slog.Level); ok {
				a.Value = slog.StringValue(getLevelName(lvl, cfg.LevelNames))
			}
		}

		return a
	}
}
