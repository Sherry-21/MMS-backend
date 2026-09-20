package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/pkg/utils"
)

// Common authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameRequired   = errors.New("username is required")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
	ErrUsernameTaken      = errors.New("username is already registered")
)

// AuthService defines the business logic interface for authentication
type AuthService interface {
	Register(ctx context.Context, req *payloads.RegisterRequest) (*payloads.UserResponse, error)
	Login(ctx context.Context, req *payloads.LoginRequest) (*payloads.LoginResponse, error)
}

// authService is the concrete implementation of AuthService
type authService struct {
	userRepo repository.UserRepository
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// Register validates user input, hashes the password securely with bcrypt, and creates the user
func (s *authService) Register(ctx context.Context, req *payloads.RegisterRequest) (*payloads.UserResponse, error) {
	// 1. Input sanitization and validation
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, ErrUsernameRequired
	}
	if len(req.Password) < 6 {
		return nil, ErrPasswordTooShort
	}

	// 2. Check if username is already taken
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, ErrUsernameTaken
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// 3. Hash the password with bcrypt
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to secure password: %w", err)
	}

	// 4. Create User entity with UUID as primary key
	newUser := &model.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// 5. Persist to database
	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to persist user: %w", err)
	}

	// 6. Return safe UserResponse DTO (excludes password hash)
	response := newUser.ToUserResponse()
	return &response, nil
}

// Login verifies user credentials against stored bcrypt password hash
func (s *authService) Login(ctx context.Context, req *payloads.LoginRequest) (*payloads.LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	// 1. Find user by username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// 2. Verify password with bcrypt
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// 3. Return safe LoginResponse
	return &payloads.LoginResponse{
		User:    user.ToUserResponse(),
		Message: "Authentication successful",
	}, nil
}
