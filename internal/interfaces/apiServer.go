package interfaces

import "context"

// ApiServer defines the interface for an API server.
// It includes methods for preparing, running, and shutting down the server.
type ApiServer interface {
	// CreateAndPrepare initializes and prepares the API server for running.
	// It returns the prepared ApiServer instance, allowing for chained calls.
	CreateAndPrepare() ApiServer

	// Run starts the API server, making it listen for incoming requests.
	// This method will block until the server is stopped or an error occurs.
	Run() error

	// Shutdown gracefully shuts down the API server without interrupting active connections.
	// It waits for a given duration specified by the context for connections to complete.
	Shutdown(ctx context.Context) error
}
