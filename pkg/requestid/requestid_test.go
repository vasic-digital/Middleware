package requestid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_GeneratesUUID(t *testing.T) {
	middleware := New()
	require.NotNil(t, middleware)

	var capturedID string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = FromRequest(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.NotEmpty(t, capturedID)
	assert.Equal(t, capturedID, rec.Header().Get(HeaderKey))
	// UUID v4 format: 8-4-4-4-12 hex characters.
	assert.Len(t, capturedID, 36)
}

func TestNew_ReusesExistingHeader(t *testing.T) {
	middleware := New()
	existingID := "my-custom-request-id-123"

	var capturedID string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = FromRequest(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderKey, existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, existingID, capturedID)
	assert.Equal(t, existingID, rec.Header().Get(HeaderKey))
}

func TestNew_UniquePerRequest(t *testing.T) {
	middleware := New()
	ids := make(map[string]struct{})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		id := rec.Header().Get(HeaderKey)
		assert.NotEmpty(t, id)
		_, exists := ids[id]
		assert.False(t, exists, "duplicate request ID: %s", id)
		ids[id] = struct{}{}
	}
}

func TestFromContext_EmptyWhenNotSet(t *testing.T) {
	ctx := context.Background()
	id := FromContext(ctx)
	assert.Empty(t, id)
}

func TestFromContext_ReturnsStoredID(t *testing.T) {
	middleware := New()

	var ctx context.Context
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := FromContext(ctx)
	assert.NotEmpty(t, id)
	assert.Equal(t, rec.Header().Get(HeaderKey), id)
}

func TestFromRequest_ReturnsID(t *testing.T) {
	middleware := New()

	var capturedReq *http.Request
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := FromRequest(capturedReq)
	assert.NotEmpty(t, id)
}

func TestNew_PassesThroughToHandler(t *testing.T) {
	middleware := New()

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/create", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "created", rec.Body.String())
}
