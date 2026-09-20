package payloads

import "time"

// RegisterRequest represents the payload for user registration
type RegisterRequest struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"SuperSecurePass123!"`
}

// LoginRequest represents the payload for user login
type LoginRequest struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"SuperSecurePass123!"`
}

// UserResponse represents safe user data returned in API responses (excludes sensitive hashes)
type UserResponse struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginResponse represents the response returned upon successful authentication
type LoginResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}
