# Getting Started

## Installation

```bash
go get digital.vasic.middleware
```

## Basic Server with Middleware Chain

Compose multiple middleware into a single handler:

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
    mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    combined := chain.Chain(
        requestid.New(),
        logging.New(nil),
        recovery.New(nil),
        cors.New(cors.DefaultConfig()),
    )

    http.ListenAndServe(":8080", combined(mux))
}
```

## CORS Configuration

Configure allowed origins, methods, and headers:

```go
cfg := &cors.Config{
    AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           86400,
}
handler := cors.New(cfg)(myHandler)
```

## JWT Authentication

Inject a token validator to protect routes:

```go
import "digital.vasic.middleware/pkg/auth"

// Implement the TokenValidator interface
type myValidator struct{}

func (v *myValidator) ValidateToken(token string) (auth.Claims, error) {
    // Your JWT validation logic here
    return auth.Claims{"user_id": "123", "role": "admin"}, nil
}

middleware := auth.Middleware(
    &myValidator{},
    auth.WithSkipPaths("/health", "/login"),
)
handler := middleware(protectedHandler)
```

## Rate Limiting

Limit requests per client IP:

```go
import (
    "time"
    "digital.vasic.middleware/pkg/ratelimit"
)

cfg := &ratelimit.Config{
    Rate:   60,              // 60 requests
    Window: time.Minute,     // per minute
}
handler := ratelimit.New(cfg)(myHandler)
```

## Using with Gin

Wrap any standard middleware for use with the Gin framework:

```go
import (
    middlewaregin "digital.vasic.middleware/pkg/gin"
    "digital.vasic.middleware/pkg/logging"
    "github.com/gin-gonic/gin"
)

r := gin.New()
r.Use(middlewaregin.Wrap(logging.New(nil)))
```
