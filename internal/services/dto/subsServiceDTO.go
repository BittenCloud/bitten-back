package dto

import (
	"bitback/internal/models"
	"bitback/internal/models/customTypes"
	"github.com/google/uuid"
	"time"
)

// CreateSubscriptionInput defines the data required to create a new subscription at the service layer.
type CreateSubscriptionInput struct {
	UserID        uuid.UUID                // The ID of the user for whom the subscription is being created.
	PlanName      string                   // The name of the subscription plan.
	DurationUnit  customTypes.DurationUnit // The unit of measurement for the subscription duration (e.g., day, month, year).
	DurationValue int                      // The value of the subscription duration.
	StartDate     time.Time                // The start date of the subscription can be in the future.
	Price         *float64                 // Optional: The price of the subscription.
	Currency      *string                  // Optional: The currency for the price (e.g., "USD").
	PaymentStatus string                   // The status of the payment (e.g., "paid", "pending", "failed").
	AutoRenew     bool                     // Flag indicating if the subscription should auto-renew.
}

// UpdateSubscriptionInput defines the data that can be updated for an existing subscription.
// Using pointers allows distinguishing between a field not being provided and a field being set to its zero value.
type UpdateSubscriptionInput struct {
	AutoRenew     *bool   // To change the auto-renewal flag.
	PaymentStatus *string // To update the payment status.
	// Fields like IsActive and EndDate are typically managed by system logic rather than direct client updates.
}

// ExpiringSubscriptionInfo contains concise information about a subscription that is nearing its expiration date.
type ExpiringSubscriptionInfo struct {
	ID            uuid.UUID                `json:"id"` // The ID of the subscription itself.
	PlanName      string                   `json:"plan_name"`
	EndDate       time.Time                `json:"end_date"`
	DurationUnit  customTypes.DurationUnit `json:"duration_unit"`
	DurationValue int                      `json:"duration_value"`
	AutoRenew     bool                     `json:"auto_renew"`
}

// UserWithExpiringSubscriptions groups a user with their list of subscriptions that are about to expire.
// This is used for reporting purposes.
type UserWithExpiringSubscriptions struct {
	User                  models.User
	ExpiringSubscriptions []ExpiringSubscriptionInfo
}
