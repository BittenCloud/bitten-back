package app

import (
	"bitback/internal/config"
	repoImpl "bitback/internal/connectors/sql"
	"bitback/internal/database"
	appRouter "bitback/internal/http/handlers"
	appServer "bitback/internal/http/server"
	"bitback/internal/interfaces"
	"bitback/internal/services"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Application encapsulates the core components of the service,
// including the API server, database connection, and configuration.
type Application struct {
	apiServer interfaces.ApiServer
	database  interfaces.SQLDatabase
	cfg       *config.Config
}

// NewApplication creates and initializes a new instance of the Application.
// It sets up configuration, logging, database connection, repositories, services,
// HTTP handlers, router, and the API server.
func NewApplication(ctx context.Context) (*Application, error) {
	// Load application configuration.
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use Fprintf for critical startup errors before logger might be fully initialized.
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Setup global structured logger (slog).
	if err := setupGlobalLogger(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Logger setup error: %v\n", err)
		return nil, fmt.Errorf("logger setup failed: %w", err)
	}
	slog.Info("Logger configured successfully.", "level", cfg.LogLevel)
	slog.Info("Configuration loaded successfully.")

	// Initialize database connection.
	// 'db' will be of type *database.PostgresDB, which implements interfaces.SQLDatabase.
	db, err := database.NewPostgresDB(ctx, cfg)
	if err != nil {
		slog.Error("Database initialization failed.", "error", err)
		return nil, fmt.Errorf("database setup failed: %w", err)
	}
	slog.Info("Database initialized successfully.")

	// Initialize repositories.
	userRepo := repoImpl.NewUserRepository(db)
	subscriptionRepo := repoImpl.NewSubscriptionRepository(db)
	hostRepo := repoImpl.NewHostRepository(db)
	slog.Info("Repositories initialized successfully.")

	// Initialize services.
	userService := services.NewUserService(userRepo)
	subscriptionService := services.NewSubscriptionService(subscriptionRepo, userRepo) // SubscriptionService also requires userRepo.
	hostService := services.NewHostService(hostRepo)
	keyService := services.NewKeyService(userRepo, hostRepo, subscriptionRepo) // KeyService requires userRepo and hostRepo.
	slog.Info("Services initialized successfully.")

	// Initialize HTTP handlers.
	userHandler := appRouter.NewUserHandler(userService)
	subscriptionHandler := appRouter.NewSubscriptionHandler(subscriptionService)
	hostHandler := appRouter.NewHostHandler(hostService)
	keyManagerHandler := appRouter.NewKeyHandler(keyService)
	slog.Info("HTTP handlers initialized successfully.")

	// Configure the HTTP router and register routes for each handler.
	router := appRouter.NewRouter() // router will be of type *appRouter.Router.
	router.RegisterUserRoutes(userHandler)
	router.RegisterSubscriptionRoutes(subscriptionHandler)
	router.RegisterHostRoutes(hostHandler)
	router.RegisterKeyRoutes(keyManagerHandler)
	slog.Info("Router configured successfully.")

	// Create and prepare the API server.
	apiHttpServer := appServer.NewApiServer(router, cfg)
	preparedApiServer := apiHttpServer.CreateAndPrepare()
	slog.Info("API server prepared successfully.")

	application := &Application{
		apiServer: preparedApiServer,
		database:  db,
		cfg:       cfg,
	}

	slog.Info("Application initialized successfully.")
	return application, nil
}

// setupGlobalLogger configures the global slog logger instance.
func setupGlobalLogger(_ context.Context, cfg *config.Config) error {
	logLevel := cfg.GetSlogLevel()
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,     // Include source file and line number in logs.
		Level:     logLevel, // Set the minimum log level.
	})
	slog.SetDefault(slog.New(jsonHandler))
	return nil
}

// Start begins the application's operation, primarily by running the API server.
// It also sets up signal handling for graceful shutdown.
func (app *Application) Start() {
	slog.Info("Starting application...",
		"api_address", app.cfg.GetApiAddr(),
		"log_level", app.cfg.LogLevel,
	)

	// Channel to listen for server errors.
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("API server request processing loop starting...")
		serverErrors <- app.apiServer.Run()
	}()

	// Channel to listen for OS signals for termination.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a server error or a termination signal is received.
	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("API server critical failure.", "error", err)
		} else if errors.Is(err, http.ErrServerClosed) {
			slog.Info("API server was stopped gracefully (http.ErrServerClosed received).")
		}
	case sig := <-quit:
		slog.Info("Application received termination signal.", "signal", sig.String())
	}

	// Initiate graceful shutdown.
	app.Shutdown()
}

// Shutdown performs a graceful shutdown of the application components,
// including the API server and database connection.
func (app *Application) Shutdown() {
	slog.Info("Initiating application shutdown sequence...")

	// Create a context with a timeout for the shutdown process.
	shutdownTimeout := app.cfg.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 15 * time.Second // Default fallback if config is zero or negative.
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown the API server.
	if app.apiServer != nil {
		slog.Info("Stopping API server...", "timeout", shutdownTimeout)
		if err := app.apiServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error during API server shutdown.", "error", err)
		} else {
			slog.Info("API server stopped successfully.")
		}
	}

	// Close the database connection.
	if app.database != nil {
		slog.Info("Closing database connection...")
		app.database.Shutdown()
		slog.Info("Database connection shutdown process initiated.")
	}

	slog.Info("Application shutdown process completed.")
}
