package dto

import (
	"bitback/internal/models/customTypes"
	"github.com/google/uuid"
	"time"
)

// CreateSubscriptionRequest defines the request body for creating a new subscription.
// The UserID in the path is used to identify the user for whom the subscription is created.
// If UserID is also included in the request body, it should match the path parameter or be validated
// to ensure the authenticated user has permission to create a subscription for the target UserID.
type CreateSubscriptionRequest struct {
	UserID        string                   `json:"user_id" validate:"required,uuid"` // UserID as a string; requires parsing and validation against path UserID.
	PlanName      string                   `json:"plan_name" validate:"required"`
	DurationUnit  customTypes.DurationUnit `json:"duration_unit" validate:"required"`
	DurationValue int                      `json:"duration_value" validate:"required,gt=0"`
	StartDate     time.Time                `json:"start_date" validate:"required"`                  // Consider adding validation to ensure the date is not in the past.
	Price         *float64                 `json:"price,omitempty" validate:"omitempty,gte=0"`      // Optional: Price of the subscription.
	Currency      *string                  `json:"currency,omitempty" validate:"omitempty,iso4217"` // Optional: ISO 4217 currency code.
	PaymentStatus string                   `json:"payment_status" validate:"required"`              // E.g., "pending", "paid", "failed".
	AutoRenew     bool                     `json:"auto_renew"`                                      // Flag for auto-renewal.
}

// UpdateSubscriptionPaymentRequest defines the request body for updating a subscription's payment status.
type UpdateSubscriptionPaymentRequest struct {
	PaymentStatus string `json:"payment_status" validate:"required"` // The new payment status.
}

// SetSubscriptionAutoRenewRequest defines the request body for enabling or disabling auto-renewal for a subscription.
type SetSubscriptionAutoRenewRequest struct {
	AutoRenew bool `json:"auto_renew"` // The desired auto-renewal state.
}

// SubscriptionResponse defines the standard API response for a single subscription.
type SubscriptionResponse struct {
	ID            uuid.UUID                `json:"id"`
	UserID        uuid.UUID                `json:"user_id"`
	PlanName      string                   `json:"plan_name"`
	DurationUnit  customTypes.DurationUnit `json:"duration_unit"`
	DurationValue int                      `json:"duration_value"`
	StartDate     time.Time                `json:"start_date"`
	EndDate       time.Time                `json:"end_date"`
	IsActive      bool                     `json:"is_active"`
	Price         *float64                 `json:"price,omitempty"`
	Currency      *string                  `json:"currency,omitempty"`
	PaymentStatus string                   `json:"payment_status"`
	AutoRenew     bool                     `json:"auto_renew"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

// PaginatedSubscriptionsResponse defines the structure for a paginated list of subscriptions.
type PaginatedSubscriptionsResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"` // Slice of subscription responses for the current page.
	TotalItems    int64                  `json:"total_items"`   // Total number of subscriptions matching the query.
	TotalPages    int                    `json:"total_pages"`   // Total number of pages available.
	CurrentPage   int                    `json:"current_page"`  // The current page number.
	PageSize      int                    `json:"page_size"`     // The number of items per page.
}

// ExpiringSubscriptionItemResponse DTO for an item in the list of expiring subscriptions within a report.
type ExpiringSubscriptionItemResponse struct {
	SubscriptionID uuid.UUID                `json:"subscription_id"` // ID of the expiring subscription.
	PlanName       string                   `json:"plan_name"`       // Name of the plan.
	EndDate        time.Time                `json:"end_date"`        // Date when the subscription expires.
	DurationUnit   customTypes.DurationUnit `json:"duration_unit"`   // Duration unit of the plan.
	DurationValue  int                      `json:"duration_value"`  // Duration value of the plan.
	AutoRenew      bool                     `json:"auto_renew"`      // Current auto-renewal status.
}

// UserWithExpiringSubscriptionsResponse DTO for a user along with their list of expiring subscriptions.
// Used in reports for users with subscriptions nearing expiration.
type UserWithExpiringSubscriptionsResponse struct {
	User                  UserResponse                       `json:"user"`                   // User details, using the existing UserResponse DTO.
	ExpiringSubscriptions []ExpiringSubscriptionItemResponse `json:"expiring_subscriptions"` // List of the user's expiring subscriptions.
}

// PaginatedUserExpiringSubscriptionsResponse DTO for a paginated report of users and their expiring subscriptions.
type PaginatedUserExpiringSubscriptionsResponse struct {
	Data        []UserWithExpiringSubscriptionsResponse `json:"data"`         // The list of users with their expiring subscriptions for the current page.
	TotalItems  int64                                   `json:"total_items"`  // Total number of expiring *subscriptions* across all users (or users, depending on report definition).
	CurrentPage int                                     `json:"current_page"` // The current page number of the report.
	PageSize    int                                     `json:"page_size"`    // The number of items (users with subscriptions) per page.
	TotalPages  int                                     `json:"total_pages"`  // Total number of pages in the report.
}
