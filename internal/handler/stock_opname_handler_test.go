package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
)

type mockStockOpnameService struct {
	sampleDetail payloads.StockOpnameDetailResponse
}

func (m *mockStockOpnameService) GenerateDailyAssignment(ctx context.Context, req *payloads.GenerateStockOpnameRequest, assignedBy *string) (*payloads.StockOpnameDetailResponse, error) {
	return &m.sampleDetail, nil
}

func (m *mockStockOpnameService) GetTodayAssignment(ctx context.Context) (*payloads.StockOpnameDetailResponse, error) {
	return &m.sampleDetail, nil
}

func (m *mockStockOpnameService) GetOpnameByID(ctx context.Context, id int64) (*payloads.StockOpnameDetailResponse, error) {
	if id == 999 {
		return nil, repository.ErrStockOpnameNotFound
	}
	return &m.sampleDetail, nil
}

func (m *mockStockOpnameService) ListHistory(ctx context.Context, status, fromDate, toDate string, page, limit int) ([]payloads.StockOpnameResponse, *payloads.PaginationMeta, error) {
	meta := payloads.NewPaginationMeta(1, 20, 1)
	return []payloads.StockOpnameResponse{m.sampleDetail.StockOpnameResponse}, &meta, nil
}

func (m *mockStockOpnameService) SubmitItemCounts(ctx context.Context, id int64, req *payloads.SubmitStockOpnameCountRequest) (*payloads.StockOpnameDetailResponse, error) {
	return &m.sampleDetail, nil
}

func (m *mockStockOpnameService) CompleteOpname(ctx context.Context, id int64, req *payloads.CompleteStockOpnameRequest) (*payloads.StockOpnameDetailResponse, error) {
	return &m.sampleDetail, nil
}

func setupStockOpnameTestRouter() (http.Handler, *mockStockOpnameService) {
	mockSvc := &mockStockOpnameService{
		sampleDetail: payloads.StockOpnameDetailResponse{
			StockOpnameResponse: payloads.StockOpnameResponse{
				ID:           1,
				OpnameNumber: "SO-20260926-001",
				OpnameDate:   "2026-09-26",
				Status:       model.OpnameStatusPending,
				TotalItems:   1,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			Items: []payloads.StockOpnameItemResponse{
				{
					ID:               1,
					StockOpnameID:    1,
					ProductVariantID: 1,
					SKU:              "SKU-1",
					VariantName:      "Variant 1",
					ProductName:      "Product 1",
					SystemQuantity:   10,
					Status:           model.OpnameItemStatusPending,
				},
			},
		},
	}
	h := NewStockOpnameHandler(mockSvc)

	r := chi.NewRouter()
	r.Route("/api/v1/stock-opnames", func(api chi.Router) {
		api.Post("/generate-daily", h.GenerateDailyOpname)
		api.Get("/today", h.GetTodayOpname)
		api.Get("/", h.ListOpnames)
		api.Get("/{id}", h.GetOpnameByID)
		api.Post("/{id}/count", h.SubmitItemCounts)
		api.Post("/{id}/complete", h.CompleteOpname)
	})

	return r, mockSvc
}

func TestStockOpnameHandler(t *testing.T) {
	router, _ := setupStockOpnameTestRouter()

	t.Run("POST /generate-daily returns 201", func(t *testing.T) {
		body, _ := json.Marshal(payloads.GenerateStockOpnameRequest{SampleSize: 5})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/stock-opnames/generate-daily", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}

		var res payloads.BaseResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if !res.Success {
			t.Errorf("expected success true")
		}
	})

	t.Run("GET /today returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stock-opnames/today", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GET / history returns 200 with pagination meta", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stock-opnames?page=1&limit=10", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var res payloads.BaseResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res.Meta == nil {
			t.Errorf("expected non-nil meta")
		}
	})

	t.Run("GET /{id} valid returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stock-opnames/1", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GET /{id} not found returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stock-opnames/999", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("POST /{id}/count returns 200", func(t *testing.T) {
		body, _ := json.Marshal(payloads.SubmitStockOpnameCountRequest{
			Items: []payloads.CountItemInput{
				{ItemID: 1, PhysicalQuantity: 10, Notes: "Matched"},
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/stock-opnames/1/count", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("POST /{id}/complete returns 200", func(t *testing.T) {
		body, _ := json.Marshal(payloads.CompleteStockOpnameRequest{
			ConductedBy: "Staff",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/stock-opnames/1/complete", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}
