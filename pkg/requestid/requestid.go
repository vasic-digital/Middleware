// Package requestid provides middleware that assigns a unique identifier to
// every HTTP request. If the incoming request already carries an X-Request-ID
// header, that value is reused; otherwise a new UUID v4 is generated.
//
// The request ID is stored in the request context and set as a response header.
package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// HeaderKey is the HTTP header name used to transmit the request ID.
const HeaderKey = "X-Request-ID"

// contextKey is the unexported key type used to store the request ID in the
// request context.
type contextKey struct{}

// New creates a request-ID middleware. For every request it either reuses the
// existing X-Request-ID header value or generates a new UUID v4, stores it in
// the request context, and sets it as a response header.
func New() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(HeaderKey)
			if id == "" {
				id = uuid.New().String()
			}

			// Store in context.
			ctx := context.WithValue(r.Context(), contextKey{}, id)
			r = r.WithContext(ctx)

			// Set response header.
			w.Header().Set(HeaderKey, id)

			next.ServeHTTP(w, r)
		})
	}
}

// FromContext extracts the request ID from the given context. It returns an
// empty string if no request ID is present.
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(contextKey{}).(string); ok {
		return id
	}
	return ""
}

// FromRequest is a convenience wrapper that extracts the request ID from an
// *http.Request's context.
func FromRequest(r *http.Request) string {
	return FromContext(r.Context())
}
