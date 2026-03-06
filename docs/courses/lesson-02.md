# Lesson 2: Observability: Logging, Request IDs, and Recovery

## Learning Objectives

- Log request method, path, status code, and duration with the logging middleware
- Propagate or generate request IDs for distributed tracing
- Recover from panics gracefully and return structured error responses

## Key Concepts

- **Status Recording**: The logging middleware wraps `http.ResponseWriter` with a `statusRecorder` that captures the written status code without modifying the response. This allows logging the final status after the handler completes.
- **Request ID Context**: `requestid.New()` stores the ID in the request context using an unexported key type to prevent collisions. Downstream handlers retrieve it via `requestid.FromContext(ctx)` or `requestid.FromRequest(r)`.
- **Panic Recovery**: `recovery.New()` uses `defer/recover` to catch panics. It logs the error (with optional stack trace via `runtime/debug.Stack()`) and writes an HTTP 500 response with a configurable body and content type.

## Code Walkthrough

### Source: `pkg/logging/logging.go`

The logging middleware skips configured paths, records start time, wraps the response writer, calls the next handler, then logs the result:

```go
logger.Printf("[HTTP] %s %s %d %s", r.Method, r.URL.Path, rec.statusCode, duration)
```

### Source: `pkg/requestid/requestid.go`

The middleware checks for an existing `X-Request-ID` header. If absent, it generates a UUID v4 via `uuid.New().String()`. The ID is stored in context and set as a response header.

### Source: `pkg/recovery/recovery.go`

On panic, the middleware logs the recovered value and optionally the full stack trace. The response body and content type are configurable via `Config.ResponseBody` and `Config.ResponseContentType`.

## Practice Exercise

1. Configure the logging middleware to skip `/health` and `/metrics` paths. Verify that requests to those paths produce no log output.
2. Write a handler that reads the request ID from context via `requestid.FromRequest(r)` and includes it in the JSON response body.
3. Configure the recovery middleware with a JSON error body and `application/json` content type. Trigger a panic and verify the JSON response.
