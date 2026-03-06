package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig_NoStore(t *testing.T) {
	cfg := DefaultConfig()
	assert.True(t, cfg.NoStore)
}

func TestNew_DefaultNoStore(t *testing.T) {
	mw := New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

func TestNew_PublicMaxAge(t *testing.T) {
	mw := New(&Config{Public: true, MaxAge: 3600})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cc := rec.Header().Get("Cache-Control")
	assert.Contains(t, cc, "public")
	assert.Contains(t, cc, "max-age=3600")
}

func TestNew_PrivateMustRevalidate(t *testing.T) {
	mw := New(&Config{Private: true, MaxAge: 60, MustRevalidate: true})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cc := rec.Header().Get("Cache-Control")
	assert.Contains(t, cc, "private")
	assert.Contains(t, cc, "must-revalidate")
	assert.Contains(t, cc, "max-age=60")
}

func TestNew_NoCache(t *testing.T) {
	mw := New(&Config{NoCache: true})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Contains(t, rec.Header().Get("Cache-Control"), "no-cache")
}

func TestNew_PathOverrides(t *testing.T) {
	mw := New(&Config{
		NoStore: true,
		PathOverrides: map[string]string{
			"/static/": "public, max-age=31536000",
		},
	})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Static path uses override
	req := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, "public, max-age=31536000", rec.Header().Get("Cache-Control"))

	// API path uses default
	req2 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, "no-store", rec2.Header().Get("Cache-Control"))
}

func TestStaticAssetsConfig(t *testing.T) {
	cfg := StaticAssetsConfig()
	assert.True(t, cfg.Public)
	assert.Equal(t, 31536000, cfg.MaxAge)
}
