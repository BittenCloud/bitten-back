package handlers

import (
	"net/http"
)

// Router encapsulates the HTTP multiplexer (ServeMux) and provides methods
// for registering routes for different handlers.
type Router struct {
	mux *http.ServeMux
}

// NewRouter creates and returns a new instance of Router, initializing the ServeMux.
func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// RegisterKeyRoutes registers the routes managed by KeyHandler.
// It delegates the actual route registration to the KeyHandler's RegisterRoutes method.
func (r *Router) RegisterKeyRoutes(keyHandler *KeyHandler) {
	keyHandler.RegisterRoutes(r.mux)
}

// RegisterUserRoutes registers the routes managed by UserHandler.
// It delegates the actual route registration to the UserHandler's RegisterRoutes method.
func (r *Router) RegisterUserRoutes(userHandler *UserHandler) {
	userHandler.RegisterRoutes(r.mux)
}

// RegisterSubscriptionRoutes registers the routes managed by SubscriptionHandler.
// It delegates the actual route registration to the SubscriptionHandler's RegisterRoutes method.
func (r *Router) RegisterSubscriptionRoutes(subscriptionHandler *SubscriptionHandler) {
	subscriptionHandler.RegisterRoutes(r.mux)
}

// RegisterHostRoutes registers the routes managed by HostHandler.
// It delegates the actual route registration to the HostHandler's RegisterRoutes method.
func (r *Router) RegisterHostRoutes(hostHandler *HostHandler) {
	hostHandler.RegisterRoutes(r.mux)
}

// GetHandler returns the underlying http.ServeMux instance, which implements http.Handler.
// This allows the router to be used with an http.Server.
func (r *Router) GetHandler() http.Handler {
	return r.mux
}
