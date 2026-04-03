package recovery_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"digital.vasic.middleware/pkg/chain"
	"digital.vasic.middleware/pkg/cors"
	"digital.vasic.middleware/pkg/recovery"
	"digital.vasic.middleware/pkg/requestid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Recovery from various panic types ---

func TestRecovery_PanicWithString(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	mw := recovery.New(&recovery.Config{
		Output:     &buf,
		PrintStack: false,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("string panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, buf.String(), "string panic")
}

func TestRecovery_PanicWithError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	mw := recovery.New(&recovery.Config{
		Output:     &buf,
		PrintStack: false,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(io.ErrUnexpectedEOF)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestRecovery_PanicWithInt(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	mw := recovery.New(&recovery.Config{
		Output:     &buf,
		PrintStack: false,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(42)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestRecovery_PanicWithNil(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	mw := recovery.New(&recovery.Config{
		Output:     &buf,
		PrintStack: false,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(nil)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// panic(nil) in Go 1.21+ panics with *runtime.PanicNilError
	// so recovery middleware should catch it
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- No Panic ---

func TestRecovery_NoPanic(t *testing.T) {
	t.Parallel()

	mw := recovery.New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

// --- Nil Config ---

func TestRecovery_NilConfig(t *testing.T) {
	t.Parallel()

	mw := recovery.New(nil)
	require.NotNil(t, mw)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- Custom Response Body ---

func TestRecovery_CustomResponseBody(t *testing.T) {
	t.Parallel()

	mw := recovery.New(&recovery.Config{
		Output:              io.Discard,
		PrintStack:          false,
		ResponseBody:        []byte(`{"error":"internal"}`),
		ResponseContentType: "application/json",
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("crash")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), `"error":"internal"`)
}

// --- PrintStack true vs false ---

func TestRecovery_PrintStackTrue(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	mw := recovery.New(&recovery.Config{
		Output:     &buf,
		PrintStack: true,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("stack test")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	// Stack trace should include goroutine info
	assert.Contains(t, buf.String(), "goroutine")
}

func TestRecovery_PrintStackFalse(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	mw := recovery.New(&recovery.Config{
		Output:     &buf,
		PrintStack: false,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("no stack test")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, buf.String(), "no stack test")
	// Should NOT contain goroutine stack trace
	assert.NotContains(t, buf.String(), "goroutine")
}

// --- Chain Integration: Recovery + CORS + RequestID ---

func TestChain_RecoveryWithCORSAndRequestID(t *testing.T) {
	t.Parallel()

	combined := chain.Chain(
		requestid.New(),
		recovery.New(&recovery.Config{
			Output:     io.Discard,
			PrintStack: false,
		}),
		cors.New(nil),
	)

	handler := combined(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("chain panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
}

// --- CORS Edge Cases ---

func TestCORS_NoOriginHeader(t *testing.T) {
	t.Parallel()

	mw := cors.New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Origin header
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// No CORS headers should be set without Origin
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_PreflightRequest(t *testing.T) {
	t.Parallel()

	mw := cors.New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestCORS_SpecificOrigins(t *testing.T) {
	t.Parallel()

	cfg := &cors.Config{
		AllowOrigins: []string{"http://allowed.com", "http://also-allowed.com"},
		AllowMethods: []string{"GET"},
	}
	mw := cors.New(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Allowed origin
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://allowed.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, "http://allowed.com", rr.Header().Get("Access-Control-Allow-Origin"))

	// Disallowed origin
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Origin", "http://evil.com")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Empty(t, rr2.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_NilConfig_UsesDefaults(t *testing.T) {
	t.Parallel()

	mw := cors.New(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Default config allows all origins
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}

// --- RequestID Edge Cases ---

func TestRequestID_ExistingHeader(t *testing.T) {
	t.Parallel()

	mw := requestid.New()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestid.FromRequest(r)
		assert.Equal(t, "my-custom-id", id)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, "my-custom-id", rr.Header().Get("X-Request-ID"))
}

func TestRequestID_GeneratesNewID(t *testing.T) {
	t.Parallel()

	mw := requestid.New()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestid.FromRequest(r)
		assert.NotEmpty(t, id)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
}

func TestRequestID_FromContext_Empty(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id := requestid.FromRequest(req)
	assert.Empty(t, id)
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	t.Parallel()

	mw := requestid.New()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		id := rr.Header().Get("X-Request-ID")
		assert.False(t, ids[id], "duplicate request ID: %s", id)
		ids[id] = true
	}
}

// --- Chain with empty middleware list ---

func TestChain_EmptyMiddlewareList(t *testing.T) {
	t.Parallel()

	combined := chain.Chain()
	handler := combined(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "hello", rr.Body.String())
}

// --- Extremely large header value in request ---

func TestRequestID_ExtremelyLongExistingID(t *testing.T) {
	t.Parallel()

	mw := requestid.New()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	longID := strings.Repeat("a", 10000)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", longID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should use the provided ID even if very long
	assert.Equal(t, longID, rr.Header().Get("X-Request-ID"))
}
