# Middleware Architecture

## Purpose

`digital.vasic.middleware` is a standalone Go module providing reusable HTTP middleware
built on `net/http`. It supplies CORS handling, request logging, panic recovery, request
ID propagation, Brotli compression, Alt-Svc advertisement, JWT authentication, cache
control, rate limiting, input validation, middleware chaining, and a Gin framework adapter.

All middleware uses the standard `func(http.Handler) http.Handler` signature, making every
component composable with any `net/http`-compatible router or framework.

## Package Overview

| Package | Responsibility |
|---------|---------------|
| `pkg/cors` | Configurable CORS headers and OPTIONS preflight handling |
| `pkg/logging` | Request logging with method, path, status code, and duration |
| `pkg/recovery` | Panic recovery that catches panics, logs stack traces, and returns HTTP 500 |
| `pkg/requestid` | Assigns or propagates a UUID-based X-Request-ID via header and context |
| `pkg/chain` | Composes multiple middleware into a single middleware function |
| `pkg/altsvc` | Sets the Alt-Svc response header to advertise HTTP/3 (QUIC) availability |
| `pkg/brotli` | Brotli response compression with content-type filtering and minimum size threshold |
| `pkg/gin` | Adapter that wraps standard `net/http` middleware for use with Gin |
| `pkg/auth` | JWT Bearer token authentication with injectable token validation strategy |
| `pkg/cache` | Sets Cache-Control response headers with path-based overrides |
| `pkg/ratelimit` | Token-bucket rate limiting per client key with 429 responses and Retry-After |
| `pkg/validation` | Request body size limits, content-type enforcement, and sanitization helpers |

## Design Patterns

| Package | Pattern | Rationale |
|---------|---------|-----------|
| `pkg/cors` | **Builder / Configuration Object** | `Config` struct with `DefaultConfig()` constructor separates policy from mechanism |
| `pkg/logging` | **Decorator** | `statusRecorder` wraps `http.ResponseWriter` to intercept the status code without altering behavior |
| `pkg/recovery` | **Decorator** | Uses `defer/recover` to transparently add panic safety to any handler |
| `pkg/requestid` | **Context Value** | Stores the request ID in `context.Context` for downstream access without coupling |
| `pkg/chain` | **Composite** | Folds an arbitrary list of middleware into a single middleware function |
| `pkg/altsvc` | **Decorator** | Injects a single response header with no behavioral change to the handler |
| `pkg/brotli` | **Proxy / Buffering Decorator** | `brotliWriter` buffers the response body, compresses it if eligible, then writes the final output |
| `pkg/gin` | **Adapter** | Converts the `func(http.Handler) http.Handler` signature to `gin.HandlerFunc` |
| `pkg/auth` | **Strategy + Functional Options** | `TokenValidator` interface is injected; behavior is customized via `Option` functions |
| `pkg/cache` | **Decorator** | Adds Cache-Control headers with path-prefix overrides |
| `pkg/ratelimit` | **Strategy** | `KeyFunc` is injected to determine the rate-limiting key per request |
| `pkg/validation` | **Chain of Responsibility** | Validation checks (body size, content-type) run sequentially before the handler |

## Dependency Diagram

```
                          +-----------+
                          |   chain   |
                          +-----+-----+
                                |
          +-------+-------+-----+-----+-------+-------+-------+
          |       |       |           |       |       |       |
       +--+--+ +--+--+ +-+-+    +----+----+ ++-+  +--+--+ +--+--+
       | cors| | log | |recv|   |requestid| |alt|  |brtli| |cache|
       +-----+ +-----+ +----+   +---------+ |svc|  +-----+ +-----+
                                             +---+
          +-------+-------+
          |       |       |
       +--+---+ +-+------++
       | auth | |ratelimit|
       +------+ +---------+
          |
       +--+--------+
       |validation  |
       +------------+

       +-----+
       | gin |  (adapter -- wraps any of the above for Gin)
       +-----+

  All packages are independent peers; `chain` composes them.
  `gin` adapts any middleware for Gin. No package depends on another.
```

## Key Interfaces

```go
// Every middleware follows this signature (no interface needed):
func New(cfg *Config) func(http.Handler) http.Handler

// pkg/auth -- consumers implement this to plug in their JWT library:
type TokenValidator interface {
    ValidateToken(tokenString string) (Claims, error)
}

// pkg/auth -- optional custom error responses:
type ErrorResponder interface {
    RespondUnauthorized(w http.ResponseWriter, r *http.Request, err error)
}

// pkg/ratelimit -- determines the rate-limiting key per request:
type KeyFunc func(r *http.Request) string
```

### Context Accessors

```go
// pkg/requestid
requestid.FromContext(ctx context.Context) string
requestid.FromRequest(r *http.Request) string

// pkg/auth
auth.ClaimsFromContext(ctx context.Context) (Claims, bool)
```

## Usage Example

```go
package main

import (
    "net/http"

    "digital.vasic.middleware/pkg/altsvc"
    "digital.vasic.middleware/pkg/auth"
    "digital.vasic.middleware/pkg/brotli"
    "digital.vasic.middleware/pkg/cache"
    "digital.vasic.middleware/pkg/chain"
    "digital.vasic.middleware/pkg/cors"
    "digital.vasic.middleware/pkg/logging"
    "digital.vasic.middleware/pkg/ratelimit"
    "digital.vasic.middleware/pkg/recovery"
    "digital.vasic.middleware/pkg/requestid"
    "digital.vasic.middleware/pkg/validation"
)

func main() {
    // Build the middleware stack.
    stack := chain.Chain(
        requestid.New(),
        logging.New(nil),
        recovery.New(nil),
        cors.New(cors.DefaultConfig()),
        altsvc.New(nil),
        brotli.New(brotli.DefaultConfig()),
        cache.New(cache.DefaultConfig()),
        ratelimit.New(nil),
        validation.New(nil),
        auth.Middleware(myValidator, auth.WithSkipPaths("/health", "/login")),
    )

    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("/api/data", dataHandler)

    http.ListenAndServe(":8080", stack(mux))
}
```

For Gin integration, wrap any middleware with the adapter:

```go
import ginAdapter "digital.vasic.middleware/pkg/gin"

router := gin.Default()
router.Use(ginAdapter.Wrap(cors.New(cors.DefaultConfig())))
router.Use(ginAdapter.Wrap(brotli.New(brotli.DefaultConfig())))
```
