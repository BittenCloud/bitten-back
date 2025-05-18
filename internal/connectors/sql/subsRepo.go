package sql

import (
	"bitback/internal/interfaces"
	"bitback/internal/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// subscriptionRepository implements the interfaces.SubscriptionRepository for interacting with subscription data in a SQL database.
type subscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new instance of subscriptionRepository.
func NewSubscriptionRepository(sqlDB interfaces.SQLDatabase) interfaces.SubscriptionRepository {
	return &subscriptionRepository{
		db: sqlDB.GetGormClient(),
	}
}

// Create persists a new subscription record to the database.
// Fields like EndDate and IsActive should be determined by the service layer before calling Create.
func (r *subscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	if subscription == nil {
		return errors.New("subscription to create cannot be nil")
	}
	return r.db.WithContext(ctx).Create(subscription).Error
}

// GetByID retrieves a subscription by its primary key (UUID).
// Returns gorm.ErrRecordNotFound if no subscription is found.
func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := r.db.WithContext(ctx).First(&subscription, "id = ?", id).Error; err != nil {
		return nil, err // err will be gorm.ErrRecordNotFound if the record is not found.
	}
	return &subscription, nil
}

// Update saves changes to an existing subscription record in the database.
// It uses db.Save(), which updates all fields and runs GORM hooks.
func (r *subscriptionRepository) Update(ctx context.Context, subscription *models.Subscription) error {
	if subscription == nil {
		return errors.New("subscription to update cannot be nil")
	}
	if subscription.ID == uuid.Nil {
		return errors.New("subscription ID is required for update")
	}
	return r.db.WithContext(ctx).Save(subscription).Error
}

// Delete performs a soft delete on a subscription record by its ID (uint).
// Returns gorm.ErrRecordNotFound if the subscription to delete is not found.
func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("subscription ID for delete cannot be zero")
	}
	result := r.db.WithContext(ctx).Delete(&models.Subscription{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Subscription to delete was not found.
	}
	return nil
}

// ListByUserID retrieves a paginated list of subscriptions for a specific user.
// Subscriptions can be ordered, for example, by creation date or end date.
func (r *subscriptionRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]models.Subscription, int64, error) {
	var subscriptions []models.Subscription
	var totalCount int64

	// Query to count the total number of subscriptions for the user.
	countQuery := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("user_id = ?", userID)
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user subscriptions: %w", err)
	}

	if totalCount == 0 {
		return []models.Subscription{}, 0, nil // No subscriptions found for this user.
	}

	// Query to retrieve a slice of subscriptions with pagination and ordering.
	// Example orders: "created_at DESC" (newest first) or "end_date DESC".
	listQuery := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	if err := listQuery.Find(&subscriptions).Error; err != nil {
		return nil, totalCount, fmt.Errorf("failed to list user subscriptions: %w", err)
	}

	return subscriptions, totalCount, nil
}

// ListExpiringSoon retrieves a paginated list of active subscriptions that are due to expire within a specified time window.
// Subscriptions are ordered by their end date in ascending order (soonest expiring first).
func (r *subscriptionRepository) ListExpiringSoon(ctx context.Context, thresholdDateFrom time.Time, thresholdDateTo time.Time, offset, limit int) ([]models.Subscription, int64, error) {
	var subscriptions []models.Subscription
	var totalCount int64

	// Base query for counting and selecting expiring subscriptions.
	baseQuery := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("is_active = ?", true).              // Only include active subscriptions.
		Where("end_date >= ?", thresholdDateFrom). // Subscriptions that have not yet ended (or end today).
		Where("end_date <= ?", thresholdDateTo)    // Subscriptions that end before or on the specified upper threshold date.

	// Count the total number of matching subscriptions.
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count expiring subscriptions: %w", err)
	}

	if totalCount == 0 {
		return []models.Subscription{}, 0, nil // No subscriptions are expiring soon within the criteria.
	}

	// Retrieve the paginated list of expiring subscriptions.
	query := baseQuery.Order("end_date ASC").Offset(offset).Limit(limit)
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, totalCount, fmt.Errorf("failed to list expiring subscriptions: %w", err)
	}
	return subscriptions, totalCount, nil
}

// ListActiveByPlanName retrieves a paginated list of active subscriptions for a specific plan name.
// Subscriptions are ordered by their start date in descending order (newest first).
func (r *subscriptionRepository) ListActiveByPlanName(ctx context.Context, planName string, offset, limit int) ([]models.Subscription, int64, error) {
	var subscriptions []models.Subscription
	var totalCount int64

	// Base query for active subscriptions by plan name.
	baseQuery := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("is_active = ?", true).
		Where("plan_name = ?", planName)

	// Count the total number of matching subscriptions.
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count active subscriptions for plan name '%s': %w", planName, err)
	}

	if totalCount == 0 {
		return []models.Subscription{}, 0, nil // No active subscriptions found for this plan name.
	}

	// Retrieve the paginated list.
	query := baseQuery.Order("start_date DESC").Offset(offset).Limit(limit)
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, totalCount, fmt.Errorf("failed to list active subscriptions for plan name '%s': %w", planName, err)
	}
	return subscriptions, totalCount, nil
}
