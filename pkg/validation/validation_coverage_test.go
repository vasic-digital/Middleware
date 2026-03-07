package validation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNew_EmptyBodyMethods covers the len(cfg.BodyMethods) == 0 branch in
// New() where a non-nil Config is provided but BodyMethods is empty. It
// should default to ["POST", "PUT", "PATCH"].
func TestNew_EmptyBodyMethods(t *testing.T) {
	cfg := &Config{
		MaxBodySize:        1024,
		RequireContentType: []string{"application/json"},
		BodyMethods:        []string{}, // triggers the empty BodyMethods branch
	}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// POST without Content-Type should be rejected (defaults kick in).
	body := strings.NewReader(`{"test":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/test", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)

	// POST with correct Content-Type should pass.
	body2 := strings.NewReader(`{"test":true}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/test", body2)
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

// TestNew_NilBodyMethodsSlice covers a nil BodyMethods slice, which should
// also trigger the default body methods.
func TestNew_NilBodyMethodsSlice(t *testing.T) {
	cfg := &Config{
		MaxBodySize:        1024,
		RequireContentType: []string{"application/json"},
		BodyMethods:        nil, // nil triggers len == 0
	}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// PUT without Content-Type should be rejected (PUT is in defaults).
	body := strings.NewReader(`{"test":true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/update", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
}

// TestNew_MaxBodySizeZero covers the MaxBodySize == 0 case where
// no body size limit should be enforced.
func TestNew_MaxBodySizeZero(t *testing.T) {
	cfg := &Config{
		MaxBodySize:        0, // no limit
		RequireContentType: []string{"application/json"},
		BodyMethods:        []string{"POST"},
	}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`{"test":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/test", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestNew_EmptyRequireContentType covers the case where RequireContentType
// is empty, meaning no content-type validation is performed.
func TestNew_EmptyRequireContentType(t *testing.T) {
	cfg := &Config{
		MaxBodySize:        1024,
		RequireContentType: []string{}, // no content type requirement
		BodyMethods:        []string{"POST"},
	}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// POST with no Content-Type should pass (no validation required).
	body := strings.NewReader("arbitrary data")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestNew_PATCHMethod covers the PATCH method being validated (part of defaults).
func TestNew_PATCHMethod(t *testing.T) {
	mw := New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// PATCH without Content-Type should be rejected.
	body := strings.NewReader(`{"field":"value"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/update", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)

	// PATCH with correct Content-Type should pass.
	body2 := strings.NewReader(`{"field":"value"}`)
	req2 := httptest.NewRequest(http.MethodPatch, "/api/update", body2)
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

// TestNew_DELETENotInBodyMethods covers the DELETE method being skipped by
// default body methods (DELETE is not in POST/PUT/PATCH).
func TestNew_DELETENotInBodyMethods(t *testing.T) {
	mw := New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// DELETE without Content-Type should pass (not a body method).
	req := httptest.NewRequest(http.MethodDelete, "/api/resource/1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestNew_CustomBodyMethods covers using a custom set of body methods.
func TestNew_CustomBodyMethods(t *testing.T) {
	cfg := &Config{
		RequireContentType: []string{"application/json"},
		BodyMethods:        []string{"DELETE"}, // custom: only DELETE requires content-type
	}
	mw := New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// POST without Content-Type should pass (not in custom body methods).
	body := strings.NewReader("data")
	req := httptest.NewRequest(http.MethodPost, "/api/data", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// DELETE without Content-Type should be rejected.
	req2 := httptest.NewRequest(http.MethodDelete, "/api/data", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusUnsupportedMediaType, rec2.Code)
}

// TestSanitizeString_OnlyNullBytes covers SanitizeString with only null bytes.
func TestSanitizeString_OnlyNullBytes(t *testing.T) {
	result := SanitizeString("\x00\x00\x00")
	assert.Equal(t, "", result)
}

// TestSanitizeString_MixedContent covers SanitizeString with mixed whitespace
// and null bytes.
func TestSanitizeString_MixedContent(t *testing.T) {
	result := SanitizeString("  \x00 hello \x00 world  ")
	assert.Equal(t, "hello  world", result)
}

// TestSanitizeHeader_OnlyNewlines covers SanitizeHeader with only CR/LF chars.
func TestSanitizeHeader_OnlyNewlines(t *testing.T) {
	result := SanitizeHeader("\r\n\r\n")
	assert.Equal(t, "", result)
}

// TestSanitizeHeader_CarriageReturnOnly covers SanitizeHeader with only CR.
func TestSanitizeHeader_CarriageReturnOnly(t *testing.T) {
	result := SanitizeHeader("value\rinjection")
	assert.Equal(t, "valueinjection", result)
}

// TestMatchesContentType_EmptyAllowed covers matchesContentType when the
// allowed list is empty.
func TestMatchesContentType_EmptyAllowed(t *testing.T) {
	assert.False(t, matchesContentType("application/json", []string{}))
}

// TestMatchesContentType_EmptyContentType covers matchesContentType when
// the content type is empty.
func TestMatchesContentType_EmptyContentType(t *testing.T) {
	assert.False(t, matchesContentType("", []string{"application/json"}))
}
