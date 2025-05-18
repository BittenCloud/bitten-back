package services

import (
	"bitback/internal/interfaces"
	"bitback/internal/models"
	"bitback/internal/services/dto"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userService struct {
	userRepo interfaces.UserRepository
}

// NewUserService creates a new instance of userService.
func NewUserService(userRepo interfaces.UserRepository) interfaces.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// RegisterUser handles the registration of a new user.
// It performs validation and persists the new user to the repository.
func (s *userService) RegisterUser(ctx context.Context, input dto.CreateUserInput) (*models.User, error) {
	slog.InfoContext(ctx, "RegisterUser: attempting to register user", "email", input.Email)

	// Validate input data.
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("user name cannot be empty")
	}

	// Create the user model.
	user := &models.User{
		Name:       input.Name,
		Email:      input.Email,
		TelegramID: input.TelegramID,
	}

	// Persist the user in the repository.
	if err := s.userRepo.Create(ctx, user); err != nil {
		slog.ErrorContext(ctx, "RegisterUser: failed to create user in repository", "email", input.Email, "error", err)
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, fmt.Errorf("failed to create user: a user with the provided details (e.g., email) may already exist: %w", err)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	slog.InfoContext(ctx, "RegisterUser: user registered successfully", "userID", user.ID, "email", user.Email)
	return user, nil
}

// GetUser retrieves a user by their ID.
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	slog.InfoContext(ctx, "GetUser: attempting to get user by ID", "userID", id)
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GetUser: user not found", "userID", id)
			return nil, fmt.Errorf("user with ID '%s' not found: %w", id, err)
		}
		slog.ErrorContext(ctx, "GetUser: failed to get user by ID from repository", "userID", id, "error", err)
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	slog.InfoContext(ctx, "GetUser: user retrieved successfully", "userID", user.ID, "email", user.Email)
	return user, nil
}

// UpdateUser updates an existing user's data.
// It retrieves the current user, applies provided changes, and persists them.
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, input dto.UpdateUserInput) (*models.User, error) {
	slog.InfoContext(ctx, "UpdateUser: attempting to update user", "userID", id)

	// Retrieve the current user to ensure updates are applied to the latest data
	// and that GORM knows which record to update.
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "UpdateUser: user to update not found in repository", "userID", id)
			return nil, fmt.Errorf("user with ID '%s' not found: %w", id, err)
		}
		slog.ErrorContext(ctx, "UpdateUser: failed to retrieve user for update from repository", "userID", id, "error", err)
		return nil, fmt.Errorf("could not retrieve user for update: %w", err)
	}

	changesMade := false

	// Update user's name if provided and different.
	if input.Name != nil {
		trimmedName := strings.TrimSpace(*input.Name)
		if trimmedName == "" {
			slog.WarnContext(ctx, "UpdateUser: attempt to set empty user name", "userID", id)
			return nil, errors.New("user name cannot be empty if provided for update")
		}
		if trimmedName != user.Name {
			user.Name = trimmedName
			changesMade = true
			slog.DebugContext(ctx, "UpdateUser: updating user name", "userID", id, "newName", user.Name)
		}
	}

	// Update user's email if provided and different.
	// Includes a check to ensure the new email isn't already in use by another user.
	if input.Email != nil {
		trimmedEmail := strings.ToLower(strings.TrimSpace(*input.Email))
		if trimmedEmail == "" {
			slog.WarnContext(ctx, "UpdateUser: attempt to set empty user email", "userID", id)
			return nil, errors.New("user email cannot be empty if provided for update")
		}

		if trimmedEmail != user.Email {
			existingUserWithNewEmail, errGetByEmail := s.userRepo.GetByEmail(ctx, trimmedEmail)
			if errGetByEmail == nil && existingUserWithNewEmail != nil && existingUserWithNewEmail.ID != user.ID {
				slog.WarnContext(ctx, "UpdateUser: new email already in use by another user", "userID", id, "newEmail", trimmedEmail, "conflictingUserID", existingUserWithNewEmail.ID)
				return nil, fmt.Errorf("email '%s' is already in use by another user", trimmedEmail)
			}
			// If an error occurred but it's not ErrRecordNotFound, it indicates a DB access issue.
			if errGetByEmail != nil && !errors.Is(errGetByEmail, gorm.ErrRecordNotFound) {
				slog.ErrorContext(ctx, "UpdateUser: error checking new email availability", "userID", id, "newEmail", trimmedEmail, "error", errGetByEmail)
				return nil, fmt.Errorf("could not verify new email availability: %w", errGetByEmail)
			}
			// If the email is available (errGetByEmail == gorm.ErrRecordNotFound), update it.
			user.Email = trimmedEmail
			changesMade = true
			slog.DebugContext(ctx, "UpdateUser: updating user email", "userID", id, "newEmail", user.Email)
		}
	}

	// Update user's Telegram ID if provided and different.
	if input.TelegramID != nil {
		if *input.TelegramID != user.TelegramID {
			user.TelegramID = *input.TelegramID
			changesMade = true
			slog.DebugContext(ctx, "UpdateUser: updating user Telegram ID", "userID", id, "newTelegramID", user.TelegramID)
		}
	}

	// Update user's active status if provided and different.
	if input.IsActive != nil {
		if *input.IsActive != user.IsActive {
			user.IsActive = *input.IsActive
			changesMade = true
			slog.DebugContext(ctx, "UpdateUser: updating user IsActive status", "userID", id, "newIsActive", user.IsActive)
		}
	}

	// If no changes were made, return the user without a database call.
	if !changesMade {
		slog.InfoContext(ctx, "UpdateUser: no actual changes detected for user", "userID", id)
		return user, nil
	}

	// Persist the updated user information.
	if err := s.userRepo.Update(ctx, user); err != nil {
		slog.ErrorContext(ctx, "UpdateUser: failed to update user in repository", "userID", id, "error", err)
		// Handle potential unique constraint violations that might occur at the DB level due to race conditions.
		return nil, fmt.Errorf("failed to save user updates: %w", err)
	}

	slog.InfoContext(ctx, "UpdateUser: user updated successfully", "userID", user.ID, "email", user.Email)
	return user, nil
}

// DeleteUser performs a soft delete on a user by their ID.
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	slog.InfoContext(ctx, "DeleteUser: attempting to delete user", "userID", id)

	if err := s.userRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "DeleteUser: user to delete not found in repository", "userID", id)
			return fmt.Errorf("user with ID '%s' not found: %w", id, err)
		}
		slog.ErrorContext(ctx, "DeleteUser: failed to delete user in repository", "userID", id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	slog.InfoContext(ctx, "DeleteUser: user deleted successfully", "userID", id)
	return nil
}

// ListUsers retrieves a paginated list of users.
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]models.User, int64, error) {
	slog.InfoContext(ctx, "ListUsers: attempting to list users", "page", page, "pageSize", pageSize)

	// Validate and set default pagination parameters.
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// Calculate the offset for the repository query.
	offset := (page - 1) * pageSize

	users, totalCount, err := s.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "ListUsers: failed to list users from repository", "page", page, "pageSize", pageSize, "error", err)
		return nil, 0, fmt.Errorf("could not retrieve users list: %w", err)
	}

	slog.InfoContext(ctx, "ListUsers: users listed successfully", "count", len(users), "totalCount", totalCount)
	return users, totalCount, nil
}
