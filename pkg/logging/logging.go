// Package logging provides HTTP request logging middleware that records
// the method, path, status code, and duration of every request.
package logging

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (sr *statusRecorder) WriteHeader(code int) {
	if !sr.written {
		sr.statusCode = code
		sr.written = true
	}
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if !sr.written {
		sr.statusCode = http.StatusOK
		sr.written = true
	}
	return sr.ResponseWriter.Write(b)
}

// Config holds configuration for the logging middleware.
type Config struct {
	// Output is the writer where log lines are sent. Defaults to os.Stdout.
	Output io.Writer

	// SkipPaths is a set of URL paths that should not be logged (e.g. health
	// check endpoints).
	SkipPaths map[string]struct{}
}

// DefaultConfig returns a default logging configuration that writes to stdout
// with no skipped paths.
func DefaultConfig() *Config {
	return &Config{
		Output:    os.Stdout,
		SkipPaths: make(map[string]struct{}),
	}
}

// New creates a request logging middleware. If cfg is nil the default
// configuration is used. Each request is logged in the format:
//
//	[HTTP] <method> <path> <status> <duration>
func New(cfg *Config) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	logger := log.New(cfg.Output, "", log.LstdFlags)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for configured paths.
			if _, skip := cfg.SkipPaths[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			logger.Printf("[HTTP] %s %s %d %s",
				r.Method,
				r.URL.Path,
				rec.statusCode,
				duration,
			)
		})
	}
}
