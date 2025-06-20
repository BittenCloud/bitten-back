package config

import (
	"fmt"
	gormLogger "gorm.io/gorm/logger"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config stores all application configuration parameters.
type Config struct {
	LogLevel            string        // Global logging level for slog (e.g., "debug", "info", "warn", "error").
	DBHost              string        // Database host address.
	DBPort              int           // Database port number.
	DBUser              string        // Database username.
	DBPassword          string        // Database password.
	DBName              string        // Database name.
	DBSslMode           string        // SSL mode for database connection (e.g., "disable", "require").
	DBMaxOpenConns      int           // Maximum number of open connections to the database.
	DBMaxIdleConns      int           // Maximum number of connections in the idle connection pool.
	DBConnMaxLifetime   time.Duration // Maximum amount of time a connection may be reused.
	DBGormLogLevel      string        // GORM's specific logger level (e.g., "silent", "error", "warn", "info").
	DBGormSlowThreshold time.Duration // Threshold for GORM to log slow queries.

	ApiHost           string        // Host for the API server to listen on (e.g., "0.0.0.0" for all interfaces).
	ApiPort           int           // Port for the API server to listen on.
	ReadTimeout       time.Duration // Maximum duration for reading the entire request, including the body.
	WriteTimeout      time.Duration // Maximum duration before timing out writes of the response.
	IdleTimeout       time.Duration // Maximum amount of time to wait for the next request when keep-alives are enabled.
	ReadHeaderTimeout time.Duration // Amount of time allowed to read request headers.
	ShutdownTimeout   time.Duration // Graceful shutdown period for the server.

	InstanceConnectionName string // Cloud SQL instance connection name (for Cloud Run)
}

// LoadConfig loads configuration from environment variables, applying default values if not set.
// It returns a Config struct or an error if critical configurations are invalid.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// Default values
		LogLevel:            "info",
		DBHost:              "localhost",
		DBPort:              5432,
		DBUser:              "admin",
		DBPassword:          "truapp00", // I apologize to myself for this.
		DBName:              "bitcloud",
		DBSslMode:           "disable",
		DBMaxOpenConns:      25,
		DBMaxIdleConns:      25,
		DBConnMaxLifetime:   5 * time.Minute,
		DBGormLogLevel:      "warn",
		DBGormSlowThreshold: 200 * time.Millisecond,
		ApiPort:             9080, // API_HOST defaults to "" (empty string), meaning http.Server will use localhost.
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
		IdleTimeout:         120 * time.Second,
		ReadHeaderTimeout:   5 * time.Second,
		ShutdownTimeout:     15 * time.Second,
	}

	// Load global slog logging level.
	if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
		cfg.LogLevel = strings.ToLower(logLevelEnv)
		if !isValidSlogLevel(cfg.LogLevel) {
			slog.Warn("Invalid LOG_LEVEL environment variable. Using default.", "value", logLevelEnv, "default", "info")
			cfg.LogLevel = "info"
		}
	}

	// Load database connection variables.
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.DBHost = dbHost
	}
	if dbPortStr := os.Getenv("DB_PORT"); dbPortStr != "" {
		dbPort, err := strconv.Atoi(dbPortStr)
		if err != nil {
			slog.Error("Invalid DB_PORT environment variable. Must be an integer.", "value", dbPortStr, "error", err)
			return nil, fmt.Errorf("invalid DB_PORT: %w", err)
		}
		cfg.DBPort = dbPort
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.DBUser = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		cfg.DBPassword = dbPassword
	} else if cfg.DBPassword == "truapp00" { // I don't see this.
		slog.Warn("DB_PASSWORD is using the default insecure value.")
	} else if cfg.DBPassword == "" {
		slog.Warn("DB_PASSWORD is not set and no default is configured (or default was empty). This might lead to connection issues.")
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.DBName = dbName
	}
	if dbSslMode := os.Getenv("DB_SSLMODE"); dbSslMode != "" {
		cfg.DBSslMode = dbSslMode
	}

	// Load database connection pool settings.
	if dbMaxOpenConnsStr := os.Getenv("DB_MAX_OPEN_CONNS"); dbMaxOpenConnsStr != "" {
		val, err := strconv.Atoi(dbMaxOpenConnsStr)
		if err == nil && val > 0 {
			cfg.DBMaxOpenConns = val
		} else if err != nil {
			slog.Warn("Invalid DB_MAX_OPEN_CONNS environment variable. Using default.", "value", dbMaxOpenConnsStr, "error", err)
		}
	}
	if dbMaxIdleConnsStr := os.Getenv("DB_MAX_IDLE_CONNS"); dbMaxIdleConnsStr != "" {
		val, err := strconv.Atoi(dbMaxIdleConnsStr)
		if err == nil && val > 0 {
			cfg.DBMaxIdleConns = val
		} else if err != nil {
			slog.Warn("Invalid DB_MAX_IDLE_CONNS environment variable. Using default.", "value", dbMaxIdleConnsStr, "error", err)
		}
	}
	if dbConnMaxLifetimeStr := os.Getenv("DB_CONN_MAX_LIFETIME_MINUTES"); dbConnMaxLifetimeStr != "" {
		val, err := strconv.Atoi(dbConnMaxLifetimeStr)
		if err == nil && val > 0 {
			cfg.DBConnMaxLifetime = time.Duration(val) * time.Minute
		} else if err != nil {
			slog.Warn("Invalid DB_CONN_MAX_LIFETIME_MINUTES environment variable. Using default.", "value", dbConnMaxLifetimeStr, "error", err)
		}
	}

	// Load GORM logger settings.
	if gormLogLevelEnv := os.Getenv("DB_GORM_LOG_LEVEL"); gormLogLevelEnv != "" {
		cfg.DBGormLogLevel = strings.ToLower(gormLogLevelEnv)
		if !isValidGormLogLevel(cfg.DBGormLogLevel) {
			slog.Warn("Invalid DB_GORM_LOG_LEVEL environment variable. Using default.", "value", gormLogLevelEnv, "default", "warn")
			cfg.DBGormLogLevel = "warn"
		}
	}
	if gormSlowThresholdMsStr := os.Getenv("DB_GORM_SLOW_THRESHOLD_MS"); gormSlowThresholdMsStr != "" {
		val, err := strconv.Atoi(gormSlowThresholdMsStr)
		if err == nil && val > 0 {
			cfg.DBGormSlowThreshold = time.Duration(val) * time.Millisecond
		} else if err != nil {
			slog.Warn("Invalid DB_GORM_SLOW_THRESHOLD_MS environment variable. Using default.",
				"value", gormSlowThresholdMsStr, "default_ms", cfg.DBGormSlowThreshold.Milliseconds(), "error", err)
		}
	}

	// Load API server settings.
	if apiHost := os.Getenv("API_HOST"); apiHost != "" {
		cfg.ApiHost = apiHost
	}
	if apiPortStr := os.Getenv("API_PORT"); apiPortStr != "" {
		apiPort, err := strconv.Atoi(apiPortStr)
		if err != nil {
			slog.Error("Invalid API_PORT environment variable. Must be an integer.", "value", apiPortStr, "error", err)
			return nil, fmt.Errorf("invalid API_PORT: %w", err)
		}
		cfg.ApiPort = apiPort
	}

	if instanceConnectionName := os.Getenv("INSTANCE_CONNECTION_NAME"); instanceConnectionName != "" {
		cfg.InstanceConnectionName = instanceConnectionName
	}

	// Load API server timeout settings using a helper function.
	loadDurationFromEnv("API_READ_TIMEOUT_SECONDS", &cfg.ReadTimeout, time.Second, cfg.ReadTimeout)
	loadDurationFromEnv("API_WRITE_TIMEOUT_SECONDS", &cfg.WriteTimeout, time.Second, cfg.WriteTimeout)
	loadDurationFromEnv("API_IDLE_TIMEOUT_SECONDS", &cfg.IdleTimeout, time.Second, cfg.IdleTimeout)
	loadDurationFromEnv("API_READ_HEADER_TIMEOUT_SECONDS", &cfg.ReadHeaderTimeout, time.Second, cfg.ReadHeaderTimeout)
	loadDurationFromEnv("API_SHUTDOWN_TIMEOUT_SECONDS", &cfg.ShutdownTimeout, time.Second, cfg.ShutdownTimeout)

	slog.Info("Configuration loaded successfully.")
	return cfg, nil
}

