package altsvc_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"digital.vasic.middleware/pkg/altsvc"
)

func TestAltSvc_DefaultH3(t *testing.T) {
	handler := altsvc.New(altsvc.DefaultConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Alt-Svc"), "h3=")
}

func TestAltSvc_CustomPort(t *testing.T) {
	cfg := &altsvc.Config{
		Enabled: true,
		H3Port:  "8443",
		MaxAge:  86400,
	}
	handler := altsvc.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, `h3=":8443"; ma=86400`, w.Header().Get("Alt-Svc"))
}

func TestAltSvc_Disabled(t *testing.T) {
	cfg := &altsvc.Config{Enabled: false}
	handler := altsvc.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Alt-Svc"))
}

func TestAltSvc_NilConfigUsesDefaults(t *testing.T) {
	handler := altsvc.New(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `h3=":443"; ma=86400`, w.Header().Get("Alt-Svc"))
}

func TestAltSvc_PassesThroughToHandler(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})
	handler := altsvc.New(altsvc.DefaultConfig())(inner)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/create", nil)
	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestDefaultConfig(t *testing.T) {
	cfg := altsvc.DefaultConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "443", cfg.H3Port)
	assert.Equal(t, 86400, cfg.MaxAge)
}
