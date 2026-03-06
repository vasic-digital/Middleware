# Course: HTTP Middleware in Go

## Module Overview

This course covers the `digital.vasic.middleware` module, teaching how to build, compose, and configure HTTP middleware using the standard `func(http.Handler) http.Handler` pattern. You will learn CORS, logging, recovery, request IDs, authentication, rate limiting, compression, and middleware chaining.

## Prerequisites

- Intermediate Go knowledge (interfaces, `net/http`, closures)
- Basic understanding of HTTP headers and middleware concepts
- Go 1.24+ installed

## Lessons

| # | Title | Duration |
|---|-------|----------|
| 1 | Middleware Fundamentals and Chaining | 40 min |
| 2 | Observability: Logging, Request IDs, and Recovery | 40 min |
| 3 | Security: CORS, Authentication, and Rate Limiting | 45 min |
| 4 | Performance: Brotli Compression, Caching, and Validation | 40 min |

## Source Files

- `pkg/chain/` -- Middleware chaining utility
- `pkg/cors/` -- CORS headers and preflight handling
- `pkg/logging/` -- Request logging
- `pkg/recovery/` -- Panic recovery
- `pkg/requestid/` -- Request ID propagation
- `pkg/auth/` -- JWT authentication middleware
- `pkg/ratelimit/` -- Token-bucket rate limiting
- `pkg/brotli/` -- Brotli compression
- `pkg/cache/` -- Cache-Control headers
- `pkg/validation/` -- Input validation and sanitization
- `pkg/altsvc/` -- Alt-Svc HTTP/3 advertisement
- `pkg/gin/` -- Gin framework adapter
