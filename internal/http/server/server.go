package server

import (
	"bitback/internal/interfaces"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type ApiServer struct {
	router     interfaces.HttpRouter
	httpServer *http.Server
}

func NewApiServer(r interfaces.HttpRouter) *ApiServer {
	return &ApiServer{
		router: r,
	}
}

func (a *ApiServer) CreateAndPrepare() interfaces.ApiServer {
	slog.Info("Creating and preparing API server...")

	err := a.router.CollectRoutes()
	if err != nil {
		slog.Error("Failed to collect routes", "error", err)
		panic(err)
	}

	a.httpServer = &http.Server{
		Addr:              ":9080",
		Handler:           a.router.GetHandler(),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("API server prepared", "address", a.httpServer.Addr)
	return a
}

func (a *ApiServer) Run() error {
	if a.httpServer == nil {
		slog.Error("API server is not prepared. Call CreateAndPrepare() first.")
		return fmt.Errorf("API server not prepared")
	}
	slog.Info("Starting API server...", "address", a.httpServer.Addr)
	err := a.httpServer.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("API server stopped gracefully ")
		} else {
			slog.Error("API server failed to run", "error", err)
		}
		return err
	}
	slog.Info("API server stopped gracefully")
	return nil
}

func (a *ApiServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down API server...")
	if a.httpServer != nil {
		return a.httpServer.Shutdown(ctx)
	}
	return nil
}
