package handlers

import (
	"bitback/internal/http/handlers/dto"
	"bitback/internal/interfaces"
	serviceDTO "bitback/internal/services/dto"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionHandler handles HTTP requests related to subscriptions.
type SubscriptionHandler struct {
	subService interfaces.SubscriptionService
}

// NewSubscriptionHandler creates a new instance of SubscriptionHandler.
func NewSubscriptionHandler(ss interfaces.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subService: ss,
	}
}

// RegisterRoutes registers the HTTP routes for subscription-related actions.
func (h *SubscriptionHandler) RegisterRoutes(mux *http.ServeMux) {
	// Routes for subscriptions specific to a user.
	mux.HandleFunc("POST /api/v1/users/{userID}/subscriptions", h.CreateSubscriptionForUser)
	mux.HandleFunc("GET /api/v1/users/{userID}/subscriptions", h.ListUserSubscriptions)

	// Routes for managing a specific subscription by its ID.
	mux.HandleFunc("GET /api/v1/subscriptions/{subscriptionID}", h.GetSubscriptionByID)
	mux.HandleFunc("PATCH /api/v1/subscriptions/{subscriptionID}/cancel", h.CancelSubscription)
	mux.HandleFunc("PATCH /api/v1/subscriptions/{subscriptionID}/payment", h.UpdatePaymentStatus)
	mux.HandleFunc("PATCH /api/v1/subscriptions/{subscriptionID}/autorenew", h.SetAutoRenew)

	// Reporting routes.
	mux.HandleFunc("GET /api/v1/reports/expiring-subscriptions", h.ListUsersWithExpiringSubscriptions)
	mux.HandleFunc("GET /api/v1/reports/active-by-plan", h.ListActiveSubscriptionsByPlan)
}

