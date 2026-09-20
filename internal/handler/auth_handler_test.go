package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"maiora-backend-service/config"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/internal/service"
)

func setupTestRouter() http.Handler {
	userRepo := repository.NewUserRepository(nil)
	authService := service.NewAuthService(userRepo)
	authHandler := NewAuthHandler(authService)

	cfg := &config.Config{
		App: config.AppConfig{Name: "Maiora Test", Env: "test"},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST"},
		},
	}

	return NewRouter(&RouterConfig{
		AuthHandler: authHandler,
		AppConfig:   cfg,
	})
}

func TestAuthHandler_Register_Validation(t *testing.T) {
	router := setupTestRouter()

	t.Run("POST /api/v1/auth/register empty payload returns 400", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("POST /api/v1/auth/register short password returns 400", func(t *testing.T) {
		body, _ := json.Marshal(payloads.RegisterRequest{
			Username: "valid_name",
			Password: "123", // too short
		})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}
