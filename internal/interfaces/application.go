package interfaces

// Application defines the interface for the main application.
// It outlines the core lifecycle methods for starting and stopping the application.
type Application interface {
	// Start initiates the application, beginning its primary functions
	// such as starting servers, initializing services, etc.
	Start()

	// Shutdown performs a graceful shutdown of the application,
	// ensuring all resources are released and ongoing processes are completed.
	Shutdown()
}