// loadDurationFromEnv helper loads a time.Duration value from an environment variable.
// If the environment variable is not set or invalid, it logs a warning and keeps the target unchanged (uses its default).
func loadDurationFromEnv(envKey string, target *time.Duration, unit time.Duration, defaultValue time.Duration) {
	envValStr := os.Getenv(envKey)
	if envValStr == "" {
		// Variable not set, default is already in target.
		return
	}

	val, err := strconv.Atoi(envValStr)
	if err == nil && val >= 0 { // Allow 0 for timeouts, though often implies no timeout or system default.
		*target = time.Duration(val) * unit
	} else {
		slog.Warn(fmt.Sprintf("Invalid %s environment variable. Using default.", envKey),
			"value", envValStr, "default", defaultValue.String(), "error", err)
		// *target remains as its pre-set default value.
	}
}

// GetDBDSN returns the database connection string (Data Source Name).
func (c *Config) GetDBDSN() string {
	if c.InstanceConnectionName != "" {
		return fmt.Sprintf("host=/cloudsql/%s dbname=%s user=%s password=%s sslmode=disable",
			c.InstanceConnectionName, c.DBName, c.DBUser, c.DBPassword)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSslMode)
}

// GetApiAddr returns the network address for the API server (e.g., "0.0.0.0:9080" or ":9080").
func (c *Config) GetApiAddr() string {
	return fmt.Sprintf("%s:%d", c.ApiHost, c.ApiPort)
}

// GetSlogLevel converts the configured string logging level to the slog.Level type.
// Defaults to slog.LevelInfo if an unknown level is specified.
func (c *Config) GetSlogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		slog.Warn("Unknown slog level specified in config, defaulting to Info.", "configured_level", c.LogLevel)
		return slog.LevelInfo
	}
}

// GetGormLogLevel converts the configured GORM string logging level to gormLogger.LogLevel.
// Defaults to gormLogger.Warn if an unknown level is specified.
func (c *Config) GetGormLogLevel() gormLogger.LogLevel {
	switch strings.ToLower(c.DBGormLogLevel) {
	case "silent":
		return gormLogger.Silent
	case "error":
		return gormLogger.Error
	case "warn":
		return gormLogger.Warn
	case "info":
		return gormLogger.Info
	default:
		slog.Warn("Unknown GORM log level specified in config, defaulting to Warn.", "configured_level", c.DBGormLogLevel)
		return gormLogger.Warn
	}
}

// isValidSlogLevel checks if the provided string is a valid slog log level.
func isValidSlogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "warning", "error", "err":
		return true
	default:
		return false
	}
}

// isValidGormLogLevel checks if the provided string is a valid GORM log level.
func isValidGormLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "silent", "error", "warn", "info":
		return true
	default:
		return false
	}
}
