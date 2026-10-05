package utils

import (
	"errors"
	"testing"
	"time"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	secret := "test-secret-key-12345"
	userID := "user-uuid-123"
	username := "john_doe"

	t.Run("generate and validate access token", func(t *testing.T) {
		tokenStr, err := GenerateToken(userID, username, TokenTypeAccess, secret, 15*time.Minute)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		claims, err := ValidateJWT(tokenStr, secret)
		if err != nil {
			t.Fatalf("expected valid token, got error: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("expected userID %s, got %s", userID, claims.UserID)
		}
		if claims.Username != username {
			t.Errorf("expected username %s, got %s", username, claims.Username)
		}
		if claims.TokenType != TokenTypeAccess {
			t.Errorf("expected tokenType %s, got %s", TokenTypeAccess, claims.TokenType)
		}
	})

	t.Run("generate and validate token pair", func(t *testing.T) {
		pair, err := GenerateTokenPair(userID, username, secret, 15*time.Minute, 7*24*time.Hour)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if pair.AccessToken == "" || pair.RefreshToken == "" {
			t.Fatalf("tokens should not be empty")
		}
		if pair.ExpiresIn != 900 {
			t.Errorf("expected 900 seconds, got %d", pair.ExpiresIn)
		}

		accessClaims, err := ValidateJWT(pair.AccessToken, secret)
		if err != nil || accessClaims.TokenType != TokenTypeAccess {
			t.Errorf("access token validation failed: %v", err)
		}

		refreshClaims, err := ValidateJWT(pair.RefreshToken, secret)
		if err != nil || refreshClaims.TokenType != TokenTypeRefresh {
			t.Errorf("refresh token validation failed: %v", err)
		}
	})

	t.Run("expired token returns ErrTokenExpired", func(t *testing.T) {
		tokenStr, err := GenerateToken(userID, username, TokenTypeAccess, secret, -1*time.Minute)
		if err != nil {
			t.Fatalf("expected no error generating token, got %v", err)
		}

		_, err = ValidateJWT(tokenStr, secret)
		if !errors.Is(err, ErrTokenExpired) {
			t.Errorf("expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("tampered or invalid secret returns ErrInvalidToken", func(t *testing.T) {
		tokenStr, err := GenerateToken(userID, username, TokenTypeAccess, secret, 15*time.Minute)
		if err != nil {
			t.Fatalf("expected no error generating token, got %v", err)
		}

		_, err = ValidateJWT(tokenStr, "wrong-secret-key")
		if !errors.Is(err, ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("malformed token string returns ErrInvalidToken", func(t *testing.T) {
		_, err := ValidateJWT("not.a.valid.jwt", secret)
		if !errors.Is(err, ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}
