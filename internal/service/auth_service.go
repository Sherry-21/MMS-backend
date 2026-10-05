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
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrUsernameRequired    = errors.New("username is required")
	ErrPasswordTooShort    = errors.New("password must be at least 6 characters")
	ErrUsernameTaken       = errors.New("username is already registered")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrRefreshTokenRequired = errors.New("refresh token is required")
)

// AuthService defines the business logic interface for authentication
type AuthService interface {
	Register(ctx context.Context, req *payloads.RegisterRequest) (*payloads.UserResponse, error)
	Login(ctx context.Context, req *payloads.LoginRequest) (*payloads.LoginResponse, error)
	RefreshToken(ctx context.Context, req *payloads.RefreshTokenRequest) (*payloads.TokenResponse, error)
}

// authService is the concrete implementation of AuthService
type authService struct {
	userRepo      repository.UserRepository
	tokenRepo     repository.RefreshTokenRepository
	jwtSecret     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	jwtSecret string,
	accessExpiry, refreshExpiry time.Duration,
) AuthService {
	if jwtSecret == "" {
		jwtSecret = "maiora-default-secret-key-change-in-production"
	}
	if accessExpiry <= 0 {
		accessExpiry = 15 * time.Minute
	}
	if refreshExpiry <= 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}
	return &authService{
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		jwtSecret:     jwtSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
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

	// 3. Generate access and refresh token pair
	tokens, err := utils.GenerateTokenPair(user.ID, user.Username, s.jwtSecret, s.accessExpiry, s.refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// 4. Save refresh token in dedicated refresh_tokens table (we do NOT store access token)
	if s.tokenRepo != nil {
		refreshTokenRecord := &model.RefreshToken{
			UserID:       user.ID,
			RefreshToken: tokens.RefreshToken,
			ExpiresAt:    time.Now().UTC().Add(s.refreshExpiry),
			IsRevoked:    false,
		}
		if err := s.tokenRepo.Create(ctx, refreshTokenRecord); err != nil {
			return nil, fmt.Errorf("failed to persist refresh token: %w", err)
		}
	}

	// 5. Return safe LoginResponse with access and refresh tokens
	return &payloads.LoginResponse{
		User:         user.ToUserResponse(),
		Token:        tokens.AccessToken,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
		Message:      "Authentication successful",
	}, nil
}

// RefreshToken validates an existing refresh token against the refresh_tokens table, rotates tokens, and returns a new token pair
func (s *authService) RefreshToken(ctx context.Context, req *payloads.RefreshTokenRequest) (*payloads.TokenResponse, error) {
	if req == nil || strings.TrimSpace(req.RefreshToken) == "" {
		return nil, ErrRefreshTokenRequired
	}

	refreshTokenStr := strings.TrimSpace(req.RefreshToken)

	// 1. Verify token signature, structure, and expiration
	claims, err := utils.ValidateJWT(refreshTokenStr, s.jwtSecret)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// 2. Enforce that the token type is strictly 'refresh'
	if claims.TokenType != utils.TokenTypeRefresh {
		return nil, ErrInvalidRefreshToken
	}

	// 3. Fetch user by claims UserID
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("failed to retrieve user for token refresh: %w", err)
	}

	// 4. Check refresh token in refresh_tokens table (must exist, !is_revoked, expires_at > now)
	if s.tokenRepo != nil {
		tokenRecord, err := s.tokenRepo.FindByToken(ctx, refreshTokenStr)
		if err != nil || tokenRecord == nil || tokenRecord.IsRevoked || tokenRecord.ExpiresAt.Before(time.Now().UTC()) {
			return nil, ErrInvalidRefreshToken
		}
		if tokenRecord.UserID != user.ID {
			return nil, ErrInvalidRefreshToken
		}

		// 5. Revoke the old refresh token (Refresh Token Rotation)
		if err := s.tokenRepo.RevokeToken(ctx, refreshTokenStr); err != nil {
			return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
		}
	}

	// 6. Generate a brand new token pair (Refresh Token Rotation)
	tokens, err := utils.GenerateTokenPair(user.ID, user.Username, s.jwtSecret, s.accessExpiry, s.refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	// 7. Store the new refresh token in refresh_tokens table
	if s.tokenRepo != nil {
		newRecord := &model.RefreshToken{
			UserID:       user.ID,
			RefreshToken: tokens.RefreshToken,
			ExpiresAt:    time.Now().UTC().Add(s.refreshExpiry),
			IsRevoked:    false,
		}
		if err := s.tokenRepo.Create(ctx, newRecord); err != nil {
			return nil, fmt.Errorf("failed to persist new refresh token: %w", err)
		}
	}

	return &payloads.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}