// CreateSubscriptionForUser handles the request to create a new subscription for a specified user.
// Expected route: POST /api/v1/users/{userID}/subscriptions
func (h *SubscriptionHandler) CreateSubscriptionForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := r.PathValue("userID")
	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		slog.WarnContext(ctx, "CreateSubscriptionForUser: invalid target userID format in path", "userID_str", userIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid target user ID format in path.")
		return
	}

	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "CreateSubscriptionForUser: failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// TODO: Implement request DTO validation.

	// Ensure UserID in the request body (if present) matches the UserID from the path,
	// or primarily use the UserID from the path. For consistency, the path UserID is preferred.
	if req.UserID != "" {
		parsedBodyUserID, parseErr := uuid.Parse(req.UserID)
		if parseErr != nil || parsedBodyUserID != targetUserID {
			slog.WarnContext(ctx, "CreateSubscriptionForUser: UserID in body does not match UserID in path or is invalid.",
				"path_userID", targetUserID, "body_userID_str", req.UserID)
		}
	}

	serviceInput := serviceDTO.CreateSubscriptionInput{
		UserID:        targetUserID, // Use UserID from path.
		PlanName:      req.PlanName,
		DurationUnit:  req.DurationUnit,
		DurationValue: req.DurationValue,
		StartDate:     req.StartDate,
		Price:         req.Price,
		Currency:      req.Currency,
		PaymentStatus: req.PaymentStatus,
		AutoRenew:     req.AutoRenew,
	}

	subscription, err := h.subService.CreateSubscription(ctx, serviceInput)
	if err != nil {
		slog.ErrorContext(ctx, "CreateSubscriptionForUser: failed to create subscription via service", "error", err, "userID", targetUserID, "plan", req.PlanName)
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "already exists") {
			respondWithError(w, http.StatusConflict, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to create subscription.")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, toSubscriptionResponse(subscription))
}

// GetSubscriptionByID handles the request to retrieve a subscription by its ID.
// Expected route: GET /api/v1/subscriptions/{subscriptionID}
func (h *SubscriptionHandler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subscriptionIDStr := r.PathValue("subscriptionID")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		slog.WarnContext(ctx, "GetSubscriptionByID: invalid subscription ID format in path", "subscriptionID_str", subscriptionIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid subscription ID format.")
		return
	}

	requestingUserID, err := getRequestingUserID(ctx) // Placeholder for actual user auth.
	if err != nil {
		slog.ErrorContext(ctx, "GetSubscriptionByID: failed to get requesting user ID (auth missing/failed)", "error", err)
		respondWithError(w, http.StatusUnauthorized, "Authentication required or failed: "+err.Error())
		return
	}

	subscription, err := h.subService.GetSubscriptionByID(ctx, subscriptionID, requestingUserID)
	if err != nil {
		slog.ErrorContext(ctx, "GetSubscriptionByID: failed to get subscription from service", "error", err, "subscriptionID", subscriptionID)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Subscription not found.")
		} else if strings.Contains(err.Error(), "not authorized") {
			respondWithError(w, http.StatusForbidden, "You are not authorized to view this subscription.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve subscription.")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, toSubscriptionResponse(subscription))
}

// ListUserSubscriptions handles the request to list subscriptions for a specific user.
// Expected route: GET /api/v1/users/{userID}/subscriptions
func (h *SubscriptionHandler) ListUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetUserIDStr := r.PathValue("userID")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		slog.WarnContext(ctx, "ListUserSubscriptions: invalid target userID format in path", "userID_str", targetUserIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid target user ID format in path.")
		return
	}

	// TODO: Add authorization check

	query := r.URL.Query()
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(query.Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 { // Max page size limit.
		pageSize = 100
	}

	subsModels, totalItems, err := h.subService.ListUserSubscriptions(ctx, targetUserID, page, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "ListUserSubscriptions: failed to list user subscriptions from service", "error", err, "userID", targetUserID)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve user subscriptions.")
		return
	}

	subResponses := make([]dto.SubscriptionResponse, len(subsModels))
	for i, s := range subsModels {
		subResponses[i] = toSubscriptionResponse(&s)
	}

	totalPages := 0
	if totalItems > 0 && pageSize > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}
	if page > totalPages && totalPages > 0 { // Handle out-of-bounds page requests.
		subResponses = []dto.SubscriptionResponse{}
		slog.WarnContext(ctx, "ListUserSubscriptions: requested page is out of bounds", "requested_page", page, "total_pages", totalPages)
	}

	response := dto.PaginatedSubscriptionsResponse{
		Subscriptions: subResponses,
		TotalItems:    totalItems,
		TotalPages:    totalPages,
		CurrentPage:   page,
		PageSize:      pageSize,
	}
	slog.InfoContext(ctx, "ListUserSubscriptions: successfully listed subscriptions", "userID", targetUserID, "count_in_page", len(subResponses), "total_items", totalItems)
	respondWithJSON(w, http.StatusOK, response)
}

// CancelSubscription handles the request to cancel a subscription.
// Expected route: PATCH /api/v1/subscriptions/{subscriptionID}/cancel
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subscriptionIDStr := r.PathValue("subscriptionID")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		slog.WarnContext(ctx, "CancelSubscription: invalid subscription ID format", "subscriptionID_str", subscriptionIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid subscription ID format.")
		return
	}

	requestingUserID, err := getRequestingUserID(ctx) // Placeholder for actual user auth.
	if err != nil {
		slog.ErrorContext(ctx, "CancelSubscription: failed to get requesting user ID", "error", err)
		respondWithError(w, http.StatusUnauthorized, "Authentication required or failed: "+err.Error())
		return
	}

	updatedSub, err := h.subService.CancelSubscription(ctx, subscriptionID, requestingUserID)
	if err != nil {
		slog.ErrorContext(ctx, "CancelSubscription: failed to cancel subscription via service", "error", err, "subscriptionID", subscriptionID)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Subscription not found.")
		} else if strings.Contains(err.Error(), "not authorized") {
			respondWithError(w, http.StatusForbidden, "You are not authorized to cancel this subscription.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to cancel subscription.")
		}
		return
	}
	slog.InfoContext(ctx, "CancelSubscription: subscription cancelled successfully", "subscriptionID", subscriptionID)
	respondWithJSON(w, http.StatusOK, toSubscriptionResponse(updatedSub))
}

