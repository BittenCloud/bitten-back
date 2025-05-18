package interfaces

import "gorm.io/gorm"

// SQLDatabase defines the interface for SQL database operations.
// It includes methods for health checking, graceful shutdown, and accessing the underlying GORM client.
type SQLDatabase interface {
	// Ping checks the connectivity to the database.
	Ping()

	// Shutdown gracefully closes the database connection and releases resources.
	Shutdown()

	// GetGormClient returns the underlying GORM database client instance.
	// This allows services and repositories to perform database operations using GORM.
	GetGormClient() *gorm.DB
}
