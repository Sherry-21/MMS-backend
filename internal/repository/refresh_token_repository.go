package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"maiora-backend-service/internal/model"
)

// ErrRefreshTokenNotFound is returned when a refresh token record is not found
var ErrRefreshTokenNotFound = errors.New("refresh token not found")

// RefreshTokenRepository defines database operations for the refresh_tokens table
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByToken(ctx context.Context, tokenStr string) (*model.RefreshToken, error)
	RevokeToken(ctx context.Context, tokenStr string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new instance of RefreshTokenRepository
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create inserts a new refresh token record into the refresh_tokens table
func (r *refreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	if r.db == nil {
		return nil
	}

	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

// FindByToken retrieves a refresh token record by its token string
func (r *refreshTokenRepository) FindByToken(ctx context.Context, tokenStr string) (*model.RefreshToken, error) {
	if r.db == nil {
		return nil, ErrRefreshTokenNotFound
	}

	var record model.RefreshToken
	err := r.db.WithContext(ctx).Where("refresh_token = ?", tokenStr).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to query refresh token: %w", err)
	}

	return &record, nil
}

// RevokeToken marks an active refresh token as revoked (is_revoked = true)
func (r *refreshTokenRepository) RevokeToken(ctx context.Context, tokenStr string) error {
	if r.db == nil {
		return nil
	}

	result := r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("refresh_token = ?", tokenStr).
		Update("is_revoked", true)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", result.Error)
	}

	return nil
}

// RevokeAllUserTokens revokes all active refresh tokens belonging to a specific user
func (r *refreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if r.db == nil {
		return nil
	}

	result := r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Update("is_revoked", true)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke user refresh tokens: %w", result.Error)
	}

	return nil
}
