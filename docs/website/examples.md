# Examples

## Production Middleware Stack

A complete middleware stack for a production API server with compression, security, and observability:

```go
package main

import (
    "net/http"
    "time"

    "digital.vasic.middleware/pkg/altsvc"
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
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", handleData)

    stack := chain.Chain(
        recovery.New(nil),
        requestid.New(),
        logging.New(&logging.Config{
            SkipPaths: map[string]struct{}{"/health": {}},
        }),
        cors.New(&cors.Config{
            AllowOrigins: []string{"https://app.example.com"},
            AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
            AllowHeaders: []string{"Authorization", "Content-Type"},
        }),
        altsvc.New(&altsvc.Config{Enabled: true, H3Port: "443", MaxAge: 86400}),
        brotli.New(brotli.DefaultConfig()),
        ratelimit.New(&ratelimit.Config{Rate: 100, Window: time.Minute}),
        validation.New(&validation.Config{
            MaxBodySize:        10 * 1024 * 1024,
            RequireContentType: []string{"application/json"},
        }),
        cache.New(&cache.Config{
            NoStore: true,
            PathOverrides: map[string]string{
                "/static/": "public, max-age=31536000",
            },
        }),
    )

    http.ListenAndServe(":8080", stack(mux))
}
```

## Custom Rate Limit Key Extraction

Rate limit by API key from query parameter instead of IP:

```go
import (
    "net/http"
    "time"
    "digital.vasic.middleware/pkg/ratelimit"
)

cfg := &ratelimit.Config{
    Rate:   1000,
    Window: time.Hour,
    KeyFunc: func(r *http.Request) string {
        key := r.URL.Query().Get("api_key")
        if key == "" {
            return r.RemoteAddr
        }
        return key
    },
}
handler := ratelimit.New(cfg)(myHandler)
```

## Input Sanitization

Use the validation package helpers to sanitize user input:

```go
import "digital.vasic.middleware/pkg/validation"

func handleInput(w http.ResponseWriter, r *http.Request) {
    name := validation.SanitizeString(r.URL.Query().Get("name"))
    headerVal := validation.SanitizeHeader(r.Header.Get("X-Custom"))

    // name has null bytes removed and whitespace trimmed
    // headerVal has newlines removed to prevent header injection
}
```
