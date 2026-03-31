# Architecture -- Middleware

## Purpose

Reusable HTTP middleware for Go's `net/http` standard library. Provides CORS handling, request logging, panic recovery, request ID generation/propagation, and middleware chaining -- all with zero framework dependencies.

## Structure

```
pkg/
  cors/        Configurable Cross-Origin Resource Sharing middleware with preflight handling
  logging/     Request logging with method, path, status code, and duration
  recovery/    Panic recovery that catches panics and returns HTTP 500
  requestid/   Request ID propagation via X-Request-ID header or UUID generation
  chain/       Middleware chaining utility for composing multiple middleware functions
```

## Key Components

- **`cors.New(cfg)`** -- CORS middleware with configurable origins, methods, headers, credentials, and max age
- **`logging.New(cfg)`** -- Request logger with skip paths and configurable output
- **`recovery.New(cfg)`** -- Panic recovery with optional stack trace printing and custom error response
- **`requestid.New()`** -- Generates UUID request IDs or propagates existing X-Request-ID headers; `FromRequest(r)` extracts the ID
- **`chain.Chain(middlewares...)`** -- Composes middleware in order: `Chain(a, b, c)(handler)` equals `a(b(c(handler)))`

## Data Flow

```
HTTP Request -> chain.Chain(requestid, logging, recovery, cors)(handler)
    |
    requestid: set X-Request-ID in context and response header
    |
    logging: record start time, wrap ResponseWriter, log on completion
    |
    recovery: defer recover(), return 500 on panic
    |
    cors: set CORS headers, handle OPTIONS preflight
    |
    handler: application logic
```

## Dependencies

- `github.com/google/uuid` -- UUID generation for request IDs
- `github.com/stretchr/testify` -- Test assertions

## Testing Strategy

Table-driven tests with `testify` using `httptest.NewRecorder`. Tests cover CORS header injection and preflight handling, request logging with skip paths, panic recovery with stack traces, request ID generation and propagation, and middleware chain composition order.
