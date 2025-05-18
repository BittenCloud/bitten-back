package database

import (
	"bitback/internal/config"
	"bitback/internal/models"
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormDefaultLogger "gorm.io/gorm/logger"
	"log"
	"log/slog"
	"os"
)

// PostgresDB wraps the GORM database instance and application configuration.
type PostgresDB struct {
	gorm *gorm.DB
	cfg  *config.Config
}

// NewPostgresDB initializes a new PostgreSQL database connection using GORM.
// It takes a context and configuration, sets up the GORM logger, establishes the connection,
// configures connection pool settings, and runs auto-migrations for defined models.
func NewPostgresDB(_ context.Context, cfg *config.Config) (*PostgresDB, error) {
	gormLogLevel := cfg.GetGormLogLevel()
	gormSlowThreshold := cfg.DBGormSlowThreshold

	// Configure GORM logger.
	// This logger uses the standard 'log' package for output.
	newLogger := gormDefaultLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // GORM logger writes to os.Stdout.
		gormDefaultLogger.Config{
			SlowThreshold:             gormSlowThreshold, // Threshold for slow SQL queries.
			LogLevel:                  gormLogLevel,      // GORM's own log level (Silent, Error, Warn, Info).
			IgnoreRecordNotFoundError: true,              // Suppress GORM's ErrRecordNotFound errors from logs.
			Colorful:                  true,              // Enable colorful log output.
		},
	)

	dsn := cfg.GetDBDSN()

	// Open a new GORM database connection.
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		slog.Error("Failed to connect to the database", "dsn_host", cfg.DBHost, "dsn_db", cfg.DBName, "error", err)
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Get the underlying sql.DB object for connection pool configuration.
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get underlying *sql.DB from GORM", "error", err)
		// Attempt to close the GORM DB if we failed to get the sql.DB instance.
		if closeErr := closeGormDB(db); closeErr != nil {
			slog.Error("Failed to close GORM DB after error getting *sql.DB", "close_error", closeErr)
		}
		return nil, fmt.Errorf("failed to obtain underlying sql.DB: %w", err)
	}

	// Configure connection pool settings.
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	slog.Info("PostgreSQL connection established successfully.", "host", cfg.DBHost, "port", cfg.DBPort, "dbname", cfg.DBName)
	slog.Debug("GORM logger configured.", "level", cfg.DBGormLogLevel, "slow_query_threshold_ms", gormSlowThreshold.Milliseconds())

	// Automatically migrate the schema for the specified models.
	slog.Info("Running GORM auto-migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Host{},
		&models.Subscription{},
	)
	if err != nil {
		slog.Error("GORM auto-migration failed", "error", err)
	} else {
		slog.Info("GORM auto-migrations completed successfully.")
	}

	return &PostgresDB{
		gorm: db,
		cfg:  cfg,
	}, nil
}

// GetGormClient returns the GORM database client instance.
func (pg *PostgresDB) GetGormClient() *gorm.DB {
	return pg.gorm
}

// closeGormDB attempts to close the GORM database connection if it exists.
func closeGormDB(gormDB *gorm.DB) error {
	if gormDB != nil {
		sqlDB, err := gormDB.DB()
		if err == nil && sqlDB != nil {
			return sqlDB.Close()
		}
		if err != nil {
			return fmt.Errorf("failed to get *sql.DB from gorm.DB for closing: %w", err)
		}
	}
	return nil
}

// Ping checks the database connection by sending a ping.
func (pg *PostgresDB) Ping() {
	slog.Info("Attempting to ping database...")
	if pg.gorm == nil {
		slog.Error("Database connection (gorm.DB) is nil, cannot ping.")
		return
	}
	sqlDB, err := pg.gorm.DB()
	if err != nil {
		slog.Error("Failed to get underlying *sql.DB instance for ping", "error", err)
		return
	}
	// Use a background context for the ping as it's a standalone check.
	err = sqlDB.PingContext(context.Background())
	if err != nil {
		slog.Error("Failed to ping database", "error", err)
	} else {
		slog.Info("Database ping successful.")
	}
}

// Shutdown gracefully closes the connection to the PostgreSQL database.
func (pg *PostgresDB) Shutdown() {
	slog.Info("Closing connection to PostgreSQL...")
	if pg.gorm == nil {
		slog.Warn("Attempting to close a nil database connection (gorm.DB is nil).")
		return
	}

	sqlDB, err := pg.gorm.DB()
	if err != nil {
		slog.Error("Failed to get underlying *sql.DB object for closing during shutdown", "error", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		slog.Error("Error while closing connection to PostgreSQL", "error", err)
	} else {
		slog.Info("Connection to PostgreSQL closed successfully.")
	}
}
