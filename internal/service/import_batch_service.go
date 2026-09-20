package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
)

var (
	ErrInvalidExchangeRate = errors.New("exchange rate must be greater than zero")
	ErrVariantNotFound     = errors.New("one or more product variant IDs do not exist")
	ErrBatchNumberRequired = errors.New("batch number is required")
	ErrSupplierRequired    = errors.New("supplier name is required")
	ErrCurrencyRequired    = errors.New("currency code is required")
	ErrItemsRequired       = errors.New("import batch must contain at least one valid item")
	ErrInvalidQuantity     = errors.New("quantity received must be greater than zero")
	ErrInvalidUnitCost     = errors.New("unit cost must be greater than zero")
)

// ImportBatchService defines business logic for import batches
type ImportBatchService interface {
	RegisterImportBatch(ctx context.Context, req *payloads.CreateImportBatchRequest) (*payloads.ImportBatchResponse, error)
	GetBatchByID(ctx context.Context, id int64) (*payloads.ImportBatchResponse, error)
}

type importBatchService struct {
	repo repository.ImportBatchRepository
}

// NewImportBatchService creates a new instance of ImportBatchService
func NewImportBatchService(repo repository.ImportBatchRepository) ImportBatchService {
	return &importBatchService{repo: repo}
}

// RegisterImportBatch validates input, computes unit_cost_idr, auto-populates inventory batches, and executes transaction
func (s *importBatchService) RegisterImportBatch(ctx context.Context, req *payloads.CreateImportBatchRequest) (*payloads.ImportBatchResponse, error) {
	// 1. Input Validation
	batchNumber := strings.TrimSpace(req.BatchNumber)
	if batchNumber == "" {
		return nil, ErrBatchNumberRequired
	}
	supplierName := strings.TrimSpace(req.SupplierName)
	if supplierName == "" {
		return nil, ErrSupplierRequired
	}
	currencyCode := strings.TrimSpace(req.CurrencyCode)
	if currencyCode == "" {
		return nil, ErrCurrencyRequired
	}
	if req.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidExchangeRate
	}
	if len(req.Items) == 0 {
		return nil, ErrItemsRequired
	}

	// 2. Collect & Validate Product Variant IDs
	variantIDs := make([]int64, len(req.Items))
	for i, item := range req.Items {
		if item.QuantityReceived <= 0 {
			return nil, ErrInvalidQuantity
		}
		if item.UnitCostOriginalCurrency.LessThanOrEqual(decimal.Zero) {
			return nil, ErrInvalidUnitCost
		}
		variantIDs[i] = item.ProductVariantID
	}

	validVariants, err := s.repo.ValidateVariantIDsExist(ctx, variantIDs)
	if err != nil {
		return nil, fmt.Errorf("variant validation failed: %w", err)
	}
	if !validVariants {
		return nil, ErrVariantNotFound
	}

	// 3. Compute Financials & Prepare Entities
	totalImportCost := decimal.Zero
	batchItems := make([]model.ImportBatchItem, len(req.Items))
	inventoryBatches := make([]model.InventoryBatch, len(req.Items))

	for i, item := range req.Items {
		// unit_cost_idr = unit_cost_original_currency * exchange_rate
		unitCostIDR := item.UnitCostOriginalCurrency.Mul(req.ExchangeRate).Round(2)

		// subtotal = unit_cost_idr * quantity_received
		qtyDecimal := decimal.NewFromInt(int64(item.QuantityReceived))
		itemTotalCostIDR := unitCostIDR.Mul(qtyDecimal)
		totalImportCost = totalImportCost.Add(itemTotalCostIDR)

		// Import Batch Item
		batchItems[i] = model.ImportBatchItem{
			ProductVariantID:         item.ProductVariantID,
			QuantityReceived:         item.QuantityReceived,
			UnitCostOriginalCurrency: item.UnitCostOriginalCurrency,
			UnitCostIDR:              unitCostIDR,
		}

		// Auto-populate Inventory Batch (initial stock & remaining stock)
		inventoryBatches[i] = model.InventoryBatch{
			ProductVariantID:  item.ProductVariantID,
			InitialQuantity:   item.QuantityReceived,
			RemainingQuantity: item.QuantityReceived,
			UnitCostIDR:       unitCostIDR,
		}
	}

	// 4. Header Entity
	batch := &model.ImportBatch{
		BatchNumber:     batchNumber,
		ImportDate:      req.ImportDate,
		SupplierName:    supplierName,
		CurrencyCode:    currencyCode,
		ExchangeRate:    req.ExchangeRate,
		TotalImportCost: totalImportCost,
	}

	// 5. Execute Atomic Database Transaction
	if err := s.repo.CreateImportBatchWithItems(ctx, batch, batchItems, inventoryBatches); err != nil {
		return nil, fmt.Errorf("failed to save import batch transaction: %w", err)
	}

	// 6. Map to Response Payload
	responseItems := make([]payloads.ImportBatchItemResponse, len(batchItems))
	for i, bi := range batchItems {
		responseItems[i] = payloads.ImportBatchItemResponse{
			ID:                       bi.ID,
			ProductVariantID:         bi.ProductVariantID,
			QuantityReceived:         bi.QuantityReceived,
			UnitCostOriginalCurrency: bi.UnitCostOriginalCurrency,
			UnitCostIDR:              bi.UnitCostIDR,
		}
	}

	return &payloads.ImportBatchResponse{
		ID:              batch.ID,
		BatchNumber:     batch.BatchNumber,
		ImportDate:      batch.ImportDate.Format("2006-01-02"),
		SupplierName:    batch.SupplierName,
		CurrencyCode:    batch.CurrencyCode,
		ExchangeRate:    batch.ExchangeRate,
		TotalImportCost: batch.TotalImportCost,
		Items:           responseItems,
		CreatedAt:       batch.CreatedAt,
	}, nil
}

 func (s *importBatchService) GetBatchByID(ctx context.Context, id int64) (*payloads.ImportBatchResponse, error) {
	batch, err := s.repo.GetBatchByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, errors.New("import batch not found")
	}

	responseItems := make([]payloads.ImportBatchItemResponse, len(batch.Items))
	for i, bi := range batch.Items {
		responseItems[i] = payloads.ImportBatchItemResponse{
			ID:                       bi.ID,
			ProductVariantID:         bi.ProductVariantID,
			QuantityReceived:         bi.QuantityReceived,
			UnitCostOriginalCurrency: bi.UnitCostOriginalCurrency,
			UnitCostIDR:              bi.UnitCostIDR,
		}
	}

	return &payloads.ImportBatchResponse{
		ID:              batch.ID,
		BatchNumber:     batch.BatchNumber,
		ImportDate:      batch.ImportDate.Format("2006-01-02"),
		SupplierName:    batch.SupplierName,
		CurrencyCode:    batch.CurrencyCode,
		ExchangeRate:    batch.ExchangeRate,
		TotalImportCost: batch.TotalImportCost,
		Items:           responseItems,
		CreatedAt:       batch.CreatedAt,
	}, nil
}
