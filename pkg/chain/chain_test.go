package chain

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testMiddleware creates a middleware that appends a value to the X-Order header,
// enabling verification of execution order.
func testMiddleware(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			existing := w.Header().Get("X-Order")
			if existing != "" {
				existing += ","
			}
			w.Header().Set("X-Order", existing+name)
			next.ServeHTTP(w, r)
		})
	}
}

func TestChain_ExecutionOrder(t *testing.T) {
	combined := Chain(
		testMiddleware("first"),
		testMiddleware("second"),
		testMiddleware("third"),
	)

	handler := combined(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "first,second,third", rec.Header().Get("X-Order"))
}

func TestChain_EmptyMiddleware(t *testing.T) {
	combined := Chain()

	called := false
	handler := combined(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("direct"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "direct", rec.Body.String())
}

func TestChain_SingleMiddleware(t *testing.T) {
	combined := Chain(testMiddleware("only"))

	handler := combined(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "only", rec.Header().Get("X-Order"))
}

func TestChain_MiddlewareCanShortCircuit(t *testing.T) {
	blocker := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("blocked"))
			// Does NOT call next.ServeHTTP.
		})
	}

	called := false
	combined := Chain(
		testMiddleware("first"),
		blocker,
		testMiddleware("never"),
	)

	handler := combined(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "blocked", rec.Body.String())
}

func TestChain_PreservesResponseBody(t *testing.T) {
	combined := Chain(testMiddleware("wrap"))

	handler := combined(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/create", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, `{"id":1}`, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
