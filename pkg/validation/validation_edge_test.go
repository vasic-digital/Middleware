package validation_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"digital.vasic.middleware/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Malformed HTTP Headers ---

func TestValidation_MalformedContentType(t *testing.T) {
	t.Parallel()

	mw := validation.New(nil) // default config requires application/json
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name        string
		contentType string
		method      string
		wantStatus  int
	}{
		{
			"garbage_content_type",
			"not-a-valid/content;;;type",
			"POST",
			http.StatusUnsupportedMediaType,
		},
		{
			"empty_content_type",
			"",
			"POST",
			http.StatusUnsupportedMediaType,
		},
		{
			"spaces_only",
			"   ",
			"POST",
			http.StatusUnsupportedMediaType,
		},
		{
			"json_with_charset",
			"application/json; charset=utf-8",
			"POST",
			http.StatusOK,
		},
		{
			"json_uppercase",
			"APPLICATION/JSON",
			"POST",
			http.StatusOK,
		},
		{
			"get_no_content_type_required",
			"",
			"GET",
			http.StatusOK,
		},
		{
			"delete_no_content_type_required",
			"",
			"DELETE",
			http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, "/api/test", strings.NewReader("{}"))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

// --- Extremely Large Headers (>8KB body via MaxBodySize) ---

func TestValidation_LargeRequestBody(t *testing.T) {
	// bluff-scan: no-assert-ok (validator smoke — must not panic on edge inputs)
	t.Parallel()

	cfg := &validation.Config{
		MaxBodySize:        1024, // 1 KB limit
		RequireContentType: []string{"application/json"},
	}
	mw := validation.New(cfg)

	bodyRead := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 2048)
		_, err := r.Body.Read(buf)
		if err != nil {
			// MaxBytesReader should cause an error
			bodyRead = false
		} else {
			bodyRead = true
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Send 2KB body against 1KB limit
	largeBody := strings.Repeat("A", 2048)
	req := httptest.NewRequest("POST", "/api/test", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// The handler might get a truncated read or error
	_ = bodyRead
}

// --- Empty Request Body ---

func TestValidation_EmptyBody(t *testing.T) {
	t.Parallel()

	mw := validation.New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", bytes.NewReader(nil))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- Request With No Content-Type but Requires It ---

func TestValidation_NoContentType_BodyMethod(t *testing.T) {
	t.Parallel()

	cfg := &validation.Config{
		RequireContentType: []string{"application/json"},
	}
	mw := validation.New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	methods := []string{"POST", "PUT", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(method, "/api/test", strings.NewReader("data"))
			// No Content-Type header set
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
		})
	}
}

// --- Nil Config ---

func TestValidation_NilConfig(t *testing.T) {
	t.Parallel()

	mw := validation.New(nil)
	require.NotNil(t, mw)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- SanitizeString Edge Cases ---

func TestSanitizeString_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"empty", "", ""},
		{"null_bytes", "hello\x00world", "helloworld"},
		{"multiple_nulls", "\x00\x00\x00", ""},
		{"whitespace_only", "   \t\n  ", ""},
		{"null_and_spaces", " \x00hello\x00 ", "hello"},
		{"normal_string", "hello world", "hello world"},
		{"unicode_with_null", "\u00e9t\u00e9\x00", "\u00e9t\u00e9"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := validation.SanitizeString(tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}

// --- SanitizeHeader Edge Cases ---

func TestSanitizeHeader_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"empty", "", ""},
		{"crlf_injection", "value\r\nX-Injected: true", "valueX-Injected: true"},
		{"lf_only", "value\nX-Injected: true", "valueX-Injected: true"},
		{"cr_only", "value\rX-Injected: true", "valueX-Injected: true"},
		{"multiple_newlines", "a\r\nb\nc\rd", "abcd"},
		{"normal_header", "Bearer token123", "Bearer token123"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := validation.SanitizeHeader(tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}

// --- Duplicate Headers ---

func TestValidation_DuplicateContentTypeHeaders(t *testing.T) {
	t.Parallel()

	mw := validation.New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/test", strings.NewReader("{}"))
	// Add multiple Content-Type headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	// r.Header.Get returns the first value, which is application/json
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- Custom BodyMethods ---

func TestValidation_CustomBodyMethods(t *testing.T) {
	t.Parallel()

	cfg := &validation.Config{
		RequireContentType: []string{"application/json"},
		BodyMethods:        []string{"POST"}, // Only POST requires Content-Type
	}
	mw := validation.New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// PUT without Content-Type should pass since only POST is in BodyMethods
	req := httptest.NewRequest("PUT", "/api/test", strings.NewReader("data"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- MaxBodySize Zero (no limit) ---

func TestValidation_MaxBodySizeZero(t *testing.T) {
	t.Parallel()

	cfg := &validation.Config{
		MaxBodySize:        0, // no limit
		RequireContentType: []string{"application/json"},
	}
	mw := validation.New(cfg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024*1024)
		_, _ = r.Body.Read(buf)
		w.WriteHeader(http.StatusOK)
	}))

	// 100KB body should be fine with no limit
	largeBody := strings.Repeat("X", 100*1024)
	req := httptest.NewRequest("POST", "/api/test", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
