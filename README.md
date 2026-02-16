# digital.vasic.middleware

Reusable HTTP middleware for Go's `net/http` standard library. Zero framework dependencies.

## Packages

| Package | Description |
|---|---|
| `pkg/cors` | Configurable Cross-Origin Resource Sharing middleware |
| `pkg/logging` | Request logging with method, path, status code, and duration |
| `pkg/recovery` | Panic recovery that catches panics and returns HTTP 500 |
| `pkg/requestid` | Request ID propagation via X-Request-ID header or UUID generation |
| `pkg/chain` | Middleware chaining utility for composing multiple middleware |

## Installation

```bash
go get digital.vasic.middleware
```

## Usage

```go
package main

import (
    "net/http"

    "digital.vasic.middleware/pkg/chain"
    "digital.vasic.middleware/pkg/cors"
    "digital.vasic.middleware/pkg/logging"
    "digital.vasic.middleware/pkg/recovery"
    "digital.vasic.middleware/pkg/requestid"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Compose middleware.
    middleware := chain.Chain(
        requestid.New(),
        logging.New(nil),
        recovery.New(nil),
        cors.New(cors.DefaultConfig()),
    )

    server := &http.Server{
        Addr:    ":8080",
        Handler: middleware(mux),
    }
    server.ListenAndServe()
}
```

### CORS

```go
cfg := &cors.Config{
    AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Authorization", "Content-Type"},
    ExposeHeaders:    []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           3600,
}
handler := cors.New(cfg)(myHandler)
```

### Logging

```go
cfg := &logging.Config{
    Output:    os.Stdout,
    SkipPaths: map[string]struct{}{"/health": {}},
}
handler := logging.New(cfg)(myHandler)
```

### Recovery

```go
cfg := &recovery.Config{
    Output:              os.Stderr,
    PrintStack:          true,
    ResponseBody:        []byte(`{"error":"internal_server_error"}`),
    ResponseContentType: "application/json",
}
handler := recovery.New(cfg)(myHandler)
```

### Request ID

```go
handler := requestid.New()(myHandler)

// Access the request ID inside a handler:
func myHandler(w http.ResponseWriter, r *http.Request) {
    id := requestid.FromRequest(r)
    w.Header().Set("X-Trace-ID", id)
}
```

### Chaining

```go
combined := chain.Chain(
    middleware1,
    middleware2,
    middleware3,
)
handler := combined(finalHandler)
// Equivalent to: middleware1(middleware2(middleware3(finalHandler)))
```

## Requirements

- Go 1.24.0+

## Testing

```bash
go test ./... -count=1
```

## License

See LICENSE file for details.
