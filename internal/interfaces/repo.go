package interfaces

import (
	"bitback/internal/models"
	"bitback/internal/models/customTypes"
	"context"
	"github.com/google/uuid"
	"time"
)

// UserRepository defines methods for interacting with the user data storage.
type UserRepository interface {
	// Create persists a new user to the storage.
	Create(ctx context.Context, user *models.User) error

	// GetByID retrieves a user by their unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// GetByIDs retrieves a list of users by their unique UUIDs.
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Update persists changes to an existing user in the storage.
	Update(ctx context.Context, user *models.User) error

	// Delete performs a soft delete on a user identified by their UUID.
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a paginated list of users.
	// It returns the list of users, the total count of users matching the criteria, and any error.
	List(ctx context.Context, offset, limit int) ([]models.User, int64, error)
}

// SubscriptionRepository defines methods for interacting with the subscription data storage.
type SubscriptionRepository interface {
	// Create persists a new subscription to the storage.
	Create(ctx context.Context, subscription *models.Subscription) error

	// GetByID retrieves a subscription by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)

	// Update persists changes to an existing subscription in the storage.
	Update(ctx context.Context, subscription *models.Subscription) error

	// Delete performs a soft delete on a subscription identified by its ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByUserID retrieves a paginated list of subscriptions for a specific user.
	// It returns the list of subscriptions, the total count, and any error.
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) (subscriptions []models.Subscription, totalCount int64, err error)

	// ListExpiringSoon retrieves a paginated list of active subscriptions that are due to expire within a given time window.
	// It returns the list of subscriptions, the total count, and any error.
	ListExpiringSoon(ctx context.Context, thresholdDateFrom time.Time, thresholdDateTo time.Time, offset, limit int) (subscriptions []models.Subscription, totalCount int64, err error)

	// ListActiveByPlanName retrieves a paginated list of active subscriptions matching a specific plan name.
	// It returns the list of subscriptions, the total count, and any error.
	ListActiveByPlanName(ctx context.Context, planName string, offset, limit int) (subscriptions []models.Subscription, totalCount int64, err error)
}

// HostRepository defines methods for interacting with the host data storage.
type HostRepository interface {
	// Create persists a new host to the storage.
	Create(ctx context.Context, host *models.Host) error

	// GetByID retrieves a host by its unique ID.
	GetByID(ctx context.Context, id uint) (*models.Host, error)

	// GetByAddressPortProtocolNetwork retrieves a host by its address, port, protocol, and network combination.
	// This is often used to check for uniqueness.
	GetByAddressPortProtocolNetwork(ctx context.Context, address, port, protocol, network string) (*models.Host, error)

	// GetRandomActiveHost retrieves a random, currently active host from the storage.
	GetRandomActiveHost(ctx context.Context) (*models.Host, error)

	// Update persists changes to an existing host in the storage.
	Update(ctx context.Context, host *models.Host) error

	// Delete performs a soft delete on a host identified by its ID.
	Delete(ctx context.Context, id uint) error

	// List retrieves a list of hosts based on specified filter parameters, with pagination.
	// It returns the list of hosts, the total count matching the criteria, and any error.
	List(ctx context.Context, params customTypes.ListHostsParams) (hosts []models.Host, totalCount int64, err error)
}
