package server

import (
	"bitback/internal/config"
	"bitback/internal/interfaces"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// ApiServer represents the HTTP API server.
// It holds the router, HTTP server configuration, and general application configuration.
type ApiServer struct {
	router     interfaces.HttpRouter
	httpServer *http.Server
	cfg        *config.Config
}

// NewApiServer creates a new instance of ApiServer.
// It requires an HttpRouter and application Config for setup.
func NewApiServer(r interfaces.HttpRouter, cfg *config.Config) *ApiServer {
	return &ApiServer{
		router: r,
		cfg:    cfg,
	}
}

// CreateAndPrepare initializes and configures the underlying http.Server instance.
// This includes setting the address, handler (router), and various timeout values
// based on the application configuration.
// It returns the prepared ApiServer instance to allow for method chaining.
func (a *ApiServer) CreateAndPrepare() interfaces.ApiServer {
	slog.Info("Creating and preparing API server")

	serverAddr := a.cfg.GetApiAddr()

	a.httpServer = &http.Server{
		Addr:              serverAddr,
		Handler:           a.router.GetHandler(), // Use the handler from the injected router.
		ReadTimeout:       a.cfg.ReadTimeout,
		WriteTimeout:      a.cfg.WriteTimeout,
		IdleTimeout:       a.cfg.IdleTimeout,
		ReadHeaderTimeout: a.cfg.ReadHeaderTimeout,
	}
	slog.Info("API server configured", "address", serverAddr)
	return a
}

// Run starts the HTTP server and begins listening for requests.
// This is a blocking call and will only return when the server is stopped
// or an unrecoverable error occurs.
func (a *ApiServer) Run() error {
	if a.httpServer == nil {
		slog.Error("API server is not prepared. Call CreateAndPrepare() first.")
		return fmt.Errorf("API server not prepared, call CreateAndPrepare() before Run()")
	}

	slog.Info("Starting API server listeners...", "address", a.httpServer.Addr)
	err := a.httpServer.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			// This error is expected during a graceful shutdown.
			slog.Info("API server stopped gracefully.")
		} else {
			// An unexpected error occurred.
			slog.Error("API server failed to run", "error", err)
		}
		return err // Return the error for the caller to handle.
	}
	// This part is typically not reached if ListenAndServe runs indefinitely
	// unless it's closed by a Shutdown call which would result in http.ErrServerClosed.
	slog.Info("API server has stopped (this log might indicate an unexpected stop if not preceded by ErrServerClosed).")
	return nil
}

// Shutdown gracefully shuts down the HTTP server.
// It attempts to close active connections within the timeout provided by the context.
func (a *ApiServer) Shutdown(ctx context.Context) error {
	slog.Info("Attempting to shut down API server gracefully...")
	if a.httpServer != nil {
		err := a.httpServer.Shutdown(ctx)
		if err != nil {
			slog.Error("API server shutdown error", "error", err)
			return err
		}
		slog.Info("API server shutdown completed.")
		return nil
	}
	slog.Warn("Shutdown called on an unprepared or nil API server.")
	return nil // No server to shut down.
}
