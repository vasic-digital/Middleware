package brotli_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gobrotli "github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	br "digital.vasic.middleware/pkg/brotli"
)

func TestBrotliMiddleware_CompressesWhenAccepted(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strings.Repeat("hello world ", 100)))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
	assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding")

	reader := gobrotli.NewReader(w.Body)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("hello world ", 100), string(body))
}

func TestBrotliMiddleware_SkipsWhenNotAccepted(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("hello"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, "hello", w.Body.String())
}

func TestBrotliMiddleware_SkipsSmallResponses(t *testing.T) {
	cfg := br.DefaultConfig()
	cfg.MinLength = 1024
	handler := br.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("small"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, "small", w.Body.String())
}

func TestBrotliMiddleware_SkipsNonCompressibleTypes(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte(strings.Repeat("x", 2000)))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestBrotliMiddleware_CustomLevel(t *testing.T) {
	cfg := br.DefaultConfig()
	cfg.Level = gobrotli.BestSpeed
	handler := br.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strings.Repeat("hello world ", 100)))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
}
