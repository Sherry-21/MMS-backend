package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"maiora-backend-service/config"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/internal/service"
)

type mockCategoryService struct {
	categories []payloads.CategoryResponse
}

func (m *mockCategoryService) CreateCategory(ctx context.Context, req *payloads.CreateCategoryRequest) (*payloads.CategoryResponse, error) {
	return nil, nil
}

func (m *mockCategoryService) ListCategories(ctx context.Context) ([]payloads.CategoryResponse, error) {
	return m.categories, nil
}

type mockProductService struct {
	products []payloads.ProductResponse
	total    int64
}

func (m *mockProductService) CreateProduct(ctx context.Context, req *payloads.CreateProductRequest) (*payloads.ProductResponse, error) {
	return nil, nil
}

func (m *mockProductService) CreateVariant(ctx context.Context, productID int64, req *payloads.CreateProductVariantRequest) (*payloads.ProductVariantResponse, error) {
	return nil, nil
}

func (m *mockProductService) GetProductByID(ctx context.Context, id int64) (*payloads.ProductResponse, error) {
	if len(m.products) > 0 {
		return &m.products[0], nil
	}
	return nil, repository.ErrProductNotFound
}

func (m *mockProductService) ListProducts(ctx context.Context, search string, categoryID int64, page, limit int) ([]payloads.ProductResponse, int64, error) {
	return m.products, m.total, nil
}

func (m *mockProductService) ListVariants(ctx context.Context, search string) ([]payloads.ProductVariantListItemResponse, error) {
	return nil, nil
}

func setupTestRouterWithProducts() http.Handler {
	userRepo := repository.NewUserRepository(nil)
	tokenRepo := repository.NewRefreshTokenRepository(nil)
	authService := service.NewAuthService(userRepo, tokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
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

func setupCustomMockRouter(catSvc service.CategoryService, prodSvc service.ProductService) http.Handler {
	cfg := &config.Config{
		App: config.AppConfig{Name: "Maiora Test", Env: "test"},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST"},
		},
	}

	return NewRouter(&RouterConfig{
		CategoryHandler: NewCategoryHandler(catSvc),
		ProductHandler:  NewProductHandler(prodSvc),
		AppConfig:       cfg,
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

func TestBaseResponse_MetaField_Handlers(t *testing.T) {
	mockCat := &mockCategoryService{
		categories: []payloads.CategoryResponse{
			{ID: 1, Name: "Smartphones", Slug: "smartphones", CreatedAt: time.Now()},
		},
	}

	mockProd := &mockProductService{
		products: []payloads.ProductResponse{
			{ID: 10, CategoryID: 1, Name: "iPhone 15", Description: "Apple flagship"},
		},
		total: 1,
	}

	router := setupCustomMockRouter(mockCat, mockProd)

	t.Run("GET /api/v1/categories (get all) returns empty meta {}", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, `"meta":{}`) {
			t.Errorf("expected response to contain '\"meta\":{}', got: %s", body)
		}

		var resp struct {
			Success bool                        `json:"success"`
			Message string                      `json:"message"`
			Meta    map[string]interface{}      `json:"meta"`
			Data    []payloads.CategoryResponse `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success true")
		}
		if len(resp.Meta) != 0 {
			t.Errorf("expected empty meta object, got %v", resp.Meta)
		}
		if len(resp.Data) != 1 || resp.Data[0].Name != "Smartphones" {
			t.Errorf("unexpected data: %+v", resp.Data)
		}
	})

	t.Run("GET /api/v1/products/{id} (get by id) returns empty meta {}", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products/10", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, `"meta":{}`) {
			t.Errorf("expected response to contain '\"meta\":{}', got: %s", body)
		}

		var resp struct {
			Success bool                      `json:"success"`
			Message string                    `json:"message"`
			Meta    map[string]interface{}    `json:"meta"`
			Data    *payloads.ProductResponse `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success true")
		}
		if len(resp.Meta) != 0 {
			t.Errorf("expected empty meta object, got %v", resp.Meta)
		}
		if resp.Data == nil || resp.Data.Name != "iPhone 15" {
			t.Errorf("unexpected product data: %+v", resp.Data)
		}
	})

	t.Run("GET /api/v1/products (paginated) returns populated pagination meta and products in data", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products?page=1&limit=20", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp struct {
			Success bool                       `json:"success"`
			Message string                     `json:"message"`
			Meta    payloads.PaginationMeta    `json:"meta"`
			Data    []payloads.ProductResponse `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success true")
		}
		if resp.Meta.Page != 1 || resp.Meta.Limit != 20 || resp.Meta.Total != 1 || resp.Meta.TotalPages != 1 {
			t.Errorf("expected pagination meta {1 20 1 1}, got %+v", resp.Meta)
		}
		if len(resp.Data) != 1 || resp.Data[0].Name != "iPhone 15" {
			t.Errorf("unexpected data items: %+v", resp.Data)
		}
	})
}
