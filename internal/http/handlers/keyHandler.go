package handlers

import (
	"bitback/internal/http/handlers/dto"
	"bitback/internal/interfaces"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// KeyHandler handles HTTP requests related to VLESS key generation.
type KeyHandler struct {
	keyManagerService interfaces.KeyService
}

// NewKeyHandler creates a new instance of KeyHandler.
// It takes a KeyService as a dependency.
func NewKeyHandler(kmService interfaces.KeyService) *KeyHandler {
	return &KeyHandler{
		keyManagerService: kmService,
	}
}

// RegisterRoutes registers the HTTP routes for the KeyHandler.
func (h *KeyHandler) RegisterRoutes(mux *http.ServeMux) {
	// Route for generating a VLESS key for a specific user.
	// Expects userID as a path parameter and optional 'remarks' as a query parameter.
	mux.HandleFunc("GET /api/v1/users/{userID}/vless-key", h.GenerateUserVlessKey)
}

// GenerateUserVlessKey handles the request to generate a VLESS key for a specified user.
// It extracts the userID from the path and optional remarks from query parameters.
func (h *KeyHandler) GenerateUserVlessKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := r.PathValue("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		slog.WarnContext(ctx, "GenerateUserVlessKey: invalid userID format in path", "userID_str", userIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid User ID format in path.")
		return
	}

	// Retrieve 'remarks' from query parameters; use a default if not provided.
	remarks := r.URL.Query().Get("remarks")
	if remarks == "" {
		remarks = "BittenVPN"
	}

	slog.InfoContext(ctx, "GenerateUserVlessKey: request received", "userID", userID, "remarks", remarks)

	// Call the service to generate the VLESS key.
	vlessKey, err := h.keyManagerService.GenerateVlessKeyForUser(ctx, userID, remarks)
	if err != nil {
		slog.ErrorContext(ctx, "GenerateUserVlessKey: failed to generate VLESS key via service", "userID", userID, "error", err)
		// Handle specific errors from the service.
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "no active hosts available") {
			respondWithError(w, http.StatusServiceUnavailable, "Unable to generate key: No active hosts are currently available.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to generate VLESS key.")
		}
		return
	}

	// Prepare and send the successful JSON response.
	response := dto.VlessKeyResponse{
		VlessKey: vlessKey,
		UserID:   userID.String(), // Convert UUID to string for the response.
		Remarks:  remarks,
	}
	slog.InfoContext(ctx, "GenerateUserVlessKey: VLESS key generated successfully", "userID", userID)
	respondWithJSON(w, http.StatusOK, response)
}
