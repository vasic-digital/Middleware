# Lesson 1: Middleware Fundamentals and Chaining

## Learning Objectives

- Understand the `func(http.Handler) http.Handler` middleware signature
- Compose multiple middleware using `chain.Chain()`
- Adapt standard middleware for the Gin framework with `gin.Wrap()`

## Key Concepts

- **Middleware Signature**: Every middleware in this module returns `func(http.Handler) http.Handler`. The inner function wraps the next handler, executing logic before and/or after calling `next.ServeHTTP(w, r)`.
- **Chain of Responsibility**: `chain.Chain(a, b, c)` produces `a(b(c(handler)))`. The first middleware in the argument list is the outermost wrapper and executes first on the way in.
- **Gin Adapter**: `gin.Wrap(middleware)` converts any standard middleware into a `gin.HandlerFunc` by wrapping Gin's context writer and request into `http.Handler` calls.

## Code Walkthrough

### Source: `pkg/chain/chain.go`

The `Chain` function iterates middleware in reverse order, wrapping each around the final handler:

```go
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
    return func(final http.Handler) http.Handler {
        for i := len(middlewares) - 1; i >= 0; i-- {
            final = middlewares[i](final)
        }
        return final
    }
}
```

### Source: `pkg/gin/gin.go`

The `Wrap` function bridges `net/http` middleware into Gin by creating an `http.HandlerFunc` that calls `c.Next()`:

```go
func Wrap(middleware func(http.Handler) http.Handler) gin.HandlerFunc {
    return func(c *gin.Context) {
        next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Request = r
            c.Next()
        })
        middleware(next).ServeHTTP(c.Writer, c.Request)
    }
}
```

## Practice Exercise

1. Create a middleware stack with `requestid.New()`, `logging.New(nil)`, and `recovery.New(nil)` using `chain.Chain()`. Verify that each middleware executes in order by checking response headers and log output.
2. Write a custom timing middleware that adds an `X-Response-Time` header with the request duration in milliseconds. Integrate it into the chain.
3. Set up a Gin router and use `gin.Wrap()` to apply the logging and recovery middleware. Verify log output and panic handling.
