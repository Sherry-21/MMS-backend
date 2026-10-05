package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/pkg/utils"
)

type mockUserRepoForMiddleware struct {
	users map[string]*model.User
}

func (m *mockUserRepoForMiddleware) CreateUser(ctx context.Context, user *model.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepoForMiddleware) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepoForMiddleware) FindByID(ctx context.Context, userID string) (*model.User, error) {
	u, ok := m.users[userID]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func TestAuthMiddleware(t *testing.T) {
	secret := "secret-for-middleware-tests"
	repo := &mockUserRepoForMiddleware{
		users: make(map[string]*model.User),
	}
	middleware := NewAuthMiddleware(secret, repo)

	// Dummy next handler that records execution
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok || user == nil {
			http.Error(w, "missing context user", http.StatusInternalServerError)
			return
		}
		claims, ok := GetClaimsFromContext(r.Context())
		if !ok || claims == nil {
			http.Error(w, "missing context claims", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authorized: " + user.Username))
	})

	handlerToTest := middleware.Authenticate(nextHandler)

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("Invalid Header Format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Basic 12345")
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		expiredToken, err := utils.GenerateToken("usr-exp", "exp_user", utils.TokenTypeAccess, secret, -1*time.Minute)
		if err != nil {
			t.Fatalf("failed to generate expired token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("Refresh Token Passed as Access Token", func(t *testing.T) {
		refreshToken, err := utils.GenerateToken("usr-refresh", "refresh_user", utils.TokenTypeRefresh, secret, 15*time.Minute)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+refreshToken)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("User Not Found in Database", func(t *testing.T) {
		validToken, err := utils.GenerateToken("usr-deleted", "deleted_user", utils.TokenTypeAccess, secret, 15*time.Minute)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("Successful Stateless Authentication", func(t *testing.T) {
		validToken, err := utils.GenerateToken("usr-success", "success_user", utils.TokenTypeAccess, secret, 15*time.Minute)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		repo.users["usr-success"] = &model.User{
			ID:       "usr-success",
			Username: "success_user",
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		if rec.Body.String() != "authorized: success_user" {
			t.Errorf("expected response 'authorized: success_user', got '%s'", rec.Body.String())
		}
	})
}
