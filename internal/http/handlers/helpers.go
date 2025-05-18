package handlers

import (
	"bitback/internal/http/handlers/dto"
	"bitback/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"strconv"
)

// respondWithError logs an error and sends a JSON error response to the client.
func respondWithError(w http.ResponseWriter, code int, message string) {
	slog.Error("Responding with error", "code", code, "message", message)
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON marshals the payload to JSON and sends it as an HTTP response.
// It sets the Content-Type header to "application/json; charset=utf-8".
// If marshalling fails, it logs the error and sends a 500 Internal Server Error.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal JSON response", "error", err, "payload_type", fmt.Sprintf("%T", payload))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		// Provide a generic error message to the client in case of marshalling failure.
		errorResponse := `{"error": "An internal server error occurred while processing your request."}`
		_, writeErr := w.Write([]byte(errorResponse))
		if writeErr != nil {
			slog.Error("Failed to write error response after marshalling error", "original_error", err, "write_error", writeErr)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		// This error usually means the client disconnected or there was a network issue.
		slog.Error("Failed to write JSON response to client", "error", err)
	}
}

// toSubscriptionResponse converts a models.Subscription to a dto.SubscriptionResponse.
// It handles optional fields like Price and Currency, setting them only if they have meaningful values.
func toSubscriptionResponse(sub *models.Subscription) dto.SubscriptionResponse {
	resp := dto.SubscriptionResponse{
		ID:            sub.ID,
		UserID:        sub.UserID,
		PlanName:      sub.PlanName,
		DurationUnit:  sub.DurationUnit,
		DurationValue: sub.DurationValue,
		StartDate:     sub.StartDate,
		EndDate:       sub.EndDate,
		IsActive:      sub.IsActive,
		PaymentStatus: sub.PaymentStatus,
		AutoRenew:     sub.AutoRenew,
		CreatedAt:     sub.CreatedAt,
		UpdatedAt:     sub.UpdatedAt,
	}
	// Only include price if it's non-zero (assuming price cannot be negative).
	if sub.Price != 0 {
		resp.Price = &sub.Price
	}
	// Only include currency if it's not an empty string.
	if sub.Currency != "" {
		resp.Currency = &sub.Currency
	}
	return resp
}

// getRequestingUserID extracts the authenticated user's ID from the request context.
// This is a placeholder.
func getRequestingUserID(ctx context.Context) (uuid.UUID, error) {
	// TODO: Implement actual user ID retrieval from context.
	dummyUserID, _ := uuid.NewRandom()
	slog.WarnContext(ctx, "getRequestingUserID: USING DUMMY USER ID. Implement proper user ID retrieval from context.")
	return dummyUserID, nil
}

// toHostResponse converts a models.Host to a dto.HostResponse.
func toHostResponse(host *models.Host) dto.HostResponse {
	return dto.HostResponse{
		ID:            host.ID,
		HostName:      host.HostName,
		Country:       host.Country,
		City:          host.City,
		Address:       host.Address,
		Port:          host.Port,
		Protocol:      host.Protocol,
		Network:       host.Network, // Network type.
		PublicKey:     host.PublicKey,
		Flow:          host.Flow,
		RSID:          host.RSID,
		SecurityType:  host.SecurityType,
		SNI:           host.SNI,
		Fingerprint:   host.Fingerprint,
		IsPrivate:     host.IsPrivate,
		IsOnline:      host.IsOnline,
		Status:        host.Status,
		LastCheckedAt: host.LastCheckedAt,
		Region:        host.Region,
		Provider:      host.Provider,
		CreatedAt:     host.CreatedAt,
		UpdatedAt:     host.UpdatedAt,
	}
}

// toUserResponse converts a models.User to a dto.UserResponse.
func toUserResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		TelegramID: user.TelegramID,
		IsActive:   user.IsActive,
		LastLogin:  user.LastLogin,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

// parseUint converts a string to a uint.
// It is a utility function for parsing uint path parameters or query strings.
func parseUint(s string) (uint, error) {
	val, err := strconv.ParseUint(s, 10, 32) // Parse as Uint64, then cast to uint (which is 32 or 64 bit).
	if err != nil {
		return 0, fmt.Errorf("failed to parse '%s' as unsigned integer: %w", s, err)
	}
	return uint(val), nil
}
