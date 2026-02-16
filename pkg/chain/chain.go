// Package chain provides middleware chaining for net/http. It allows multiple
// middleware functions to be composed into a single middleware that applies
// them in the order they are specified.
package chain

import "net/http"

// Chain chains multiple middleware functions into a single middleware.
// Middleware functions are applied in the order provided, meaning the first
// middleware in the list is the outermost wrapper and executes first.
//
// Example:
//
//	combined := chain.Chain(logging, recovery, cors)
//	handler := combined(myHandler)
//
// This is equivalent to: logging(recovery(cors(myHandler)))
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
