package model

import (
	"time"

	"maiora-backend-service/internal/payloads"
)

// User represents the MySQL database entity for registered users
type User struct {
	ID           string    `gorm:"column:user_id;type:varchar(36);primaryKey" json:"user_id"`
	Username     string    `gorm:"column:username;type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"column:password;type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the MySQL table name for User
func (User) TableName() string {
	return "users"
}

func (u *User) ToUserResponse() payloads.UserResponse {
	return payloads.UserResponse{
		UserID:    u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
