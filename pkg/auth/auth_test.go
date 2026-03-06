package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockValidator struct {
	claims Claims
	err    error
}

func (m *mockValidator) ValidateToken(token string) (Claims, error) {
	return m.claims, m.err
}

func TestMiddleware_ValidToken(t *testing.T) {
	v := &mockValidator{claims: Claims{"sub": "user1", "role": "admin"}}
	mw := Middleware(v)

	var gotClaims Claims
	var ok bool
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims, ok = ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.True(t, ok)
	assert.Equal(t, "user1", gotClaims.GetString("sub"))
	assert.Equal(t, "admin", gotClaims.GetString("role"))
}

func TestMiddleware_MissingToken(t *testing.T) {
	v := &mockValidator{}
	mw := Middleware(v)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMiddleware_InvalidToken(t *testing.T) {
	v := &mockValidator{err: errors.New("token expired")}
	mw := Middleware(v)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMiddleware_SkipPaths(t *testing.T) {
	v := &mockValidator{err: errors.New("should not validate")}
	mw := Middleware(v, WithSkipPaths("/health", "/public"))

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

type customResponder struct {
	called bool
}

func (c *customResponder) RespondUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	c.called = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}

func TestMiddleware_CustomErrorResponder(t *testing.T) {
	resp := &customResponder{}
	v := &mockValidator{err: errors.New("bad")}
	mw := Middleware(v, WithErrorResponder(resp))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Authorization", "Bearer bad")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, resp.called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestClaims_Get(t *testing.T) {
	c := Claims{"key": 42, "str": "hello"}
	assert.Equal(t, 42, c.Get("key"))
	assert.Nil(t, c.Get("missing"))
}

func TestClaims_GetString(t *testing.T) {
	c := Claims{"str": "hello", "num": 42}
	assert.Equal(t, "hello", c.GetString("str"))
	assert.Equal(t, "", c.GetString("num"))
	assert.Equal(t, "", c.GetString("missing"))
}

func TestClaimsFromContext_NoClaims(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := ClaimsFromContext(req.Context())
	assert.False(t, ok)
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"valid bearer", "Bearer abc123", "abc123"},
		{"no prefix", "abc123", ""},
		{"empty", "", ""},
		{"wrong prefix", "Basic abc123", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			assert.Equal(t, tt.want, extractToken(req))
		})
	}
}
