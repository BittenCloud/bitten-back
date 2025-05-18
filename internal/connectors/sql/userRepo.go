package sql

import (
	"bitback/internal/interfaces"
	"bitback/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepository implements the interfaces.UserRepository for interacting with user data in a SQL database.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of userRepository.
func NewUserRepository(sqlDB interfaces.SQLDatabase) interfaces.UserRepository {
	return &userRepository{
		db: sqlDB.GetGormClient(),
	}
}

// Create persists a new user record to the database.
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user to create cannot be nil")
	}
	// GORM's Create method will also trigger BeforeCreate hooks on the user model.
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by their unique UUID.
// Returns gorm.ErrRecordNotFound if no user is found.
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err // err will be gorm.ErrRecordNotFound if the record is not found.
	}
	return &user, nil
}

// GetByIDs retrieves a list of users based on a slice of UUIDs.
// If the ids slice is empty, it returns an empty list of users without querying the database.
func (r *userRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	if len(ids) == 0 {
		return []models.User{}, nil
	}
	var users []models.User
	// GORM automatically handles constructing an IN query for a slice of IDs: WHERE id IN (...).
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}
	return users, nil
}

// GetByEmail retrieves a user by their email address.
// Returns gorm.ErrRecordNotFound if no user with the specified email is found.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err // err will be gorm.ErrRecordNotFound if the record is not found.
	}
	return &user, nil
}

// Update saves changes to an existing user record in the database.
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user to update cannot be nil")
	}
	if user.ID == uuid.Nil {
		return errors.New("user ID is required for update")
	}

	err := r.db.WithContext(ctx).Updates(user).Error
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete performs a soft delete on a user record by setting the DeletedAt timestamp.
// Returns gorm.ErrRecordNotFound if the user to delete is not found.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("user ID is required for delete")
	}

	// GORM's Delete method on a model with gorm.DeletedAt will perform a soft delete.
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		// This means no record was found with the given ID to delete.
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List retrieves a paginated list of users, ordered by creation date (newest first).
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Count the total number of users (without pagination constraints) for pagination metadata.
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	if total == 0 {
		return []models.User{}, 0, nil
	}

	// Retrieve the paginated slice of users.
	query := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC")

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	return users, total, nil
}
