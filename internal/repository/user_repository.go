package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"maiora-backend-service/internal/model"
)

// ErrUserNotFound is returned when a user record does not exist
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists is returned when attempting to register an existing username
var ErrUserAlreadyExists = errors.New("username is already taken")

// UserRepository defines the interface for User data access operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByID(ctx context.Context, userID string) (*model.User, error)
}

// userRepository is the GORM implementation of UserRepository
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository backed by GORM
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// CreateUser inserts a new user record into MySQL
func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	if r.db == nil {
		return nil
	}

	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByUsername retrieves a user by their username
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	if r.db == nil {
		return nil, ErrUserNotFound
	}

	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by username: %w", err)
	}

	return &user, nil
}

// FindByID retrieves a user by their user_id primary key
func (r *userRepository) FindByID(ctx context.Context, userID string) (*model.User, error) {
	if r.db == nil {
		return nil, ErrUserNotFound
	}

	var user model.User
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}

	return &user, nil
}
