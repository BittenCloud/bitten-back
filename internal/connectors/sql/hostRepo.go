package sql

import (
	"bitback/internal/interfaces"
	"bitback/internal/models"
	"bitback/internal/models/customTypes"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// hostRepository implements the interfaces.HostRepository for interacting with host data in a SQL database.
type hostRepository struct {
	db *gorm.DB
}

// NewHostRepository creates a new instance of hostRepository.
func NewHostRepository(sqlDB interfaces.SQLDatabase) interfaces.HostRepository {
	return &hostRepository{
		db: sqlDB.GetGormClient(),
	}
}

// Create persists a new host record to the database.
func (r *hostRepository) Create(ctx context.Context, host *models.Host) error {
	if host == nil {
		return errors.New("host to create cannot be nil")
	}

	return r.db.WithContext(ctx).Create(host).Error
}

// GetByID retrieves a host by its primary key ID.
// Returns gorm.ErrRecordNotFound if no host is found.
func (r *hostRepository) GetByID(ctx context.Context, id uint) (*models.Host, error) {
	var host models.Host
	if err := r.db.WithContext(ctx).First(&host, id).Error; err != nil {
		return nil, err // err will be gorm.ErrRecordNotFound if the record is not found.
	}
	return &host, nil
}

// GetByAddressPortProtocolNetwork retrieves a host by a unique combination of its address, port, protocol, and network.
// This is typically used to check for the existence of a host before creation.
func (r *hostRepository) GetByAddressPortProtocolNetwork(ctx context.Context, address, port, protocol, network string) (*models.Host, error) {
	var host models.Host
	err := r.db.WithContext(ctx).
		Where("address = ? AND port = ? AND protocol = ? AND network = ?", address, port, protocol, network).
		First(&host).Error
	if err != nil {
		return nil, err // err will be gorm.ErrRecordNotFound if no matching host is found.
	}
	return &host, nil
}

// GetRandomActiveHost retrieves a random, active host from the database.
// It prioritizes hosts that are online (is_online = true) and have a status of 'active'.
// If no such hosts are found, it falls back to any host that is simply 'is_online = true'.
func (r *hostRepository) GetRandomActiveHost(ctx context.Context) (*models.Host, error) {
	var host models.Host
	var count int64

	// Attempt to find hosts that are online AND have status = models.StatusActive.
	err := r.db.WithContext(ctx).Model(&models.Host{}).Where("is_online = ? AND status = ?", true, customTypes.StatusActive).Count(&count).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active hosts with 'active' status: %w", err)
	}

	if count > 0 {
		// Hosts with StatusActive found, select randomly from this pool.
		err = r.db.WithContext(ctx).
			Where("is_online = ? AND status = ?", true, customTypes.StatusActive).
			Order("RANDOM()"). // Use RANDOM() for random selection (PostgreSQL specific).
			First(&host).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get random host with 'active' status: %w", err)
		}
		return &host, nil
	}

	// Fallback: If no hosts with 'active' status were found, try to find any host that is is_online = true.
	err = r.db.WithContext(ctx).Model(&models.Host{}).Where("is_online = ?", true).Count(&count).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count any online hosts (fallback query): %w", err)
	}

	if count == 0 {
		// No hosts are online at all.
		return nil, gorm.ErrRecordNotFound
	}

	// Select randomly from any online hosts.
	err = r.db.WithContext(ctx).
		Where("is_online = ?", true).
		Order("RANDOM()").
		First(&host).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get random online host (fallback query): %w", err)
	}
	return &host, nil
}

// Update saves changes to an existing host record in the database.
// It uses db.Save(), which updates all fields and runs GORM hooks.
func (r *hostRepository) Update(ctx context.Context, host *models.Host) error {
	if host == nil {
		return errors.New("host to update cannot be nil")
	}
	if host.ID == 0 {
		return errors.New("host ID is required for update")
	}
	return r.db.WithContext(ctx).Save(host).Error
}

// Delete performs a soft delete on a host record by setting the DeletedAt timestamp.
// Returns gorm.ErrRecordNotFound if the host to delete is not found.
func (r *hostRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("host ID is required for delete")
	}
	result := r.db.WithContext(ctx).Delete(&models.Host{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Host to delete was not found.
	}
	return nil
}

// List retrieves a list of hosts with filtering, pagination, and sorting.
func (r *hostRepository) List(ctx context.Context, params customTypes.ListHostsParams) ([]models.Host, int64, error) {
	var hosts []models.Host
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.Host{})

	// Apply filters based on provided parameters.
	if params.HostName != nil && *params.HostName != "" {
		query = query.Where("LOWER(host_name) LIKE LOWER(?)", "%"+*params.HostName+"%")
	}
	if params.Address != nil && *params.Address != "" {
		query = query.Where("LOWER(address) LIKE LOWER(?)", "%"+*params.Address+"%")
	}
	if params.Country != nil && *params.Country != "" {
		query = query.Where("LOWER(country) = LOWER(?)", *params.Country)
	}
	if params.City != nil && *params.City != "" {
		query = query.Where("LOWER(city) = LOWER(?)", *params.City)
	}
	if params.Protocol != nil && *params.Protocol != "" {
		query = query.Where("LOWER(protocol) = LOWER(?)", *params.Protocol)
	}
	if params.IsOnline != nil {
		query = query.Where("is_online = ?", *params.IsOnline)
	}
	if params.IsPrivate != nil {
		query = query.Where("is_private = ?", *params.IsPrivate)
	}
	if params.Network != nil && *params.Network != "" {
		query = query.Where("LOWER(network) = LOWER(?)", *params.Network)
	}
	if params.Status != nil {
		statusValue := *params.Status
		if statusValue.IsValid() {
			query = query.Where("status = ?", statusValue)
		}
	}

	// Count the total number of records matching the filters before applying pagination.
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count hosts: %w", err)
	}

	if totalCount == 0 {
		return []models.Host{}, 0, nil // No records match, return an empty list.
	}

	// Apply sorting.
	if params.SortBy != "" {
		order := "ASC"
		if strings.ToLower(params.SortOrder) == "desc" {
			order = "DESC"
		}
		// Whitelist valid sortable columns to prevent SQL injection.
		validSortableColumns := map[string]string{
			"created_at": "created_at",
			"host_name":  "host_name",
			"address":    "address",
			"status":     "status",
			"country":    "country",
			"city":       "city",
		}
		sortByField := strings.ToLower(params.SortBy)
		if dbColumn, ok := validSortableColumns[sortByField]; ok {
			query = query.Order(fmt.Sprintf("%s %s", dbColumn, order))
		} else {
			query = query.Order("created_at DESC") // Default sort order.
		}
	} else {
		query = query.Order("created_at DESC") // Default sort order if SortBy is not specified.
	}

	// Apply pagination (must be after counting and sorting).
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	// A zero offset is valid and means starting from the beginning.
	if params.Offset >= 0 {
		query = query.Offset(params.Offset)
	}

	// Execute the query to retrieve the host data.
	if err := query.Find(&hosts).Error; err != nil {
		return nil, totalCount, fmt.Errorf("failed to list hosts: %w", err)
	}

	return hosts, totalCount, nil
}
