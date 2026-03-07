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

// TestBrotliWriter_WriteHeader_ExplicitStatusCode covers the WriteHeader method
// on the brotliWriter (0% coverage). When a handler explicitly calls WriteHeader
// before Write, the status code must be captured and forwarded.
func TestBrotliWriter_WriteHeader_ExplicitStatusCode(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(strings.Repeat("hello world ", 100)))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))

	reader := gobrotli.NewReader(w.Body)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("hello world ", 100), string(body))
}

// TestBrotliWriter_WriteHeader_NonCompressible covers WriteHeader when the
// response is too small or non-compressible, ensuring the explicit status code
// is forwarded without compression.
func TestBrotliWriter_WriteHeader_NonCompressible(t *testing.T) {
	cfg := br.DefaultConfig()
	cfg.MinLength = 10000

	handler := br.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("small body"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, "small body", w.Body.String())
}

// TestBrotliWriter_WriteHeader_204NoContent covers WriteHeader with 204 status
// and no body written at all.
func TestBrotliWriter_WriteHeader_204NoContent(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

// TestIsCompressible_EmptyContentType ensures that a response with no
// Content-Type header is not compressed (covers the contentType == "" branch).
func TestIsCompressible_EmptyContentType(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do NOT set Content-Type. The isCompressible check should return false.
		w.Write([]byte(strings.Repeat("data ", 500)))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	// Without a Content-Type, brotli should not compress
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

// TestBrotliWriter_WriteHeader_MultipleWrites covers multiple Write calls
// to verify buffering works correctly with explicit status.
func TestBrotliWriter_WriteHeader_MultipleWrites(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strings.Repeat(`{"key":"`, 50)))
		w.Write([]byte(strings.Repeat(`value"}`, 50)))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))

	reader := gobrotli.NewReader(w.Body)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)

	expected := strings.Repeat(`{"key":"`, 50) + strings.Repeat(`value"}`, 50)
	assert.Equal(t, expected, string(body))
}

// TestBrotliWriter_WriteHeader_304NotModified covers WriteHeader with 304.
func TestBrotliWriter_WriteHeader_304NotModified(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotModified, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

// TestBrotli_CompressibleTypesCaseInsensitive ensures content-type matching
// is case-insensitive (covers the strings.ToLower path in isCompressible).
func TestBrotli_CompressibleTypesCaseInsensitive(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "TEXT/PLAIN")
		w.Write([]byte(strings.Repeat("hello ", 200)))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
}

// TestBrotli_ApplicationWasm covers the application/wasm compressible type.
func TestBrotli_ApplicationWasm(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		w.Write([]byte(strings.Repeat("\x00\x61\x73\x6d", 200)))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
}

// TestBrotli_WriteHeaderOnly_NoWrite covers WriteHeader called with no Write.
// The statusCode should still be 0 if never explicitly written, but here
// we explicitly set it to test the writeHeader-only path in finish().
func TestBrotli_WriteHeaderOnly_StatusForwarded(t *testing.T) {
	handler := br.New(br.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		// No Write call - body is empty, below MinLength
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}
