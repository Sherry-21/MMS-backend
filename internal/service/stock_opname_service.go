package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
)

var (
	ErrEmptyCountSubmission = errors.New("at least one item count must be submitted")
)

// StockOpnameService defines business operations for daily cycle counts and history tracking
type StockOpnameService interface {
	GenerateDailyAssignment(ctx context.Context, req *payloads.GenerateStockOpnameRequest, assignedBy *string) (*payloads.StockOpnameDetailResponse, error)
	GetTodayAssignment(ctx context.Context) (*payloads.StockOpnameDetailResponse, error)
	GetOpnameByID(ctx context.Context, id int64) (*payloads.StockOpnameDetailResponse, error)
	ListHistory(ctx context.Context, status, fromDate, toDate string, page, limit int) ([]payloads.StockOpnameResponse, *payloads.PaginationMeta, error)
	SubmitItemCounts(ctx context.Context, id int64, req *payloads.SubmitStockOpnameCountRequest) (*payloads.StockOpnameDetailResponse, error)
	CompleteOpname(ctx context.Context, id int64, req *payloads.CompleteStockOpnameRequest) (*payloads.StockOpnameDetailResponse, error)
}

type stockOpnameService struct {
	opnameRepo repository.StockOpnameRepository
}

// NewStockOpnameService creates a new StockOpnameService instance
func NewStockOpnameService(opnameRepo repository.StockOpnameRepository) StockOpnameService {
	return &stockOpnameService{opnameRepo: opnameRepo}
}

