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

// RefreshTokenRequest represents the payload for refreshing an access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UserResponse represents safe user data returned in API responses (excludes sensitive hashes and tokens)
type UserResponse struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TokenResponse represents a newly issued access and refresh token pair
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// LoginResponse represents the response returned upon successful authentication
type LoginResponse struct {
	User         UserResponse `json:"user"`
	Token        string       `json:"token"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	Message      string       `json:"message"`
}
