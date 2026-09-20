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

func setupTestRouterWithProducts() http.Handler {
	userRepo := repository.NewUserRepository(nil)
	authService := service.NewAuthService(userRepo)
	authHandler := NewAuthHandler(authService)

	importRepo := repository.NewImportBatchRepository(nil)
	importService := service.NewImportBatchService(importRepo)
	importHandler := NewImportBatchHandler(importService)

	categoryRepo := repository.NewCategoryRepository(nil)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := NewCategoryHandler(categoryService)

	productRepo := repository.NewProductRepository(nil)
	productService := service.NewProductService(productRepo, categoryRepo)
	productHandler := NewProductHandler(productService)

	cfg := &config.Config{
		App: config.AppConfig{Name: "Maiora Test", Env: "test"},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST"},
		},
	}

	return NewRouter(&RouterConfig{
		AuthHandler:        authHandler,
		ImportBatchHandler: importHandler,
		CategoryHandler:    categoryHandler,
		ProductHandler:     productHandler,
		AppConfig:          cfg,
	})
}

func TestCategoryHandler_CreateCategory_Validation(t *testing.T) {
	router := setupTestRouterWithProducts()

	t.Run("POST /api/v1/categories empty name returns 400", func(t *testing.T) {
		body, _ := json.Marshal(payloads.CreateCategoryRequest{Name: ""})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestProductHandler_CreateProduct_Validation(t *testing.T) {
	router := setupTestRouterWithProducts()

	t.Run("POST /api/v1/products empty name returns 400", func(t *testing.T) {
		body, _ := json.Marshal(payloads.CreateProductRequest{Name: "", CategoryID: 1})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}
