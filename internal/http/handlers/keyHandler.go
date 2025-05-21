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
	// Expects userID as a path parameter and optional 'remarks' & 'country' as query parameters.
	mux.HandleFunc("GET /v1/users/{userID}/vless-key", h.GenerateUserVlessKey)
	// Route for generating a VLESS key for a free user.
	// Expects optional 'remarks' & 'country' as query parameters.
	mux.HandleFunc("GET /v1/key/free", h.GenerateFreeVlessKey)
}

// GenerateUserVlessKey handles the request to generate a VLESS key for a specified user.
// It extracts the userID from the path and optional remarks & country from query parameters.
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
		remarks = "BittenVPN" // Default remarks
	}

	// Retrieve 'country' from query parameters.
	countryQuery := r.URL.Query().Get("country")
	var countryPtr *string
	if countryQuery != "" {
		countryPtr = &countryQuery
	}

	slog.InfoContext(ctx, "GenerateUserVlessKey: request received", "userID", userID, "remarks", remarks, "country", countryQuery)

	// Call the service to generate the VLESS key.
	result, err := h.keyManagerService.GenerateVlessKeyForUser(ctx, userID, remarks, countryPtr)
	if err != nil {
		slog.ErrorContext(ctx, "GenerateUserVlessKey: failed to generate VLESS key via service", "userID", userID, "error", err)
		if strings.Contains(err.Error(), "not found") { // User not found
			respondWithError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "no active hosts available") {
			respondWithError(w, http.StatusServiceUnavailable, "Unable to generate key: No active hosts are currently available for your criteria.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to generate VLESS key.")
		}
		return
	}

	// Prepare and send the successful JSON response.
	response := dto.VlessKeyResponse{
		VlessKey:              result.VlessKey,
		UserID:                userID.String(),
		Remarks:               remarks,
		HasActiveSubscription: &result.HasActiveSubscription,
	}
	slog.InfoContext(ctx, "GenerateUserVlessKey: VLESS key generated successfully", "userID", userID, "hasActiveSubscription", result.HasActiveSubscription)
	respondWithJSON(w, http.StatusOK, response)
}

// GenerateFreeVlessKey handles the request to generate a VLESS key for a free user.
func (h *KeyHandler) GenerateFreeVlessKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Retrieve 'remarks' from query parameters; use a default if not provided.
	remarks := r.URL.Query().Get("remarks")
	if remarks == "" {
		remarks = "BittenVPN-Free" // Default remarks for free key
	}

	// Retrieve 'country' from query parameters.
	countryQuery := r.URL.Query().Get("country")
	var countryPtr *string
	if countryQuery != "" {
		countryPtr = &countryQuery
	}

	slog.InfoContext(ctx, "GenerateFreeVlessKey: request received", "remarks", remarks, "country", countryQuery)

	// Call the service to generate the VLESS key.
	vlessKey, err := h.keyManagerService.GenerateFreeVlessKey(ctx, remarks, countryPtr)
	if err != nil {
		slog.ErrorContext(ctx, "GenerateFreeVlessKey: failed to generate VLESS key via service", "error", err)
		if strings.Contains(err.Error(), "no active free hosts available") {
			respondWithError(w, http.StatusServiceUnavailable, "Unable to generate key: No active free hosts are currently available.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to generate VLESS key.")
		}
		return
	}

	// Prepare and send the successful JSON response.
	// UserID is omitted as this key uses a predefined generic user ID.
	// HasActiveSubscription is not applicable here.
	response := dto.VlessKeyResponse{
		VlessKey: vlessKey,
		Remarks:  remarks,
	}
	slog.InfoContext(ctx, "GenerateFreeVlessKey: VLESS key generated successfully")
	respondWithJSON(w, http.StatusOK, response)
}
