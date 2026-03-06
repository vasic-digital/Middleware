package validation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, int64(10*1024*1024), cfg.MaxBodySize)
	assert.Contains(t, cfg.RequireContentType, "application/json")
}

func TestNew_ValidRequest(t *testing.T) {
	mw := New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/data", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestNew_MissingContentType(t *testing.T) {
	mw := New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/data", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
}

func TestNew_WrongContentType(t *testing.T) {
	mw := New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	body := strings.NewReader(`name=test`)
	req := httptest.NewRequest(http.MethodPost, "/api/data", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
}

func TestNew_ContentTypeWithCharset(t *testing.T) {
	mw := New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/data", body)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestNew_GETSkipsValidation(t *testing.T) {
	mw := New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestNew_CustomContentTypes(t *testing.T) {
	mw := New(&Config{
		RequireContentType: []string{"application/json", "application/xml"},
		BodyMethods:        []string{"POST"},
	})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`<root/>`)
	req := httptest.NewRequest(http.MethodPost, "/api/data", body)
	req.Header.Set("Content-Type", "application/xml")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"  hello  ", "hello"},
		{"hel\x00lo", "hello"},
		{"\x00\x00", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, SanitizeString(tt.input))
	}
}

func TestSanitizeHeader(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal-value", "normal-value"},
		{"value\r\ninjected: header", "valueinjected: header"},
		{"value\ninjected", "valueinjected"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, SanitizeHeader(tt.input))
	}
}

func TestMatchesContentType(t *testing.T) {
	assert.True(t, matchesContentType("application/json", []string{"application/json"}))
	assert.True(t, matchesContentType("Application/JSON", []string{"application/json"}))
	assert.True(t, matchesContentType("application/json; charset=utf-8", []string{"application/json"}))
	assert.False(t, matchesContentType("text/plain", []string{"application/json"}))
}
