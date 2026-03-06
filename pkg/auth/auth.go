// Package auth provides generic JWT authentication middleware for net/http.
//
// It defines a TokenValidator interface that consumers implement with their
// JWT library of choice. The middleware extracts the Bearer token from the
// Authorization header and validates it via the injected validator.
//
// Design pattern: Strategy (token validation strategy is injected).
package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const claimsKey contextKey = "auth_claims"

// Claims holds validated token claims.
type Claims map[string]interface{}

// Get retrieves a claim value by key.
func (c Claims) Get(key string) interface{} {
	return c[key]
}

// GetString retrieves a claim value as string.
func (c Claims) GetString(key string) string {
	v, ok := c[key].(string)
	if !ok {
		return ""
	}
	return v
}

// TokenValidator validates JWT tokens and returns claims.
type TokenValidator interface {
	ValidateToken(tokenString string) (Claims, error)
}

// ErrorResponder writes authentication error responses.
type ErrorResponder interface {
	RespondUnauthorized(w http.ResponseWriter, r *http.Request, err error)
}

// defaultErrorResponder writes a plain-text 401 response.
type defaultErrorResponder struct{}

func (d defaultErrorResponder) RespondUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// Middleware returns HTTP middleware that validates JWT tokens.
// Requests without a valid token receive a 401 response.
// Valid claims are stored in the request context.
func Middleware(validator TokenValidator, opts ...Option) func(http.Handler) http.Handler {
	cfg := config{
		responder:  defaultErrorResponder{},
		skipPaths:  make(map[string]bool),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.skipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			token := extractToken(r)
			if token == "" {
				cfg.responder.RespondUnauthorized(w, r, nil)
				return
			}

			claims, err := validator.ValidateToken(token)
			if err != nil {
				cfg.responder.RespondUnauthorized(w, r, err)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext extracts claims from the request context.
func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(Claims)
	return claims, ok
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// Option configures the auth middleware.
type Option func(*config)

type config struct {
	responder ErrorResponder
	skipPaths map[string]bool
}

// WithErrorResponder sets a custom error responder.
func WithErrorResponder(r ErrorResponder) Option {
	return func(c *config) { c.responder = r }
}

// WithSkipPaths sets paths that bypass authentication.
func WithSkipPaths(paths ...string) Option {
	return func(c *config) {
		for _, p := range paths {
			c.skipPaths[p] = true
		}
	}
}
