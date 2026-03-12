# Slogx: Dynamic and Safe Logger

`Slogx` is an advanced wrapper over Go's standard `log/slog` package (Go 1.21+). It is designed for systems where you need to change log settings (level, format, masking) at runtime without restarting the application.

## Key Features

*   🚀 **Runtime configuration**: Change log level, format (JSON/Text), and output stream on the fly via atomic operations (`atomic.Pointer`).
*   🛡️ **Smart masking (redaction)**: Built-in masking for sensitive data (email, cards, phones) while keeping partial data for debugging.
*   🔍 **Context auto-injection**: Automatically extracts fields from `context.Context` (TraceID, RequestID, etc.) by configured keys.
*   🧩 **Zero-dependency**: Uses only the Go standard library.
*   ⚡ **High performance**: Optimized for high-load use (~500ns per operation).

---

## Quick Start

### Install

```bash
go get github.com/salivare-io/slogx@latest
```

## Basic Example

```go
package main

import (
	"context"
	"github.com/salivare-io/slogx"
	"log/slog"
)

func main() {
	// Initialize logger using builder options
	log := slogx.New(
		slogx.WithLevel(slogx.LevelTrace),
		slogx.WithContextKeys("trace_id", "request_id"),
		// Use MaskRules to group masking config
		slogx.WithMaskRules(slogx.NewMaskRules().
			Add("email", slogx.MaskEmail).
			Add("phone", slogx.MaskPhone),
		),
		// Remove fields
		slogx.WithRemoval(
			slogx.NewRemovalSet().
				Add("password").
				Add("token"),
		),
	)

	// Add data into context
	ctx := context.WithValue(context.Background(), "trace_id", "req-123")

	// Logger automatically extracts trace_id and masks email
	log.InfoContext(ctx, "User login attempt", slog.String("email", "admin@example.com"))
	// Output (Text): level=INFO msg="User login attempt" trace_id=req-123 email=ad***m@example.com
}

```

## HTTP middleware
Ready-to-use middleware for `net/http`, `chi`, `gin`, and `echo` with unified log format and options via `middleware/httplog`.

### net/http
```go
log := slogx.New(
	slogx.WithLevel(slogx.LevelTrace),
	slogx.WithFormat(slogx.FormatJSON),
)

mux := http.NewServeMux()
mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
	log := slogx.FromContext(r.Context())
	log.Info("handler called")
	w.WriteHeader(http.StatusNoContent)
})

handler := nethttp.LoggerContext(
	log,
	httplog.WithRequestIDFromHeader("X-Request-ID"),
)(
	nethttp.Logger(
		log,
		httplog.WithRequestIDFromHeader("X-Request-ID"),
	)(mux),
)

http.ListenAndServe(":8080", handler)
```

### chi
```go
r := chi.NewRouter()
r.Use(chimw.RequestID)
r.Use(slogchi.LoggerContext(log))
r.Use(slogchi.Logger(log))

r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
	slogx.FromContext(r.Context()).Info("handler called")
	w.WriteHeader(http.StatusNoContent)
})
```

### gin
```go
r := gin.New()
r.Use(sloggin.LoggerContext(log, httplog.WithRequestIDFromHeader("X-Request-ID")))
r.Use(sloggin.Logger(log, httplog.WithRequestIDFromHeader("X-Request-ID")))

r.GET("/ping", func(c *gin.Context) {
	slogx.FromContext(c.Request.Context()).Info("handler called")
	c.Status(http.StatusNoContent)
})
```

### echo
```go
e := echo.New()
e.Use(slogecho.LoggerContext(log, httplog.WithRequestIDFromHeader("X-Request-ID")))
e.Use(slogecho.Logger(log, httplog.WithRequestIDFromHeader("X-Request-ID")))

e.GET("/ping", func(c echo.Context) error {
	slogx.FromContext(c.Request().Context()).Info("handler called")
	return c.NoContent(http.StatusNoContent)
})
```

## Advanced configuration
#### Dynamic config updates
#### You can update any logger settings atomically:
```go
log.UpdateConfig(func(c *logger.Config) {
    c.Level = slog.LevelDebug   // Enable debug
    c.Format = logger.FormatJSON // Switch to JSON for ELK/Loki
})

```
### Data masking (Redaction)

#### The logger protects sensitive data. You map log keys to mask types:

| Mask Type | Result | Description |
| :--- | :--- | :--- |
| **MaskEmail** | `an***h@gmail.com` | First 2 chars + mask + last char |
| **MaskPhone** | `+7 9*******456` | Country prefix and last 3 digits |
| **MaskCard** | `4276 **** **** 0000` | First and last 4 digits |
| **MaskSecret** | `[SECRET]` | Full masking |

### Field removal and helpers
#### Ensure sensitive fields are never logged and use typed errors:
```go
// Field removal

logger := slogx.New(
    slogx.WithRemoval(
        slogx.NewRemovalSet().
            Add("password").
            Add("token"),
    ),
)

// Error helper
log.Error("db connection failed", logger.Err(err))

```
#### Test stub
```go
log := slogx.NewNop()
```
## Project architecture
* logger.go — constructor, Trace/Fatal methods, atomic config management.
* handler.go — DynamicHandler with WithAttrs/WithGroup and hot format switching.
* options.go — Config struct and all functional options.
* masker.go — data masking algorithms.
* context.go — context.Context helpers (Getter/Setter/TraceID).
* levels.go — custom TRACE/FATAL levels.
* middleware/ — HTTP middleware for popular frameworks.

## Performance (Apple M3 Max)

The logger is optimized with internal caching of static handler chains. This minimizes overhead even with complex dynamic configs.

| Scenario | Speed | Memory | Allocations |
| :--- | :--- | :--- | :--- |
| **Simple log (JSON)** | `499.2 ns/op` | `0 B/op` | **0 allocs/op** |
| **With masking** | `733.9 ns/op` | `112 B/op` | 5 allocs/op |
| **Config update (Atomic)** | `564.3 ns/op` | `640 B/op` | 10 allocs/op |
| **Level check (Skip)** | `3.24 ns/op` | `0 B/op` | 0 allocs/op |

## Comparison with other loggers

| Feature | Standard `slog` | Uber `zap` | `slogx` |
| :--- | :---: | :---: | :---: |
| **Dynamic level** | ⚠️ (via LevelVar) | ✅ | ✅ |
| **Format switching at runtime** | ❌ | ❌ | ✅ |
| **Data masking** | ❌ | ❌ | ✅ |
| **Context extraction** | ❌ | ❌ | ✅ |
| **Zero allocations** | ✅ | ✅ | ✅ (cache) |
| **Custom Trace/Fatal** | ❌ | ✅ | ✅ |

### Why `slogx`?
1. **Standard `slog`**: Very fast, but static. Changing output format or masking rules requires recreating the logger, which is not safe for concurrent use without extra wrappers.
2. **Uber `zap`**: Powerful and fast, but has no built-in context integration and does not support JSON/Text switching at runtime.
3. **`slogx`**: Combines `zap`-level speed with dynamic middleware flexibility. Ideal for microservices where you need to enable DEBUG or change masking rules in production without restarts.

## License
#### This project is distributed under the MIT license.
