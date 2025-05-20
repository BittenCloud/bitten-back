package handlers

import (
	"bitback/internal/http/handlers/dto"
	"bitback/internal/interfaces"
	serviceDTO "bitback/internal/services/dto"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// UserHandler handles HTTP requests related to users.
type UserHandler struct {
	userService interfaces.UserService
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(us interfaces.UserService) *UserHandler {
	return &UserHandler{
		userService: us,
	}
}

// RegisterRoutes registers the HTTP routes for user-related actions.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/users", h.CreateUser)
	mux.HandleFunc("GET /v1/users/{userID}", h.GetUser)
	mux.HandleFunc("PUT /v1/users/{userID}", h.UpdateUser)
	mux.HandleFunc("DELETE /v1/users/{userID}", h.DeleteUser)
	mux.HandleFunc("GET /v1/users", h.ListUsers)
}

// CreateUser handles the request to create a new user.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "CreateUser: failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// TODO: Add request DTO validation here

	serviceInput := serviceDTO.CreateUserInput{
		Name:       req.Name,
		Email:      req.Email,
		TelegramID: req.TelegramID,
	}

	user, err := h.userService.RegisterUser(r.Context(), serviceInput)
	if err != nil {
		slog.ErrorContext(ctx, "CreateUser: failed to register user via service", "error", err, "email", req.Email)
		// Check for specific errors like duplicate email.
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			(err.Error() == fmt.Sprintf("user with email '%s' already exists", req.Email)) ||
			strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate key") {
			respondWithError(w, http.StatusConflict, "User with this email already exists.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to create user.")
		}
		return
	}

	slog.InfoContext(ctx, "CreateUser: user created successfully", "userID", user.ID)
	respondWithJSON(w, http.StatusCreated, toUserResponse(user))
}

// GetUser handles the request to retrieve a user by their ID.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := r.PathValue("userID")
	if userIDStr == "" {
		slog.WarnContext(ctx, "GetUser: userID path parameter is missing")
		respondWithError(w, http.StatusBadRequest, "User ID is missing in path.")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		slog.WarnContext(ctx, "GetUser: invalid user ID format in path", "userID_str", userIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid user ID format.")
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		slog.ErrorContext(ctx, "GetUser: failed to get user from service", "userID", userID, "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "User not found.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve user.")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, toUserResponse(user))
}

// UpdateUser handles the request to update an existing user.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := r.PathValue("userID")
	if userIDStr == "" {
		slog.WarnContext(ctx, "UpdateUser: userID path parameter is missing")
		respondWithError(w, http.StatusBadRequest, "User ID is missing in path.")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		slog.WarnContext(ctx, "UpdateUser: invalid user ID format in path", "userID_str", userIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid user ID format.")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "UpdateUser: failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// TODO: Add request DTO validation here.

	serviceInput := serviceDTO.UpdateUserInput{
		Name:       req.Name,
		Email:      req.Email,
		TelegramID: req.TelegramID,
		IsActive:   req.IsActive,
	}

	updatedUser, err := h.userService.UpdateUser(r.Context(), userID, serviceInput)
	if err != nil {
		slog.ErrorContext(ctx, "UpdateUser: failed to update user via service", "userID", userID, "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "User not found.")
		} else if strings.Contains(err.Error(), "email is already in use") {
			respondWithError(w, http.StatusConflict, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to update user.")
		}
		return
	}

	slog.InfoContext(ctx, "UpdateUser: user updated successfully", "userID", updatedUser.ID)
	respondWithJSON(w, http.StatusOK, toUserResponse(updatedUser))
}

// DeleteUser handles the request to (soft) delete a user.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := r.PathValue("userID")
	if userIDStr == "" {
		slog.WarnContext(ctx, "DeleteUser: userID path parameter is missing")
		respondWithError(w, http.StatusBadRequest, "User ID is missing in path.")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		slog.WarnContext(ctx, "DeleteUser: invalid user ID format in path", "userID_str", userIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid user ID format.")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		slog.ErrorContext(ctx, "DeleteUser: failed to delete user via service", "userID", userID, "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "User not found.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete user.")
		}
		return
	}

	slog.InfoContext(ctx, "DeleteUser: user deleted successfully", "userID", userID)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully."})
}

// ListUsers handles the request to retrieve a paginated list of users.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.InfoContext(ctx, "ListUsers: received request to list users")

	// Get pagination parameters from query string.
	query := r.URL.Query()
	pageStr := query.Get("page")
	pageSizeStr := query.Get("pageSize")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // Default to page 1.
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10 // Default page size.
	}
	if pageSize > 100 {
		pageSize = 100
	}

	usersModels, totalItems, err := h.userService.ListUsers(ctx, page, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "ListUsers: failed to retrieve users from service", "error", err, "page", page, "pageSize", pageSize)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve users list.")
		return
	}

	// Convert []models.User to []dto.UserResponse.
	userResponses := make([]dto.UserResponse, len(usersModels))
	for i, u := range usersModels {
		userResponses[i] = toUserResponse(&u)
	}

	totalPages := 0
	if totalItems > 0 && pageSize > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}

	// Handle cases where the requested page is out of bounds.
	// If totalPages is 0 (no items), this condition won't be met.
	if page > totalPages && totalPages > 0 {
		userResponses = []dto.UserResponse{}
		slog.WarnContext(ctx, "ListUsers: requested page is out of bounds",
			"requested_page", page, "total_pages", totalPages, "total_items", totalItems)
	}

	response := dto.PaginatedUsersResponse{
		Users:       userResponses,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}

	slog.InfoContext(ctx, "ListUsers: successfully listed users", "count_in_page", len(userResponses), "total_items", totalItems, "current_page", page)
	respondWithJSON(w, http.StatusOK, response)
}
