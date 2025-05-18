package dto

// CreateUserInput defines the data required for creating a user at the service layer.
type CreateUserInput struct {
	Name       string // The name of the user.
	Email      string // The email address of the user.
	TelegramID int64  // Optional: The user's Telegram ID.
}

// UpdateUserInput defines the data for updating an existing user at the service layer.
// Fields are pointers to distinguish between zero values (e.g., empty string or 0)
// and fields that were not provided for update.
type UpdateUserInput struct {
	Name       *string // The new name of the user.
	Email      *string // The new email address of the user.
	TelegramID *int64  // The new Telegram ID of the user.
	IsActive   *bool   // The new active status of the user.
}
