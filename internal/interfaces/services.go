package interfaces

import (
	"bitback/internal/models"
	serviceDTO "bitback/internal/services/dto"
	"context"
	"github.com/google/uuid"
)

// KeyService defines methods for managing and generating keys.
type KeyService interface {
	// GenerateVlessKeyForUser creates a VLESS key string for a specified user,
	// optionally including remarks for identification.
	GenerateVlessKeyForUser(ctx context.Context, userID uuid.UUID, remarks string) (string, error)
}

// UserService defines the business logic methods for user management.
type UserService interface {
	// RegisterUser creates a new user account.
	RegisterUser(ctx context.Context, input serviceDTO.CreateUserInput) (*models.User, error)

	// GetUser retrieves a user by their unique ID.
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)

	// UpdateUser modifies an existing user's information.
	UpdateUser(ctx context.Context, id uuid.UUID, input serviceDTO.UpdateUserInput) (*models.User, error)

	// DeleteUser performs a soft delete on a user.
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// ListUsers retrieves a paginated list of users.
	// It returns the slice of users, the total count of users, and any error encountered.
	ListUsers(ctx context.Context, page, pageSize int) (users []models.User, totalCount int64, err error)
}

// SubscriptionService defines the business logic methods for managing user subscriptions.
type SubscriptionService interface {
	// CreateSubscription establishes a new subscription for a user based on the provided input.
	CreateSubscription(ctx context.Context, input serviceDTO.CreateSubscriptionInput) (*models.Subscription, error)

	// GetSubscriptionByID retrieves a specific subscription by its ID.
	// The requestingUserID is used for authorization to ensure the user has rights to view it.
	GetSubscriptionByID(ctx context.Context, subscriptionID uuid.UUID, requestingUserID uuid.UUID) (*models.Subscription, error)

	// ListUserSubscriptions retrieves a paginated list of all subscriptions for a given user.
	ListUserSubscriptions(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Subscription, int64, error)

	// GetUsersWithExpiringSubscriptions generates a report of users whose subscriptions are nearing expiration.
	// The report is paginated and includes details of the expiring subscriptions for each user.
	// Returns a slice of UserWithExpiringSubscriptions, the total count of such users (or subscriptions, depending on pagination strategy), and any error.
	GetUsersWithExpiringSubscriptions(ctx context.Context, daysInAdvance int, page, pageSize int) (reportData []serviceDTO.UserWithExpiringSubscriptions, totalCount int64, err error)

	// ListActiveSubscriptionsByPlan retrieves a paginated list of active subscriptions for a specific plan name.
	ListActiveSubscriptionsByPlan(ctx context.Context, planName string, page, pageSize int) (subscriptions []models.Subscription, totalCount int64, err error)

	// CancelSubscription cancels a subscription, which might involve disabling auto-renewal or deactivating it.
	// The requestingUserID is used for authorization.
	CancelSubscription(ctx context.Context, subscriptionID uuid.UUID, requestingUserID uuid.UUID) (*models.Subscription, error)

	// UpdatePaymentStatus updates the payment status of a specific subscription.
	UpdatePaymentStatus(ctx context.Context, subscriptionID uuid.UUID, paymentStatus string) (*models.Subscription, error)

	// SetAutoRenew enables or disables the auto-renewal feature for a subscription.
	// The requestingUserID is used for authorization.
	SetAutoRenew(ctx context.Context, subscriptionID uuid.UUID, requestingUserID uuid.UUID, autoRenew bool) (*models.Subscription, error)
	// CheckUserActiveSubscription(ctx context.Context, userID uuid.UUID, planName *string) (*models.Subscription, error)
}

// HostService defines the business logic methods for managing hosts or servers.
type HostService interface {
	// AddHost adds a new host to the system based on the provided input.
	AddHost(ctx context.Context, input serviceDTO.CreateHostInput) (*models.Host, error)

	// GetHostByID retrieves a host by its unique ID.
	GetHostByID(ctx context.Context, hostID uint) (*models.Host, error)

	// UpdateHost modifies an existing host's information.
	UpdateHost(ctx context.Context, hostID uint, input serviceDTO.UpdateHostInput) (*models.Host, error)

	// RemoveHost performs a soft delete on a host.
	RemoveHost(ctx context.Context, hostID uint) error

	// ListHosts retrieves a paginated and filtered list of hosts.
	// It returns the slice of hosts, the total count of hosts matching the criteria, and any error.
	ListHosts(ctx context.Context, params serviceDTO.ListHostsServiceParams) (hosts []models.Host, totalCount int64, err error)

	// UpdateHostOnlineStatus updates the online status and other related metrics of a host.
	UpdateHostOnlineStatus(ctx context.Context, hostID uint, input serviceDTO.UpdateHostStatusInput) (*models.Host, error)
}
