package models

import (
	"bitback/internal/models/customTypes"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// Subscription defines the database model for a user's subscription plan.
type Subscription struct {
	ID            uuid.UUID                `gorm:"type:uuid;primary_key" json:"id"`                                           // Unique identifier for the subscription.
	UserID        uuid.UUID                `json:"user_id" gorm:"type:uuid;not null;index"`                                   // Foreign key linking to the User.
	User          User                     `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // Associated User model (ignored in JSON, handled by foreign key).
	PlanName      string                   `json:"plan_name" gorm:"not null"`                                                 // Name of the subscription plan.
	DurationUnit  customTypes.DurationUnit `json:"duration_unit" gorm:"type:varchar(10);not null"`                            // Unit for the duration (e.g., day, month, year).
	DurationValue int                      `json:"duration_value" gorm:"not null"`                                            // Value for the duration in DurationUnit.
	StartDate     time.Time                `json:"start_date" gorm:"not null"`                                                // Date when the subscription starts.
	EndDate       time.Time                `json:"end_date" gorm:"not null"`                                                  // Date when the subscription ends.
	Currency      string                   `json:"currency,omitempty" gorm:"type:varchar(3)"`                                 // Optional: Currency code for the price (e.g., "USD").
	Price         float64                  `json:"price,omitempty"`                                                           // Optional: Price of the subscription.
	IsActive      bool                     `json:"is_active"`                                                                 // Indicates if the subscription is currently active.
	PaymentStatus string                   `json:"payment_status,omitempty" gorm:"type:varchar(20);index"`                    // Status of the payment (e.g., "paid", "pending").
	AutoRenew     bool                     `json:"auto_renew" gorm:"default:false"`                                           // Flag indicating if the subscription should auto-renew; defaults to false.
	CreatedAt     time.Time                `json:"created_at"`                                                                // Timestamp of creation.
	UpdatedAt     time.Time                `json:"updated_at"`                                                                // Timestamp of the last update.
	DeletedAt     gorm.DeletedAt           `gorm:"index" json:"deleted_at,omitempty"`                                         // Timestamp for soft deletion.
}

// BeforeCreate is a GORM hook that runs before a new subscription record is created.
// It generates a new UUID (version 7) for the subscription's ID.
func (s *Subscription) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID, err = uuid.NewV7()
	return err
}
