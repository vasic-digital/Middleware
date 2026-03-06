# FAQ

## Is this module tied to a specific HTTP framework?

No. All middleware uses the standard `func(http.Handler) http.Handler` signature, making it compatible with any `net/http`-based router (standard library, Chi, Gorilla, etc.). A Gin adapter is provided in `pkg/gin` via the `Wrap()` function for Gin users.

## What happens when a nil config is passed to middleware constructors?

Every middleware accepts a nil config and falls back to `DefaultConfig()`. This means you can use `cors.New(nil)`, `logging.New(nil)`, or `recovery.New(nil)` for sensible defaults without explicitly constructing a config.

## How does the rate limiter handle multiple clients behind a NAT?

By default, the rate limiter keys on `r.RemoteAddr`, which includes the port. For clients behind a NAT or reverse proxy, inject a custom `KeyFunc` that extracts the real client identifier (e.g., `X-Forwarded-For` header, API key, or authenticated user ID).

## Does the Brotli middleware fall back to uncompressed responses?

Yes. The middleware only compresses responses when all of these conditions are met: the client sends `Accept-Encoding: br`, the response body exceeds `MinLength` (default 256 bytes), and the response Content-Type matches a compressible type. Otherwise the response is passed through uncompressed.

## How do I extract the request ID in downstream handlers?

Use `requestid.FromRequest(r)` or `requestid.FromContext(ctx)` to retrieve the request ID string from any handler or middleware further down the chain.
