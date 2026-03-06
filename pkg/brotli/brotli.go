// Package brotli provides Brotli compression middleware for net/http.
package brotli

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

// Config holds Brotli middleware configuration.
type Config struct {
	Level             int
	MinLength         int
	CompressibleTypes []string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Level:     brotli.DefaultCompression,
		MinLength: 256,
		CompressibleTypes: []string{
			"text/",
			"application/json",
			"application/javascript",
			"application/xml",
			"application/xhtml+xml",
			"application/wasm",
			"image/svg+xml",
		},
	}
}

// New creates a Brotli compression middleware with the given config.
func New(cfg *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !acceptsBrotli(r) {
				next.ServeHTTP(w, r)
				return
			}

			bw := &brotliWriter{
				ResponseWriter: w,
				cfg:            cfg,
				buf:            &bytes.Buffer{},
			}
			next.ServeHTTP(bw, r)
			bw.finish()
		})
	}
}

type brotliWriter struct {
	http.ResponseWriter
	cfg         *Config
	buf         *bytes.Buffer
	wroteHeader bool
	statusCode  int
}

func (bw *brotliWriter) WriteHeader(code int) {
	bw.statusCode = code
	bw.wroteHeader = true
}

func (bw *brotliWriter) Write(b []byte) (int, error) {
	if !bw.wroteHeader {
		bw.statusCode = http.StatusOK
		bw.wroteHeader = true
	}
	return bw.buf.Write(b)
}

func (bw *brotliWriter) finish() {
	body := bw.buf.Bytes()
	ct := bw.ResponseWriter.Header().Get("Content-Type")

	if len(body) < bw.cfg.MinLength || !isCompressible(ct, bw.cfg.CompressibleTypes) {
		if bw.statusCode != 0 {
			bw.ResponseWriter.WriteHeader(bw.statusCode)
		}
		bw.ResponseWriter.Write(body)
		return
	}

	var compressed bytes.Buffer
	writer := brotli.NewWriterLevel(&compressed, bw.cfg.Level)
	writer.Write(body)
	writer.Close()

	bw.ResponseWriter.Header().Set("Content-Encoding", "br")
	bw.ResponseWriter.Header().Add("Vary", "Accept-Encoding")
	bw.ResponseWriter.Header().Del("Content-Length")
	if bw.statusCode != 0 {
		bw.ResponseWriter.WriteHeader(bw.statusCode)
	}
	bw.ResponseWriter.Write(compressed.Bytes())
}

func acceptsBrotli(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "br")
}

func isCompressible(contentType string, types []string) bool {
	if contentType == "" {
		return false
	}
	ct := strings.ToLower(contentType)
	for _, t := range types {
		if strings.HasPrefix(ct, t) {
			return true
		}
	}
	return false
}