// UpdatePaymentStatus handles the request to update a subscription's payment status.
// Expected route: PATCH /api/v1/subscriptions/{subscriptionID}/payment
func (h *SubscriptionHandler) UpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subscriptionIDStr := r.PathValue("subscriptionID")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		slog.WarnContext(ctx, "UpdatePaymentStatus: invalid subscription ID format", "subscriptionID_str", subscriptionIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid subscription ID format.")
		return
	}

	// TODO: Add authorization check

	var req dto.UpdateSubscriptionPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "UpdatePaymentStatus: failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// TODO: Validate req.PaymentStatus against a list of allowed statuses.

	updatedSub, err := h.subService.UpdatePaymentStatus(ctx, subscriptionID, req.PaymentStatus)
	if err != nil {
		slog.ErrorContext(ctx, "UpdatePaymentStatus: failed to update payment status via service", "error", err, "subscriptionID", subscriptionID)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Subscription not found.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to update payment status.")
		}
		return
	}
	slog.InfoContext(ctx, "UpdatePaymentStatus: payment status updated successfully", "subscriptionID", subscriptionID, "new_status", req.PaymentStatus)
	respondWithJSON(w, http.StatusOK, toSubscriptionResponse(updatedSub))
}

// SetAutoRenew handles the request to set the auto-renewal flag for a subscription.
// Expected route: PATCH /api/v1/subscriptions/{subscriptionID}/autorenew
func (h *SubscriptionHandler) SetAutoRenew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subscriptionIDStr := r.PathValue("subscriptionID")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		slog.WarnContext(ctx, "SetAutoRenew: invalid subscription ID format", "subscriptionID_str", subscriptionIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid subscription ID format.")
		return
	}

	requestingUserID, err := getRequestingUserID(ctx) // Placeholder for actual user auth.
	if err != nil {
		slog.ErrorContext(ctx, "SetAutoRenew: failed to get requesting user ID", "error", err)
		respondWithError(w, http.StatusUnauthorized, "Authentication required or failed: "+err.Error())
		return
	}

	var req dto.SetSubscriptionAutoRenewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "SetAutoRenew: failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	updatedSub, err := h.subService.SetAutoRenew(ctx, subscriptionID, requestingUserID, req.AutoRenew)
	if err != nil {
		slog.ErrorContext(ctx, "SetAutoRenew: failed to set auto-renew status via service", "error", err, "subscriptionID", subscriptionID)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Subscription not found.")
		} else if strings.Contains(err.Error(), "not authorized") {
			respondWithError(w, http.StatusForbidden, "You are not authorized to modify this subscription.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to set auto-renew status.")
		}
		return
	}
	slog.InfoContext(ctx, "SetAutoRenew: auto-renew status updated successfully", "subscriptionID", subscriptionID, "auto_renew_set_to", req.AutoRenew)
	respondWithJSON(w, http.StatusOK, toSubscriptionResponse(updatedSub))
}

