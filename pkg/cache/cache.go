// Package cache provides HTTP Cache-Control header middleware for net/http.
//
// It sets configurable cache directives on responses, supporting public/private
// caching, max-age, no-cache, no-store, and path-based overrides.
//
// Design pattern: Decorator (adds caching headers to response).
package cache

import (
	"fmt"
	"net/http"
	"strings"
)

// Config holds cache control configuration.
type Config struct {
	// Public marks responses as cacheable by shared caches.
	Public bool
	// Private marks responses as cacheable only by the browser.
	Private bool
	// MaxAge is the maximum age in seconds for cached responses.
	MaxAge int
	// NoCache forces revalidation before using cached response.
	NoCache bool
	// NoStore prevents any caching of the response.
	NoStore bool
	// MustRevalidate requires caches to revalidate stale responses.
	MustRevalidate bool
	// PathOverrides maps path prefixes to specific Cache-Control values.
	PathOverrides map[string]string
}

// DefaultConfig returns a default cache configuration with no-store
// (safe default: do not cache).
func DefaultConfig() *Config {
	return &Config{
		NoStore: true,
	}
}

// StaticAssetsConfig returns a configuration suitable for static assets
// with a 1-year max-age and public caching.
func StaticAssetsConfig() *Config {
	return &Config{
		Public: true,
		MaxAge: 31536000,
	}
}

// New creates a cache control middleware with the given configuration.
func New(cfg *Config) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	defaultDirective := buildDirective(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			directive := defaultDirective
			for prefix, override := range cfg.PathOverrides {
				if strings.HasPrefix(r.URL.Path, prefix) {
					directive = override
					break
				}
			}

			w.Header().Set("Cache-Control", directive)
			next.ServeHTTP(w, r)
		})
	}
}

func buildDirective(cfg *Config) string {
	var parts []string

	if cfg.NoStore {
		return "no-store"
	}
	if cfg.NoCache {
		parts = append(parts, "no-cache")
	}
	if cfg.Public {
		parts = append(parts, "public")
	}
	if cfg.Private {
		parts = append(parts, "private")
	}
	if cfg.MaxAge > 0 {
		parts = append(parts, fmt.Sprintf("max-age=%d", cfg.MaxAge))
	}
	if cfg.MustRevalidate {
		parts = append(parts, "must-revalidate")
	}

	if len(parts) == 0 {
		return "no-store"
	}
	return strings.Join(parts, ", ")
}
