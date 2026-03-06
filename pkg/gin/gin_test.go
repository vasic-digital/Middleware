package gin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"digital.vasic.middleware/pkg/cors"
	"digital.vasic.middleware/pkg/logging"
	"digital.vasic.middleware/pkg/recovery"
	"digital.vasic.middleware/pkg/requestid"

	adapter "digital.vasic.middleware/pkg/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestWrap_CORSMiddleware(t *testing.T) {
	r := gin.New()
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(adapter.Wrap(cors.New(cfg)))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestWrap_RecoveryMiddleware(t *testing.T) {
	r := gin.New()
	cfg := recovery.DefaultConfig()
	r.Use(adapter.Wrap(recovery.New(cfg)))
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWrap_RequestIDMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(adapter.Wrap(requestid.New()))
	r.GET("/test", func(c *gin.Context) {
		id := requestid.FromRequest(c.Request)
		c.String(http.StatusOK, id)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestWrap_LoggingMiddleware(t *testing.T) {
	r := gin.New()
	cfg := logging.DefaultConfig()
	r.Use(adapter.Wrap(logging.New(cfg)))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWrap_ChainedMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(adapter.Wrap(requestid.New()))
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = []string{"*"}
	r.Use(adapter.Wrap(cors.New(corsCfg)))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}
