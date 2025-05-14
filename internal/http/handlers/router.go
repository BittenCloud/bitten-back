package handlers

import (
	"encoding/json" // Импортируем пакет json
	"fmt"
	"log/slog"
	"net/http"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	slog.Error("Responding with error", "code", code, "message", message)
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal JSON response", "error", err, "payload", payload)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := fmt.Sprintf(`{"error": "failed to marshal JSON response: %v"}`, err)
		_, writeErr := w.Write([]byte(errorResponse))
		if writeErr != nil {
			slog.Error("Failed to write error response after marshalling error", "error", writeErr)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		slog.Error("Failed to write JSON response to client", "error", err)
	}
}

func (r *Router) CollectRoutes() error {
	slog.Info("Collecting routes...")

	r.mux.HandleFunc("GET /health", func(w http.ResponseWriter, req *http.Request) {

		slog.Info("Handling /health request", "method", req.Method, "path", req.URL.Path)

		respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})

		slog.Info("Handled /health request successfully")
	})

	r.mux.HandleFunc("GET /randomKey", r.GetRandomKey)

	slog.Info("Routes collected successfully")
	return nil
}

func (r *Router) GetHandler() http.Handler {
	return r.mux
}