// GenerateDailyAssignment creates or returns today's random cycle count assignment
func (s *stockOpnameService) GenerateDailyAssignment(ctx context.Context, req *payloads.GenerateStockOpnameRequest, assignedBy *string) (*payloads.StockOpnameDetailResponse, error) {
	now := time.Now().UTC()

	// 1. Check if an active (PENDING/IN_PROGRESS) assignment already exists for today
	existing, err := s.opnameRepo.GetTodayActiveOpname(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing today opname: %w", err)
	}
	if existing != nil {
		resp := s.toOpnameDetailResponse(existing)
		return &resp, nil
	}

	// 2. Determine sample size
	sampleSize := 5
	notes := ""
	if req != nil {
		if req.SampleSize > 0 {
			sampleSize = req.SampleSize
		}
		notes = req.Notes
	}

	// 3. Select random product variants with aggregated on-hand stock
	variants, err := s.opnameRepo.SelectRandomVariantsForOpname(ctx, sampleSize)
	if err != nil {
		return nil, fmt.Errorf("failed to sample random variants: %w", err)
	}
	if len(variants) == 0 {
		return nil, repository.ErrNoVariantsAvailable
	}

	// 4. Generate unique opname number: SO-YYYYMMDD-XXX
	countToday, err := s.opnameRepo.CountOpnamesForDate(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("failed to determine opname sequence: %w", err)
	}
	opnameNumber := fmt.Sprintf("SO-%s-%03d", now.Format("20060102"), countToday+1)

	// 5. Construct Opname Header and Line Items
	header := &model.StockOpname{
		OpnameNumber:     opnameNumber,
		OpnameDate:       now,
		Status:           model.OpnameStatusPending,
		TotalItems:       len(variants),
		TotalDiscrepancy: 0,
		Notes:            notes,
		AssignedBy:       assignedBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	items := make([]model.StockOpnameItem, len(variants))
	for i, v := range variants {
		items[i] = model.StockOpnameItem{
			ProductVariantID: v.ProductVariantID,
			SystemQuantity:   v.CurrentStock,
			PhysicalQuantity: nil,
			Discrepancy:      0,
			Status:           model.OpnameItemStatusPending,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
	}

	// 6. Persist to database
	if err := s.opnameRepo.CreateOpnameWithItems(ctx, header, items); err != nil {
		return nil, fmt.Errorf("failed to persist stock opname: %w", err)
	}

	// 7. Reload complete entity with preloaded relations
	reloaded, err := s.opnameRepo.GetOpnameByID(ctx, header.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created stock opname: %w", err)
	}

	resp := s.toOpnameDetailResponse(reloaded)
	return &resp, nil
}

// GetTodayAssignment retrieves today's active stock opname assignment
func (s *stockOpnameService) GetTodayAssignment(ctx context.Context) (*payloads.StockOpnameDetailResponse, error) {
	now := time.Now().UTC()
	opname, err := s.opnameRepo.GetTodayActiveOpname(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve today's opname: %w", err)
	}
	if opname == nil {
		return nil, repository.ErrStockOpnameNotFound
	}

	resp := s.toOpnameDetailResponse(opname)
	return &resp, nil
}

// GetOpnameByID retrieves full details of a stock opname session by ID
func (s *stockOpnameService) GetOpnameByID(ctx context.Context, id int64) (*payloads.StockOpnameDetailResponse, error) {
	opname, err := s.opnameRepo.GetOpnameByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := s.toOpnameDetailResponse(opname)
	return &resp, nil
}

// ListHistory retrieves paginated stock opnames with filters
func (s *stockOpnameService) ListHistory(ctx context.Context, status, fromDate, toDate string, page, limit int) ([]payloads.StockOpnameResponse, *payloads.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	opnames, total, err := s.opnameRepo.ListOpnames(ctx, status, fromDate, toDate, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve stock opnames history: %w", err)
	}

	results := make([]payloads.StockOpnameResponse, len(opnames))
	for i := range opnames {
		results[i] = s.toOpnameResponse(&opnames[i])
	}

	meta := payloads.NewPaginationMeta(page, limit, total)
	return results, &meta, nil
}

// SubmitItemCounts records counted physical quantities and calculates discrepancies
func (s *stockOpnameService) SubmitItemCounts(ctx context.Context, id int64, req *payloads.SubmitStockOpnameCountRequest) (*payloads.StockOpnameDetailResponse, error) {
	if req == nil || len(req.Items) == 0 {
		return nil, ErrEmptyCountSubmission
	}

	if err := s.opnameRepo.UpdateOpnameItemCounts(ctx, id, req.Items); err != nil {
		return nil, err
	}

	return s.GetOpnameByID(ctx, id)
}

// CompleteOpname finalizes the stock opname session
func (s *stockOpnameService) CompleteOpname(ctx context.Context, id int64, req *payloads.CompleteStockOpnameRequest) (*payloads.StockOpnameDetailResponse, error) {
	conductedBy := ""
	notes := ""
	if req != nil {
		conductedBy = req.ConductedBy
		notes = req.Notes
	}

	completed, err := s.opnameRepo.CompleteOpname(ctx, id, conductedBy, notes)
	if err != nil {
		return nil, err
	}

	resp := s.toOpnameDetailResponse(completed)
	return &resp, nil
}

// Helper: converts model.StockOpname to payloads.StockOpnameResponse
func (s *stockOpnameService) toOpnameResponse(m *model.StockOpname) payloads.StockOpnameResponse {
	return payloads.StockOpnameResponse{
		ID:               m.ID,
		OpnameNumber:     m.OpnameNumber,
		OpnameDate:       m.OpnameDate.Format("2006-01-02"),
		Status:           m.Status,
		TotalItems:       m.TotalItems,
		TotalDiscrepancy: m.TotalDiscrepancy,
		Notes:            m.Notes,
		AssignedBy:       m.AssignedBy,
		ConductedBy:      m.ConductedBy,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		CompletedAt:      m.CompletedAt,
	}
}

// Helper: converts model.StockOpname to payloads.StockOpnameDetailResponse
func (s *stockOpnameService) toOpnameDetailResponse(m *model.StockOpname) payloads.StockOpnameDetailResponse {
	base := s.toOpnameResponse(m)
	items := make([]payloads.StockOpnameItemResponse, len(m.Items))

	for i, item := range m.Items {
		sku := ""
		variantName := ""
		productName := ""
		imageURL := ""

		if item.ProductVariant != nil {
			sku = item.ProductVariant.SKU
			variantName = item.ProductVariant.VariantName
			imageURL = item.ProductVariant.ImageURL
			if item.ProductVariant.Product != nil {
				productName = item.ProductVariant.Product.Name
			}
		}

		items[i] = payloads.StockOpnameItemResponse{
			ID:               item.ID,
			StockOpnameID:    item.StockOpnameID,
			ProductVariantID: item.ProductVariantID,
			SKU:              sku,
			VariantName:      variantName,
			ProductName:      productName,
			ImageURL:         imageURL,
			SystemQuantity:   item.SystemQuantity,
			PhysicalQuantity: item.PhysicalQuantity,
			Discrepancy:      item.Discrepancy,
			Status:           item.Status,
			Notes:            item.Notes,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
		}
	}

	return payloads.StockOpnameDetailResponse{
		StockOpnameResponse: base,
		Items:               items,
	}
}
