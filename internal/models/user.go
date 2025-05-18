package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// User defines the database model for a user.
type User struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`   // Unique identifier for the user.
	Name       string         `json:"name" gorm:"not null"`              // Name of the user.
	Email      string         `json:"email"`                             // Email address of the user.
	TelegramID int64          `json:"telegram_id,omitempty"`             // Optional: User's Telegram ID.
	IsActive   bool           `json:"is_active" gorm:"default:true"`     // Indicates if the user account is active; defaults to true.
	LastLogin  *time.Time     `json:"last_login,omitempty"`              // Optional: Timestamp of the user's last login.
	CreatedAt  time.Time      `json:"created_at"`                        // Timestamp of creation.
	UpdatedAt  time.Time      `json:"updated_at"`                        // Timestamp of the last update.
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // Timestamp for soft deletion.
}

// BeforeCreate is a GORM hook that runs before a new user record is created.
// It generates a new UUID (version 7) for the user's ID.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID, err = uuid.NewV7()
	return err
}
