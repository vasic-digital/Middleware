// Package cors provides configurable Cross-Origin Resource Sharing middleware.
package cors

import (
	"net/http"
	"strconv"
	"strings"
)

// Config holds CORS configuration options.
type Config struct {
	// AllowOrigins is a list of origins that are allowed to make requests.
	// Use ["*"] to allow all origins.
	AllowOrigins []string

	// AllowMethods is a list of HTTP methods allowed for cross-origin requests.
	AllowMethods []string

	// AllowHeaders is a list of HTTP headers allowed in cross-origin requests.
	AllowHeaders []string

	// ExposeHeaders is a list of headers that browsers are allowed to access.
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials
	// such as cookies, HTTP authentication, or client-side SSL certificates.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached. A value of 0 means the header is not set.
	MaxAge int
}

// DefaultConfig returns a permissive default CORS configuration suitable for
// development environments.
func DefaultConfig() *Config {
	return &Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		MaxAge:       86400,
	}
}

// New creates a CORS middleware with the given configuration. The returned
// function wraps an http.Handler, injecting the appropriate CORS headers
// into every response and short-circuiting preflight OPTIONS requests with
// a 204 No Content status.
func New(cfg *Config) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	allowAllOrigins := len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*"
	originsSet := make(map[string]struct{}, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		originsSet[o] = struct{}{}
	}

	methods := strings.Join(cfg.AllowMethods, ", ")
	headers := strings.Join(cfg.AllowHeaders, ", ")
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ", ")
	maxAge := ""
	if cfg.MaxAge > 0 {
		maxAge = strconv.Itoa(cfg.MaxAge)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Determine if the origin is allowed.
			if origin != "" {
				if allowAllOrigins {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if _, ok := originsSet[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			}

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(cfg.AllowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", methods)
			}

			if len(cfg.AllowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}

			if exposeHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
			}

			if maxAge != "" {
				w.Header().Set("Access-Control-Max-Age", maxAge)
			}

			// Handle preflight requests.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
