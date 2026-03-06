# Lesson 3: Security: CORS, Authentication, and Rate Limiting

## Learning Objectives

- Configure CORS for development and production environments
- Implement JWT authentication with the Strategy pattern
- Apply per-client rate limiting with custom key extraction

## Key Concepts

- **CORS Preflight**: The CORS middleware short-circuits OPTIONS requests with 204 No Content after setting the appropriate headers. Origin matching supports wildcard (`*`) and explicit origin lists with O(1) set lookup.
- **Strategy Pattern (Auth)**: The `TokenValidator` interface decouples JWT validation from the middleware. Consumers inject their validation logic (any JWT library) without the middleware depending on it. Claims are stored in context and accessible via `ClaimsFromContext()`.
- **Token Bucket**: The rate limiter tracks per-key request counts within a time window. When the bucket is empty, requests receive 429 Too Many Requests with a `Retry-After` header indicating seconds until the window resets.

## Code Walkthrough

### Source: `pkg/cors/cors.go`

Pre-computes header strings at construction time for zero per-request allocation:

```go
originsSet := make(map[string]struct{}, len(cfg.AllowOrigins))
```

### Source: `pkg/auth/auth.go`

Extracts the Bearer token from the `Authorization` header, passes it to the injected `TokenValidator`, and stores validated claims in request context. Supports `WithSkipPaths()` and `WithErrorResponder()` options.

### Source: `pkg/ratelimit/ratelimit.go`

Each key gets a `bucket` with a token count and last-reset timestamp. The `KeyFunc` field allows custom key extraction (IP, API key, user ID).

## Practice Exercise

1. Configure CORS to allow only `https://app.example.com`. Test that requests from other origins do not receive CORS headers.
2. Implement a `TokenValidator` and apply the auth middleware with `/login` and `/health` as skip paths.
3. Set up a rate limiter with 10 requests per 10 seconds. Verify that excess requests receive 429 with a `Retry-After` header.
