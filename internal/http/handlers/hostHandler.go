package handlers

import (
	"bitback/internal/http/handlers/dto"
	"bitback/internal/interfaces"
	"bitback/internal/models/customTypes"
	serviceDTO "bitback/internal/services/dto"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// HostHandler handles HTTP requests related to hosts.
type HostHandler struct {
	hostService interfaces.HostService
}

// NewHostHandler creates a new instance of HostHandler.
func NewHostHandler(hs interfaces.HostService) *HostHandler {
	return &HostHandler{
		hostService: hs,
	}
}

// RegisterRoutes registers the HTTP routes for host-related actions.
func (h *HostHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/hosts", h.CreateHost)
	mux.HandleFunc("GET /api/v1/hosts", h.ListHosts)
	mux.HandleFunc("GET /api/v1/hosts/{hostID}", h.GetHostByID)
	mux.HandleFunc("PUT /api/v1/hosts/{hostID}", h.UpdateHost)
	mux.HandleFunc("DELETE /api/v1/hosts/{hostID}", h.DeleteHost) // Soft delete.
	mux.HandleFunc("PATCH /api/v1/hosts/{hostID}/status", h.UpdateHostOnlineStatus)
}

// CreateHost handles the request to create a new host.
func (h *HostHandler) CreateHost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.CreateHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "CreateHost: failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// TODO: Implement request DTO validation.

	// Map the handler DTO to the service layer input DTO.
	serviceInput := serviceDTO.CreateHostInput{
		HostName:     req.HostName,
		Country:      req.Country,
		City:         req.City,
		Address:      req.Address,
		Port:         req.Port,
		Protocol:     req.Protocol,
		Network:      req.Network,
		PublicKey:    req.PublicKey,
		Flow:         req.Flow,
		RSID:         req.RSID,
		SecurityType: req.SecurityType,
		SNI:          req.SNI,
		Fingerprint:  req.Fingerprint,
		IsPrivate:    req.IsPrivate,
		Region:       req.Region,
		Provider:     req.Provider,
	}

	host, err := h.hostService.AddHost(ctx, serviceInput)
	if err != nil {
		slog.ErrorContext(ctx, "CreateHost: failed to add host via service", "error", err, "address", req.Address)
		if strings.Contains(err.Error(), "already exists") {
			respondWithError(w, http.StatusConflict, err.Error())
		} else if strings.Contains(err.Error(), "cannot be empty") {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to add host.")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, toHostResponse(host))
}

// GetHostByID handles the request to retrieve a host by its ID.
func (h *HostHandler) GetHostByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hostIDStr := r.PathValue("hostID")
	hostID, err := parseUint(hostIDStr)
	if err != nil {
		slog.WarnContext(ctx, "GetHostByID: invalid host ID format in path", "hostID_str", hostIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid host ID format provided.")
		return
	}

	host, err := h.hostService.GetHostByID(ctx, hostID)
	if err != nil {
		slog.ErrorContext(ctx, "GetHostByID: failed to get host from service", "error", err, "hostID", hostID)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Host not found.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve host.")
		}
		return
	}
	respondWithJSON(w, http.StatusOK, toHostResponse(host))
}

// ListHosts handles the request to retrieve a list of hosts with filtering and pagination.
func (h *HostHandler) ListHosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.InfoContext(ctx, "ListHosts: received request to list hosts")
	query := r.URL.Query()

	// Parse pagination parameters.
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1 // Default to page 1.
	}

	pageSize, err := strconv.Atoi(query.Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10 // Default page size.
	}
	if pageSize > 100 { // Max page size limit.
		pageSize = 100
	}

	// Prepare service parameters for listing hosts.
	serviceParams := serviceDTO.ListHostsServiceParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    query.Get("sort_by"),    // E.g., "created_at"
		SortOrder: query.Get("sort_order"), // E.g., "asc" or "desc"
	}

	// Apply optional filters from query parameters.
	if country := query.Get("country"); country != "" {
		serviceParams.Country = &country
	}
	if city := query.Get("city"); city != "" {
		serviceParams.City = &city
	}
	if protocol := query.Get("protocol"); protocol != "" {
		serviceParams.Protocol = &protocol
	}
	if hostName := query.Get("host_name"); hostName != "" {
		serviceParams.HostName = &hostName
	}
	if address := query.Get("address"); address != "" {
		serviceParams.Address = &address
	}
	if network := query.Get("network"); network != "" {
		serviceParams.Network = &network
	}
	if statusStr := query.Get("status"); statusStr != "" {
		status := customTypes.HostStatus(statusStr)
		if status.IsValid() {
			serviceParams.Status = &status
		} else {
			slog.WarnContext(ctx, "ListHosts: invalid 'status' query parameter provided", "status_param", statusStr)
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid 'status' query parameter: %s", statusStr))
			return
		}
	}
	if isOnlineStr := query.Get("is_online"); isOnlineStr != "" {
		isOnline, err := strconv.ParseBool(isOnlineStr)
		if err == nil {
			serviceParams.IsOnline = &isOnline
		} else {
			slog.WarnContext(ctx, "ListHosts: invalid 'is_online' query parameter", "is_online_param", isOnlineStr, "error", err)
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid 'is_online' query parameter (must be true or false): %s", isOnlineStr))
			return
		}
	}
	if isPrivateStr := query.Get("is_private"); isPrivateStr != "" {
		isPrivate, err := strconv.ParseBool(isPrivateStr)
		if err == nil {
			serviceParams.IsPrivate = &isPrivate
		} else {
			slog.WarnContext(ctx, "ListHosts: invalid 'is_private' query parameter", "is_private_param", isPrivateStr, "error", err)
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid 'is_private' query parameter (must be true or false): %s", isPrivateStr))
			return
		}
	}

	hostsModels, totalItems, err := h.hostService.ListHosts(ctx, serviceParams)
	if err != nil {
		slog.ErrorContext(ctx, "ListHosts: failed to retrieve hosts from service", "error", err, "params", serviceParams)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve hosts list.")
		return
	}

	hostResponses := make([]dto.HostResponse, len(hostsModels))
	for i, hModel := range hostsModels {
		hostResponses[i] = toHostResponse(&hModel)
	}

	totalPages := 0
	if totalItems > 0 && pageSize > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}
	// If requested page is out of bounds but there are items, return an empty list for that page.
	if page > totalPages && totalPages > 0 {
		hostResponses = []dto.HostResponse{}
		slog.WarnContext(ctx, "ListHosts: requested page is out of bounds", "requested_page", page, "total_pages", totalPages)
	}

	response := dto.PaginatedHostsResponse{
		Hosts:       hostResponses,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}
	slog.InfoContext(ctx, "ListHosts: successfully listed hosts", "count_in_page", len(hostResponses), "total_items", totalItems, "current_page", page)
	respondWithJSON(w, http.StatusOK, response)
}

