package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
)

type mockImportBatchRepository struct {
	createdBatch            *model.ImportBatch
	createdItems            []model.ImportBatchItem
	createdInventoryBatches []model.InventoryBatch
	variantIDsExist         bool
	err                     error
}

func (m *mockImportBatchRepository) ValidateVariantIDsExist(ctx context.Context, variantIDs []int64) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.variantIDsExist, nil
}

func (m *mockImportBatchRepository) CreateImportBatchWithItems(
	ctx context.Context,
	batch *model.ImportBatch,
	items []model.ImportBatchItem,
	inventoryBatches []model.InventoryBatch,
) error {
	if m.err != nil {
		return m.err
	}
	m.createdBatch = batch
	m.createdItems = items
	m.createdInventoryBatches = inventoryBatches
	return nil
}

func (m *mockImportBatchRepository) GetBatchByID(ctx context.Context, id int64) (*model.ImportBatch, error) {
	return m.createdBatch, m.err
}

func TestImportBatchService_RegisterImportBatch_Success(t *testing.T) {
	mockRepo := &mockImportBatchRepository{
		variantIDsExist: true,
	}
	svc := NewImportBatchService(mockRepo)

	req := &payloads.CreateImportBatchRequest{
		BatchNumber:  "IMP-2026-001",
		ImportDate:   time.Now(),
		SupplierName: "Global Tech Supplier",
		CurrencyCode: "USD",
		ExchangeRate: decimal.NewFromFloat(15000.00),
		Items: []payloads.CreateImportBatchItemRequest{
			{
				ProductVariantID:         1,
				QuantityReceived:         10,
				UnitCostOriginalCurrency: decimal.NewFromFloat(10.00), // 10 * 15,000 = 150,000 IDR per unit. Total for 10 units = 1,500,000 IDR
			},
			{
				ProductVariantID:         2,
				QuantityReceived:         5,
				UnitCostOriginalCurrency: decimal.NewFromFloat(20.00), // 20 * 15,000 = 300,000 IDR per unit. Total for 5 units = 1,500,000 IDR
			},
		},
	}

	res, err := svc.RegisterImportBatch(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify Total Import Cost = 1,500,000 + 1,500,000 = 3,000,000 IDR
	expectedTotal := decimal.NewFromFloat(3000000.00)
	if !res.TotalImportCost.Equal(expectedTotal) {
		t.Errorf("expected total import cost %s, got %s", expectedTotal.String(), res.TotalImportCost.String())
	}

	// Verify Inventory Batches auto-population
	if len(mockRepo.createdInventoryBatches) != 2 {
		t.Fatalf("expected 2 inventory batches created, got %d", len(mockRepo.createdInventoryBatches))
	}

	item1Inv := mockRepo.createdInventoryBatches[0]
	if item1Inv.InitialQuantity != 10 || item1Inv.RemainingQuantity != 10 {
		t.Errorf("inventory batch 1 quantities mismatch: initial %d, remaining %d", item1Inv.InitialQuantity, item1Inv.RemainingQuantity)
	}

	expectedUnitCost1 := decimal.NewFromFloat(150000.00)
	if !item1Inv.UnitCostIDR.Equal(expectedUnitCost1) {
		t.Errorf("expected unit cost IDR %s, got %s", expectedUnitCost1.String(), item1Inv.UnitCostIDR.String())
	}
}

func TestImportBatchService_RegisterImportBatch_InvalidExchangeRate(t *testing.T) {
	mockRepo := &mockImportBatchRepository{variantIDsExist: true}
	svc := NewImportBatchService(mockRepo)

	req := &payloads.CreateImportBatchRequest{
		BatchNumber:  "IMP-2026-002",
		SupplierName: "Supplier",
		CurrencyCode: "USD",
		ExchangeRate: decimal.NewFromFloat(-10.00),
		Items: []payloads.CreateImportBatchItemRequest{
			{ProductVariantID: 1, QuantityReceived: 10, UnitCostOriginalCurrency: decimal.NewFromFloat(5.00)},
		},
	}

	_, err := svc.RegisterImportBatch(context.Background(), req)
	if err != ErrInvalidExchangeRate {
		t.Errorf("expected ErrInvalidExchangeRate, got %v", err)
	}
}
