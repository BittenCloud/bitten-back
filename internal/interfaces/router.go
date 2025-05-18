package interfaces

import "net/http"

// HttpRouter defines the interface for an HTTP router.
// It provides a way to retrieve the configured HTTP handler.
type HttpRouter interface {
	// GetHandler returns the underlying http.Handler.
	GetHandler() http.Handler
}
