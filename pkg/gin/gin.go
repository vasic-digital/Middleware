// Package gin provides adapters to use digital.vasic.middleware with the Gin framework.
package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Wrap converts a standard net/http middleware into a gin.HandlerFunc.
func Wrap(middleware func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Next()
		})
		middleware(next).ServeHTTP(c.Writer, c.Request)
	}
}
