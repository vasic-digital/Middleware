// Package recovery provides panic recovery middleware for net/http handlers.
// When a downstream handler panics, the middleware recovers, logs the error,
// and returns an HTTP 500 Internal Server Error response.
package recovery

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
)

// Config holds configuration for the recovery middleware.
type Config struct {
	// Output is the writer where panic details are logged. Defaults to
	// os.Stderr.
	Output io.Writer

	// PrintStack controls whether the full stack trace is included in the
	// log output. Defaults to true.
	PrintStack bool

	// ResponseBody is the body returned to the client on panic. If nil a
	// default plain-text message is used.
	ResponseBody []byte

	// ResponseContentType is the Content-Type header for the error response.
	// Defaults to "text/plain; charset=utf-8".
	ResponseContentType string
}

// DefaultConfig returns a default recovery configuration that logs to stderr
// with stack traces enabled.
func DefaultConfig() *Config {
	return &Config{
		Output:              os.Stderr,
		PrintStack:          true,
		ResponseBody:        []byte("Internal Server Error\n"),
		ResponseContentType: "text/plain; charset=utf-8",
	}
}

// New creates a panic recovery middleware. If cfg is nil the default
// configuration is used.
func New(cfg *Config) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	if cfg.ResponseBody == nil {
		cfg.ResponseBody = []byte("Internal Server Error\n")
	}
	if cfg.ResponseContentType == "" {
		cfg.ResponseContentType = "text/plain; charset=utf-8"
	}

	logger := log.New(cfg.Output, "", log.LstdFlags)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if cfg.PrintStack {
						logger.Printf("[RECOVERY] panic recovered: %v\n%s", err, debug.Stack())
					} else {
						logger.Printf("[RECOVERY] panic recovered: %v", err)
					}

					w.Header().Set("Content-Type", cfg.ResponseContentType)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, string(cfg.ResponseBody))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
