package app

import (
	appRouter "bitback/internal/http/handlers"
	appServer "bitback/internal/http/server"
	"bitback/internal/interfaces"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	apiServer interfaces.ApiServer
}

func NewApplication(ctx context.Context) *Application {
	application := new(Application)
	err := application.initialize(ctx)
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		panic(err)
	}

	slog.Info("Application initialized")
	return application
}

func (app *Application) Start() {
	slog.Info("Starting application...")

	go func() {
		if err := app.apiServer.Run(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error while running api server", "error", err)
			app.Shutdown()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Application received shutdown signal")
	app.Shutdown()
}

func (app *Application) Shutdown() {
	slog.Info("Shutting down application...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if app.apiServer != nil {
		err := app.apiServer.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("Error during API server shutdown", "error", err)
		}
		slog.Info("Application shutdown complete")
	}
}

func (app *Application) initialize(ctx context.Context) error {
	if err := app.setupLogger(ctx); err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	if err := app.setupApiServer(ctx); err != nil {
		return fmt.Errorf("failed to setup API server: %w", err)
	}

	return nil
}

func (app *Application) setupLogger(_ context.Context) error {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
	slog.Info("Logger configured")
	return nil
}

func (app *Application) setupApiServer(ctx context.Context) error {
	slog.Info("Setting up API server...")

	myRouter := appRouter.NewRouter()

	app.apiServer = appServer.NewApiServer(myRouter)
	app.apiServer.CreateAndPrepare()
	slog.Info("API server setup complete")
	return nil
}
