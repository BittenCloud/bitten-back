package services

import (
	"bitback/internal/interfaces"
	"bitback/internal/models"
	"bitback/internal/models/customTypes"
	"bitback/internal/services/dto"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log/slog"
	"strings"
	"time"
)

type hostService struct {
	hostRepo interfaces.HostRepository
}

// NewHostService creates a new instance of hostService.
func NewHostService(hr interfaces.HostRepository) interfaces.HostService {
	return &hostService{
		hostRepo: hr,
	}
}

// AddHost handles the logic for adding a new host.
// It includes input validation, uniqueness checks, and persistence.
func (s *hostService) AddHost(ctx context.Context, input dto.CreateHostInput) (*models.Host, error) {
	slog.InfoContext(ctx, "AddHost: attempting to add new host", "address", input.Address, "port", input.Port, "protocol", input.Protocol)

	// Perform basic input validation.
	if strings.TrimSpace(input.Address) == "" {
		return nil, errors.New("host address cannot be empty")
	}
	if strings.TrimSpace(input.Port) == "" {
		return nil, errors.New("host port cannot be empty")
	}
	if strings.TrimSpace(input.Protocol) == "" {
		return nil, errors.New("host protocol cannot be empty")
	}
	network := input.Network
	if network == "" {
		network = "tcp" // Set an explicit default network type at the service level if necessary.
	}
	// TODO: Implement more comprehensive validation (e.g., IP/domain format, port range, allowed protocols).

	// Verify that a host with the same address, port, protocol, and network does not already exist.
	existingHost, err := s.hostRepo.GetByAddressPortProtocolNetwork(ctx, input.Address, input.Port, input.Protocol, network)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.ErrorContext(ctx, "AddHost: error checking for existing host", "address", input.Address, "error", err)
		return nil, fmt.Errorf("could not verify host uniqueness: %w", err)
	}
	if existingHost != nil {
		slog.WarnContext(ctx, "AddHost: host already exists", "address", input.Address, "port", input.Port, "protocol", input.Protocol, "network", network, "existingID", existingHost.ID)
		return nil, fmt.Errorf("host with address '%s', port '%s', protocol '%s', and network '%s' already exists", input.Address, input.Port, input.Protocol, network)
	}

	// Prepare the Host model for creation.
	host := &models.Host{
		HostName:     input.HostName,
		Country:      input.Country,
		City:         input.City,
		Address:      input.Address,
		Port:         input.Port,
		Protocol:     input.Protocol,
		Network:      network,
		PublicKey:    input.PublicKey,
		Flow:         input.Flow,
		RSID:         input.RSID,
		SecurityType: input.SecurityType,
		SNI:          input.SNI,
		Fingerprint:  input.Fingerprint,
		IsPrivate:    input.IsPrivate,
		IsOnline:     false, // New hosts are considered offline by default until a status check.
		Status:       customTypes.StatusUnknown,
		Region:       input.Region,
		Provider:     input.Provider,
	}

	// Persist the new host to the repository.
	if err := s.hostRepo.Create(ctx, host); err != nil {
		slog.ErrorContext(ctx, "AddHost: failed to create host in repository", "address", input.Address, "error", err)
		return nil, fmt.Errorf("could not add host: %w", err)
	}

	slog.InfoContext(ctx, "AddHost: host added successfully", "hostID", host.ID, "address", host.Address)
	return host, nil
}

// GetHostByID retrieves a host by its unique ID.
func (s *hostService) GetHostByID(ctx context.Context, hostID uint) (*models.Host, error) {
	slog.InfoContext(ctx, "GetHostByID: attempting to get host", "hostID", hostID)
	host, err := s.hostRepo.GetByID(ctx, hostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GetHostByID: host not found", "hostID", hostID)
			return nil, fmt.Errorf("host with ID %d not found: %w", hostID, err)
		}
		slog.ErrorContext(ctx, "GetHostByID: failed to get host from repository", "hostID", hostID, "error", err)
		return nil, fmt.Errorf("could not retrieve host: %w", err)
	}
	slog.InfoContext(ctx, "GetHostByID: host retrieved successfully", "hostID", host.ID)
	return host, nil
}

