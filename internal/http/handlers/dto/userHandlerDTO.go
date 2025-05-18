package dto

import (
	"github.com/google/uuid"
	"time"
)

// CreateUserRequest defines the request body for creating a new user.
type CreateUserRequest struct {
	Name       string `json:"name" validate:"required,min=2,max=100"` // User's full name.
	Email      string `json:"email" validate:"required,email"`        // User's email address.
	TelegramID int64  `json:"telegram_id,omitempty"`                  // Optional: User's Telegram ID.
}

// UpdateUserRequest defines the request body for updating an existing user.
// Fields are pointers to distinguish between fields not provided for update and
// fields intentionally set to their zero value (e.g., an empty string or 0).
type UpdateUserRequest struct {
	Name       *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"` // New name for the user.
	Email      *string `json:"email,omitempty" validate:"omitempty,email"`        // New email address for the user.
	TelegramID *int64  `json:"telegram_id,omitempty"`                             // New Telegram ID for the user.
	IsActive   *bool   `json:"is_active,omitempty"`                               // New active status for the user.
}

// UserResponse defines the standard API response for a single user's details.
type UserResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email,omitempty"`
	TelegramID int64      `json:"telegram_id,omitempty"`
	IsActive   bool       `json:"is_active"`
	Role       string     `json:"role,omitempty"`       // Optional: User's role within the system.
	LastLogin  *time.Time `json:"last_login,omitempty"` // Optional: Timestamp of the user's last login.
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// PaginatedUsersResponse defines the structure for a paginated list of users.
type PaginatedUsersResponse struct {
	Users       []UserResponse `json:"users"`        // Slice of user responses for the current page.
	TotalItems  int64          `json:"total_items"`  // Total number of users matching the query.
	TotalPages  int            `json:"total_pages"`  // Total number of pages available.
	CurrentPage int            `json:"current_page"` // The current page number.
	PageSize    int            `json:"page_size"`    // The number of items per page.
}
