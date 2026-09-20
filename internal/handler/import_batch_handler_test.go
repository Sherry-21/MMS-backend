package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"maiora-backend-service/config"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/internal/service"
)

func setupTestRouterWithImportBatch() http.Handler {
	userRepo := repository.NewUserRepository(nil)
	authService := service.NewAuthService(userRepo)
	authHandler := NewAuthHandler(authService)

	importRepo := repository.NewImportBatchRepository(nil)
	importService := service.NewImportBatchService(importRepo)
	importHandler := NewImportBatchHandler(importService)

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
		AppConfig:          cfg,
	})
}

func TestImportBatchHandler_CreateImportBatch_Validation(t *testing.T) {
	router := setupTestRouterWithImportBatch()

	t.Run("POST /api/v1/import-batches empty items returns 400", func(t *testing.T) {
		payload := payloads.CreateImportBatchRequest{
			BatchNumber:  "IMP-EMPTY",
			ImportDate:   time.Now(),
			SupplierName: "Supplier",
			CurrencyCode: "USD",
			ExchangeRate: decimal.NewFromFloat(15000),
			Items:        []payloads.CreateImportBatchItemRequest{},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/import-batches", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("POST /api/v1/import-batches invalid exchange rate returns 400", func(t *testing.T) {
		payload := payloads.CreateImportBatchRequest{
			BatchNumber:  "IMP-INVALID-RATE",
			ImportDate:   time.Now(),
			SupplierName: "Supplier",
			CurrencyCode: "USD",
			ExchangeRate: decimal.NewFromFloat(-1),
			Items: []payloads.CreateImportBatchItemRequest{
				{ProductVariantID: 1, QuantityReceived: 10, UnitCostOriginalCurrency: decimal.NewFromFloat(5.00)},
			},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/import-batches", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}
