package logging

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew_NilOutputUsesDefault covers the cfg.Output == nil branch in New()
// where a non-nil Config is passed but with a nil Output writer. The middleware
// should default to os.Stdout without panicking.
func TestNew_NilOutputUsesDefault(t *testing.T) {
	cfg := &Config{
		Output:    nil, // triggers the nil Output branch
		SkipPaths: make(map[string]struct{}),
	}
	middleware := New(cfg)
	require.NotNil(t, middleware)

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestStatusRecorder_WriteHeaderIdempotent verifies that only the first call
// to WriteHeader captures the status code (covers the sr.written guard).
func TestStatusRecorder_WriteHeaderIdempotent(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Output:    &buf,
		SkipPaths: make(map[string]struct{}),
	}
	middleware := New(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // first call: captured
		w.WriteHeader(http.StatusOK)      // second call: ignored by recorder
	}))

	req := httptest.NewRequest(http.MethodGet, "/double-header", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "201")
	assert.Contains(t, logOutput, "/double-header")
}

// TestStatusRecorder_WriteBeforeWriteHeader covers the statusRecorder.Write
// path where Write is called before WriteHeader, setting the implicit 200.
func TestStatusRecorder_WriteBeforeWriteHeader(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Output:    &buf,
		SkipPaths: make(map[string]struct{}),
	}
	middleware := New(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data")) // implicit 200
	}))

	req := httptest.NewRequest(http.MethodGet, "/implicit", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "200")
	assert.Contains(t, logOutput, "/implicit")
}

// TestStatusRecorder_WriteAfterWriteHeader covers Write called after an
// explicit WriteHeader to verify the written flag prevents double-status.
func TestStatusRecorder_WriteAfterWriteHeader(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Output:    &buf,
		SkipPaths: make(map[string]struct{}),
	}
	middleware := New(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/write-after-header", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "404")
	assert.Contains(t, logOutput, "/write-after-header")
}

// TestNew_EmptySkipPaths covers using an empty SkipPaths map (not nil).
func TestNew_EmptySkipPaths(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Output:    &buf,
		SkipPaths: map[string]struct{}{},
	}
	middleware := New(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Contains(t, buf.String(), "/health")
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestNew_MultipleSkipPaths covers multiple entries in the SkipPaths map.
func TestNew_MultipleSkipPaths(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Output: &buf,
		SkipPaths: map[string]struct{}{
			"/health":  {},
			"/metrics": {},
			"/ready":   {},
		},
	}
	middleware := New(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// All skip paths should not produce log output
	for _, path := range []string{"/health", "/metrics", "/ready"} {
		buf.Reset()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Empty(t, buf.String(), "path %s should be skipped", path)
	}

	// Non-skipped path should log
	buf.Reset()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Contains(t, buf.String(), "/api/data")
}
