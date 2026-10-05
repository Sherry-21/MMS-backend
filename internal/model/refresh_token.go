package model

import "time"

// RefreshToken represents a stored refresh token with revocation tracking (exactly 5 columns)
type RefreshToken struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       string    `gorm:"column:user_id;type:varchar(36);not null;index" json:"user_id"`
	RefreshToken string    `gorm:"column:refresh_token;type:varchar(500);not null;index" json:"refresh_token"`
	ExpiresAt    time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	IsRevoked    bool      `gorm:"column:is_revoked;not null;default:false" json:"is_revoked"`
}

// TableName returns the MySQL table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