// UpdateHost handles the request to update an existing host.
func (h *HostHandler) UpdateHost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hostIDStr := r.PathValue("hostID")
	hostID, err := parseUint(hostIDStr)
	if err != nil {
		slog.WarnContext(ctx, "UpdateHost: invalid host ID format in path", "hostID_str", hostIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid host ID format provided.")
		return
	}

	var req dto.UpdateHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "UpdateHost: failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// TODO: Implement request DTO validation.

	serviceInput := serviceDTO.UpdateHostInput{
		HostName:     req.HostName,
		Country:      req.Country,
		City:         req.City,
		Address:      req.Address,
		Port:         req.Port,
		Protocol:     req.Protocol,
		Network:      req.Network,
		PublicKey:    req.PublicKey,
		Flow:         req.Flow,
		RSID:         req.RSID,
		SecurityType: req.SecurityType,
		SNI:          req.SNI,
		Fingerprint:  req.Fingerprint,
		IsPrivate:    req.IsPrivate,
		Region:       req.Region,
		Provider:     req.Provider,
	}

	updatedHost, err := h.hostService.UpdateHost(ctx, hostID, serviceInput)
	if err != nil {
		slog.ErrorContext(ctx, "UpdateHost: failed to update host via service", "error", err, "hostID", hostID)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Host not found.")
		} else if strings.Contains(err.Error(), "uniqueness constraint") || strings.Contains(err.Error(), "already exists") {
			respondWithError(w, http.StatusConflict, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to update host.")
		}
		return
	}
	respondWithJSON(w, http.StatusOK, toHostResponse(updatedHost))
}

// DeleteHost handles the request to (soft) delete a host.
func (h *HostHandler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hostIDStr := r.PathValue("hostID")
	hostID, err := parseUint(hostIDStr)
	if err != nil {
		slog.WarnContext(ctx, "DeleteHost: invalid host ID format in path", "hostID_str", hostIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid host ID format provided.")
		return
	}

	if err := h.hostService.RemoveHost(ctx, hostID); err != nil {
		slog.ErrorContext(ctx, "DeleteHost: failed to remove host via service", "error", err, "hostID", hostID)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Host not found.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to remove host.")
		}
		return
	}
	slog.InfoContext(ctx, "DeleteHost: host deleted successfully", "hostID", hostID)
	w.WriteHeader(http.StatusNoContent)
}

// UpdateHostOnlineStatus handles the request to update a host's online status and general status.
func (h *HostHandler) UpdateHostOnlineStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hostIDStr := r.PathValue("hostID")
	hostID, err := parseUint(hostIDStr)
	if err != nil {
		slog.WarnContext(ctx, "UpdateHostOnlineStatus: invalid host ID format in path", "hostID_str", hostIDStr, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid host ID format provided.")
		return
	}

	var req dto.UpdateHostStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "UpdateHostOnlineStatus: failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Validate the HostStatus from the request.
	if !req.Status.IsValid() {
		slog.WarnContext(ctx, "UpdateHostOnlineStatus: invalid status value provided in request", "status_value", req.Status)
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid status value provided: %s", req.Status))
		return
	}

	serviceInput := serviceDTO.UpdateHostStatusInput{
		IsOnline: req.IsOnline,
		Status:   req.Status,
	}

	updatedHost, err := h.hostService.UpdateHostOnlineStatus(ctx, hostID, serviceInput)
	if err != nil {
		slog.ErrorContext(ctx, "UpdateHostOnlineStatus: failed to update host status via service", "error", err, "hostID", hostID)
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Host not found.")
		} else if strings.Contains(err.Error(), "invalid host status") { // Specific error from service.
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to update host status.")
		}
		return
	}
	slog.InfoContext(ctx, "UpdateHostOnlineStatus: host status updated successfully", "hostID", hostID, "new_is_online", updatedHost.IsOnline, "new_status", updatedHost.Status)
	respondWithJSON(w, http.StatusOK, toHostResponse(updatedHost))
}
