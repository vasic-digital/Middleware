package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildDirective_EmptyConfig covers the fallback case in buildDirective
// where no flags are set (no NoStore, no NoCache, no Public, no Private,
// no MaxAge, no MustRevalidate) and the result defaults to "no-store".
func TestBuildDirective_EmptyConfig(t *testing.T) {
	cfg := &Config{}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

// TestBuildDirective_NoCacheOnly covers the path where only NoCache is set
// without Public/Private/MaxAge, verifying partial directive building.
func TestBuildDirective_NoCacheOnly(t *testing.T) {
	cfg := &Config{NoCache: true}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
}

// TestBuildDirective_AllFlags covers buildDirective with all non-conflicting
// flags enabled simultaneously (NoCache, Public, MaxAge, MustRevalidate).
func TestBuildDirective_AllFlags(t *testing.T) {
	cfg := &Config{
		NoCache:        true,
		Public:         true,
		MaxAge:         300,
		MustRevalidate: true,
	}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cc := rec.Header().Get("Cache-Control")
	assert.Contains(t, cc, "no-cache")
	assert.Contains(t, cc, "public")
	assert.Contains(t, cc, "max-age=300")
	assert.Contains(t, cc, "must-revalidate")
}

// TestBuildDirective_PrivateOnly covers the path where only Private is set.
func TestBuildDirective_PrivateOnly(t *testing.T) {
	cfg := &Config{Private: true}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "private", rec.Header().Get("Cache-Control"))
}

// TestBuildDirective_MaxAgeOnly covers the path where only MaxAge is set
// (without Public/Private).
func TestBuildDirective_MaxAgeOnly(t *testing.T) {
	cfg := &Config{MaxAge: 600}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "max-age=600", rec.Header().Get("Cache-Control"))
}

// TestBuildDirective_MustRevalidateOnly covers the path where only
// MustRevalidate is set (without any other flags).
func TestBuildDirective_MustRevalidateOnly(t *testing.T) {
	cfg := &Config{MustRevalidate: true}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "must-revalidate", rec.Header().Get("Cache-Control"))
}

// TestNew_PathOverride_NoMatch covers the case where PathOverrides exist
// but none match the request path, so the default directive is used.
func TestNew_PathOverride_NoMatch(t *testing.T) {
	mw := New(&Config{
		Public: true,
		MaxAge: 3600,
		PathOverrides: map[string]string{
			"/static/": "no-store",
		},
	})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cc := rec.Header().Get("Cache-Control")
	assert.Contains(t, cc, "public")
	assert.Contains(t, cc, "max-age=3600")
}

// TestBuildDirective_NoStoreOverridesAll covers the NoStore flag's early
// return, ensuring all other flags are ignored when NoStore is true.
func TestBuildDirective_NoStoreOverridesAll(t *testing.T) {
	cfg := &Config{
		NoStore:        true,
		Public:         true,
		MaxAge:         3600,
		MustRevalidate: true,
	}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

// TestNew_MultiplePathOverrides covers handling of multiple path overrides.
func TestNew_MultiplePathOverrides(t *testing.T) {
	mw := New(&Config{
		NoStore: true,
		PathOverrides: map[string]string{
			"/static/":  "public, max-age=31536000",
			"/api/auth": "no-store, no-cache",
		},
	})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test /api/auth path override
	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, "no-store, no-cache", rec.Header().Get("Cache-Control"))
}
