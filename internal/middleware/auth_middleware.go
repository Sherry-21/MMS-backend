package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/pkg/utils"
)

type contextKey string

const (
	UserContextKey   contextKey = "currentUser"
	ClaimsContextKey contextKey = "jwtClaims"
)

// AuthMiddleware validates stateless JWT access tokens and ensures user account validity
type AuthMiddleware struct {
	jwtSecret string
	userRepo  repository.UserRepository
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(jwtSecret string, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		userRepo:  userRepo,
	}
}

// Authenticate verifies the incoming stateless JWT access token from the Authorization header
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteError(w, http.StatusUnauthorized, "Authorization header is required", nil)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.WriteError(w, http.StatusUnauthorized, "Invalid authorization header format. Expected 'Bearer <token>'", nil)
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			utils.WriteError(w, http.StatusUnauthorized, "Token is required", nil)
			return
		}

		// 1. Cryptographic validation of access token (stateless)
		claims, err := utils.ValidateJWT(tokenStr, m.jwtSecret)
		if err != nil {
			if errors.Is(err, utils.ErrTokenExpired) {
				utils.WriteError(w, http.StatusUnauthorized, "Access token has expired. Please refresh your token", nil)
				return
			}
			utils.WriteError(w, http.StatusUnauthorized, "Invalid or malformed access token", nil)
			return
		}

		// 2. Ensure token type is strictly 'access'
		if claims.TokenType != utils.TokenTypeAccess {
			utils.WriteError(w, http.StatusUnauthorized, "Invalid token type for authorization", nil)
			return
		}

		// 3. Verify user account exists and is valid
		user, err := m.userRepo.FindByID(r.Context(), claims.UserID)
		if err != nil || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User account not found or deactivated", nil)
			return
		}

		// Attach user and claims to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts the authenticated model.User from request context
func GetUserFromContext(ctx context.Context) (*model.User, bool) {
	u, ok := ctx.Value(UserContextKey).(*model.User)
	return u, ok
}

// GetClaimsFromContext extracts the validated utils.JWTClaims from request context
func GetClaimsFromContext(ctx context.Context) (*utils.JWTClaims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*utils.JWTClaims)
	return claims, ok
}
