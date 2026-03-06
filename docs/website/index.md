# Middleware Module

`digital.vasic.middleware` is a standalone Go module providing reusable HTTP middleware for `net/http`. It offers CORS handling, request logging, panic recovery, request ID generation, JWT authentication, rate limiting, Brotli compression, cache control, input validation, Alt-Svc headers, middleware chaining, and a Gin framework adapter -- all built on the Go standard library.

## Key Features

- **CORS** -- Configurable cross-origin resource sharing with preflight handling
- **Request logging** -- Method, path, status code, and duration for every request
- **Panic recovery** -- Catches panics in downstream handlers and returns HTTP 500
- **Request ID** -- Propagates or generates UUID v4 via X-Request-ID header
- **JWT authentication** -- Strategy-based token validation with skip paths and custom error responders
- **Rate limiting** -- Token-bucket rate limiter per client with configurable key extraction
- **Brotli compression** -- Transparent Brotli compression for compressible response types
- **Cache control** -- Configurable Cache-Control headers with path-based overrides
- **Input validation** -- Request body size limits and content-type enforcement
- **Alt-Svc** -- Advertises HTTP/3 (QUIC) availability via Alt-Svc header
- **Middleware chaining** -- Composes multiple middleware into a single function
- **Gin adapter** -- Wraps standard `net/http` middleware for use with Gin

## Package Overview

| Package | Purpose |
|---------|---------|
| `pkg/cors` | Cross-Origin Resource Sharing headers and preflight |
| `pkg/logging` | Request logging with status and duration |
| `pkg/recovery` | Panic recovery returning HTTP 500 |
| `pkg/requestid` | X-Request-ID propagation and generation |
| `pkg/auth` | JWT authentication middleware |
| `pkg/ratelimit` | Token-bucket rate limiting |
| `pkg/brotli` | Brotli compression |
| `pkg/cache` | Cache-Control header management |
| `pkg/validation` | Input validation and sanitization |
| `pkg/altsvc` | Alt-Svc header for HTTP/3 |
| `pkg/chain` | Middleware chaining utility |
| `pkg/gin` | Gin framework adapter |

## Installation

```bash
go get digital.vasic.middleware
```

Requires Go 1.24 or later.