// UpdateHost applies updates to an existing host's data.
func (s *hostService) UpdateHost(ctx context.Context, hostID uint, input dto.UpdateHostInput) (*models.Host, error) {
	slog.InfoContext(ctx, "UpdateHost: attempting to update host", "hostID", hostID)

	host, err := s.hostRepo.GetByID(ctx, hostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "UpdateHost: host to update not found", "hostID", hostID)
			return nil, fmt.Errorf("host with ID %d not found for update: %w", hostID, err)
		}
		slog.ErrorContext(ctx, "UpdateHost: failed to retrieve host for update", "hostID", hostID, "error", err)
		return nil, fmt.Errorf("could not retrieve host for update: %w", err)
	}

	changesMade := false
	if input.HostName != nil && *input.HostName != host.HostName {
		host.HostName = *input.HostName
		changesMade = true
	}
	if input.Country != nil && *input.Country != host.Country {
		host.Country = *input.Country
		changesMade = true
	}
	if input.City != nil && *input.City != host.City {
		host.City = *input.City
		changesMade = true
	}
	if input.Flow != nil && *input.Flow != host.Flow {
		host.Flow = *input.Flow
		changesMade = true
	}
	if input.RSID != nil && *input.RSID != host.RSID {
		host.RSID = *input.RSID
		changesMade = true
	}
	if input.SecurityType != nil && *input.SecurityType != host.SecurityType {
		host.SecurityType = *input.SecurityType
		changesMade = true
	}
	if input.SNI != nil && *input.SNI != host.SNI {
		host.SNI = *input.SNI
		changesMade = true
	}
	if input.Fingerprint != nil && *input.Fingerprint != host.Fingerprint {
		host.Fingerprint = *input.Fingerprint
		changesMade = true
	}
	if input.IsPrivate != nil && *input.IsPrivate != host.IsPrivate {
		host.IsPrivate = *input.IsPrivate
		changesMade = true
	}
	if input.PublicKey != nil && *input.PublicKey != host.PublicKey {
		host.PublicKey = *input.PublicKey
		changesMade = true
	}
	if input.Region != nil && *input.Region != host.Region {
		host.Region = *input.Region
		changesMade = true
	}
	if input.Provider != nil && *input.Provider != host.Provider {
		host.Provider = *input.Provider
		changesMade = true
	}
	if input.Network != nil && *input.Network != host.Network {
		// TODO: If Address, Port, Protocol, or Network fields are changed,
		host.Network = *input.Network
		changesMade = true
	}

	if !changesMade {
		slog.InfoContext(ctx, "UpdateHost: no actual changes detected for host", "hostID", hostID)
		return host, nil
	}

	if err := s.hostRepo.Update(ctx, host); err != nil {
		slog.ErrorContext(ctx, "UpdateHost: failed to update host in repository", "hostID", hostID, "error", err)
		return nil, fmt.Errorf("could not save host updates: %w", err)
	}

	slog.InfoContext(ctx, "UpdateHost: host updated successfully", "hostID", host.ID)
	return host, nil
}

// RemoveHost performs a soft delete on a host.
// The repository handles the existence check and returns gorm.ErrRecordNotFound if applicable.
func (s *hostService) RemoveHost(ctx context.Context, hostID uint) error {
	slog.InfoContext(ctx, "RemoveHost: attempting to remove host", "hostID", hostID)
	if err := s.hostRepo.Delete(ctx, hostID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "RemoveHost: host to remove not found", "hostID", hostID)
			return fmt.Errorf("host with ID %d not found for removal: %w", hostID, err)
		}
		slog.ErrorContext(ctx, "RemoveHost: failed to remove host from repository", "hostID", hostID, "error", err)
		return fmt.Errorf("could not remove host: %w", err)
	}
	slog.InfoContext(ctx, "RemoveHost: host removed successfully", "hostID", hostID)
	return nil
}

// ListHosts retrieves a paginated and filtered list of hosts.
func (s *hostService) ListHosts(ctx context.Context, params dto.ListHostsServiceParams) ([]models.Host, int64, error) {
	slog.InfoContext(ctx, "ListHosts: attempting to list hosts", "params", fmt.Sprintf("%+v", params))

	// Convert service-layer DTO parameters to repository-layer parameters.
	repoParams := customTypes.ListHostsParams{
		Country:   params.Country,
		City:      params.City,
		Protocol:  params.Protocol,
		Network:   params.Network,
		IsOnline:  params.IsOnline,
		IsPrivate: params.IsPrivate,
		Status:    params.Status,
		HostName:  params.HostName,
		Address:   params.Address,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}

	// Validate and set default values for pagination.
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = defaultPageSize
	}
	if params.PageSize > maxPageSize {
		params.PageSize = maxPageSize
	}
	repoParams.Offset = (params.Page - 1) * params.PageSize
	repoParams.Limit = params.PageSize

	hosts, totalCount, err := s.hostRepo.List(ctx, repoParams)
	if err != nil {
		slog.ErrorContext(ctx, "ListHosts: failed to list hosts from repository", "error", err)
		return nil, 0, fmt.Errorf("could not retrieve hosts list: %w", err)
	}
	slog.InfoContext(ctx, "ListHosts: hosts listed successfully", "count", len(hosts), "totalCount", totalCount)
	return hosts, totalCount, nil
}

// UpdateHostOnlineStatus updates a host's online status, typically called by a monitoring system.
// This includes IsOnline, Status, and LastCheckedAt fields.
func (s *hostService) UpdateHostOnlineStatus(ctx context.Context, hostID uint, input dto.UpdateHostStatusInput) (*models.Host, error) {
	slog.InfoContext(ctx, "UpdateHostOnlineStatus: attempting to update host status", "hostID", hostID, "isOnline", input.IsOnline, "newStatus", input.Status)

	host, err := s.hostRepo.GetByID(ctx, hostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "UpdateHostOnlineStatus: host not found", "hostID", hostID)
			return nil, fmt.Errorf("host with ID %d not found: %w", hostID, err)
		}
		slog.ErrorContext(ctx, "UpdateHostOnlineStatus: failed to retrieve host", "hostID", hostID, "error", err)
		return nil, fmt.Errorf("could not retrieve host: %w", err)
	}

	if !input.Status.IsValid() {
		slog.WarnContext(ctx, "UpdateHostOnlineStatus: invalid status provided", "hostID", hostID, "status", input.Status)
		return nil, fmt.Errorf("invalid host status provided: %s", input.Status)
	}

	host.IsOnline = input.IsOnline
	host.Status = input.Status
	now := time.Now()
	host.LastCheckedAt = &now

	if err := s.hostRepo.Update(ctx, host); err != nil {
		slog.ErrorContext(ctx, "UpdateHostOnlineStatus: failed to update host status in repository", "hostID", hostID, "error", err)
		return nil, fmt.Errorf("could not save host status update: %w", err)
	}
	slog.InfoContext(ctx, "UpdateHostOnlineStatus: host status updated successfully", "hostID", host.ID)
	return host, nil
}
