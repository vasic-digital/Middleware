// Package altsvc provides HTTP middleware that advertises HTTP/3 (QUIC) via Alt-Svc headers.
package altsvc

import (
	"fmt"
	"net/http"
)

// Config holds Alt-Svc middleware configuration.
type Config struct {
	Enabled bool
	H3Port  string
	MaxAge  int
}

// DefaultConfig returns a Config advertising HTTP/3 on port 443.
func DefaultConfig() *Config {
	return &Config{
		Enabled: true,
		H3Port:  "443",
		MaxAge:  86400,
	}
}

// New creates Alt-Svc middleware that sets the Alt-Svc response header.
func New(cfg *Config) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	header := fmt.Sprintf(`h3=":%s"; ma=%d`, cfg.H3Port, cfg.MaxAge)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Enabled {
				w.Header().Set("Alt-Svc", header)
			}
			next.ServeHTTP(w, r)
		})
	}
}