// ListUsersWithExpiringSubscriptions handles the request to generate a report of users with subscriptions nearing expiration.
// Expected route: GET /api/v1/reports/expiring-subscriptions
func (h *SubscriptionHandler) ListUsersWithExpiringSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.InfoContext(ctx, "ListUsersWithExpiringSubscriptions: received request for expiring subscriptions report")

	// TODO: Add authorization check

	query := r.URL.Query()
	daysStr := query.Get("days_in_advance")
	pageStr := query.Get("page")
	pageSizeStr := query.Get("pageSize")

	daysInAdvance, err := strconv.Atoi(daysStr)
	if err != nil || daysInAdvance < 0 {
		daysInAdvance = 7 // Default to 7 days in advance.
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 { // Max page size limit for reports.
		pageSize = 100
	}

	reportData, totalItems, err := h.subService.GetUsersWithExpiringSubscriptions(ctx, daysInAdvance, page, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "ListUsersWithExpiringSubscriptions: failed to get report from service", "error", err, "days_in_advance", daysInAdvance, "page", page)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate expiring subscriptions report.")
		return
	}

	// Convert service DTO to handler DTO.
	responseData := make([]dto.UserWithExpiringSubscriptionsResponse, len(reportData))
	for i, data := range reportData {
		expiringSubsDTO := make([]dto.ExpiringSubscriptionItemResponse, len(data.ExpiringSubscriptions))
		for j, subInfo := range data.ExpiringSubscriptions {
			expiringSubsDTO[j] = dto.ExpiringSubscriptionItemResponse{
				SubscriptionID: subInfo.ID,
				PlanName:       subInfo.PlanName,
				EndDate:        subInfo.EndDate,
				DurationUnit:   subInfo.DurationUnit,
				DurationValue:  subInfo.DurationValue,
				AutoRenew:      subInfo.AutoRenew,
			}
		}
		responseData[i] = dto.UserWithExpiringSubscriptionsResponse{
			User:                  toUserResponse(&data.User),
			ExpiringSubscriptions: expiringSubsDTO,
		}
	}

	totalPages := 0
	if totalItems > 0 && pageSize > 0 {
		// totalItems here refers to the total number of expiring *subscriptions* or *users with expiring subscriptions*
		// depending on the service layer's pagination strategy.
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}
	if page > totalPages && totalPages > 0 {
		responseData = []dto.UserWithExpiringSubscriptionsResponse{}
		slog.WarnContext(ctx, "ListUsersWithExpiringSubscriptions: requested page is out of bounds", "requested_page", page, "total_pages", totalPages)
	}

	paginatedResponse := dto.PaginatedUserExpiringSubscriptionsResponse{
		Data:        responseData,
		TotalItems:  totalItems,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}

	slog.InfoContext(ctx, "ListUsersWithExpiringSubscriptions: report generated successfully", "users_in_page", len(responseData), "total_items_for_pagination", totalItems)
	respondWithJSON(w, http.StatusOK, paginatedResponse)
}

// ListActiveSubscriptionsByPlan handles the request to list active subscriptions filtered by plan name.
// Expected route: GET /api/v1/reports/active-by-plan
func (h *SubscriptionHandler) ListActiveSubscriptionsByPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.InfoContext(ctx, "ListActiveSubscriptionsByPlan: received request for active subscriptions by plan")

	// TODO: Add authorization check.

	query := r.URL.Query()
	planName := query.Get("plan_name")
	pageStr := query.Get("page")
	pageSizeStr := query.Get("pageSize")

	if strings.TrimSpace(planName) == "" {
		slog.WarnContext(ctx, "ListActiveSubscriptionsByPlan: missing 'plan_name' query parameter")
		respondWithError(w, http.StatusBadRequest, "Query parameter 'plan_name' is required.")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 { // Max page size limit.
		pageSize = 100
	}

	subsModels, totalItems, err := h.subService.ListActiveSubscriptionsByPlan(ctx, planName, page, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "ListActiveSubscriptionsByPlan: failed to retrieve subscriptions from service", "error", err, "plan_name", planName)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve subscriptions list for plan: %s.", planName))
		return
	}

	subResponses := make([]dto.SubscriptionResponse, len(subsModels))
	for i, s := range subsModels {
		subResponses[i] = toSubscriptionResponse(&s)
	}

	totalPages := 0
	if totalItems > 0 && pageSize > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}
	if page > totalPages && totalPages > 0 {
		subResponses = []dto.SubscriptionResponse{}
		slog.WarnContext(ctx, "ListActiveSubscriptionsByPlan: requested page is out of bounds", "requested_page", page, "total_pages", totalPages)
	}

	response := dto.PaginatedSubscriptionsResponse{
		Subscriptions: subResponses,
		TotalItems:    totalItems,
		TotalPages:    totalPages,
		CurrentPage:   page,
		PageSize:      pageSize,
	}

	slog.InfoContext(ctx, "ListActiveSubscriptionsByPlan: successfully listed subscriptions", "plan_name", planName, "count_in_page", len(subResponses), "total_items", totalItems)
	respondWithJSON(w, http.StatusOK, response)
}
