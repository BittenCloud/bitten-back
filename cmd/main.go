package main

import (
	"bitback/internal/app"
	"context"
	"log/slog"
	"os"
)

// main is the entry point of the application.
// It creates a new application instance and starts it.
// If application creation fails, it logs the error and exits.
func main() {
	// Create a background context for the application.
	ctx := context.Background()

	// Initialize the application.
	application, err := app.NewApplication(ctx)
	if err != nil {
		// Log the critical error during application initialization.
		// Using slog.Error here assumes the logger might have been partially
		// set up by NewApplication before failing, or it falls back to default.
		// If NewApplication fails very early (e.g., config load), it might use fmt.Fprintf.
		slog.Error("Failed to create application", "error", err)
		os.Exit(1) // Exit with a non-zero status code to indicate failure.
	}

	// Start the application. This will block until the application is shut down.
	application.Start()

	slog.Info("Application has shut down.") // This log will appear after application.Start() returns.
}
