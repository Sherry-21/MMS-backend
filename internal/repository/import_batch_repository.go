package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"maiora-backend-service/internal/model"
)

// ImportBatchRepository defines database operations for import batches
type ImportBatchRepository interface {
	CreateImportBatchWithItems(ctx context.Context, batch *model.ImportBatch, items []model.ImportBatchItem, inventoryBatches []model.InventoryBatch) error
	GetBatchByID(ctx context.Context, id int64) (*model.ImportBatch, error)
	ValidateVariantIDsExist(ctx context.Context, variantIDs []int64) (bool, error)
}

type importBatchRepository struct {
	db *gorm.DB
}

// NewImportBatchRepository creates a new instance of ImportBatchRepository
func NewImportBatchRepository(db *gorm.DB) ImportBatchRepository {
	return &importBatchRepository{db: db}
}

// ValidateVariantIDsExist checks if all given product variant IDs exist in the database
func (r *importBatchRepository) ValidateVariantIDsExist(ctx context.Context, variantIDs []int64) (bool, error) {
	if r.db == nil {
		return true, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ProductVariant{}).
		Where("id IN ?", variantIDs).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to validate product variant IDs: %w", err)
	}

	return count == int64(len(variantIDs)), nil
}

// CreateImportBatchWithItems executes the creation of import batch, items, and inventory batches inside a single DB transaction
func (r *importBatchRepository) CreateImportBatchWithItems(
	ctx context.Context,
	batch *model.ImportBatch,
	items []model.ImportBatchItem,
	inventoryBatches []model.InventoryBatch,
) error {
	if r.db == nil {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create ImportBatch Header
		if err := tx.Create(batch).Error; err != nil {
			return fmt.Errorf("failed to create import batch header: %w", err)
		}

		// 2. Iterate and create items and related inventory batches
		for i := range items {
			items[i].ImportBatchID = batch.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return fmt.Errorf("failed to create import batch item: %w", err)
			}

			// 3. Link InventoryBatch to the newly created ImportBatchItem
			inventoryBatches[i].ImportBatchItemID = items[i].ID
			inventoryBatches[i].ProductVariantID = items[i].ProductVariantID
			inventoryBatches[i].UnitCostIDR = items[i].UnitCostIDR

			if err := tx.Create(&inventoryBatches[i]).Error; err != nil {
				return fmt.Errorf("failed to populate inventory batch: %w", err)
			}
		}

		return nil
	})
}

// GetBatchByID retrieves an import batch by ID with its line items preloaded
func (r *importBatchRepository) GetBatchByID(ctx context.Context, id int64) (*model.ImportBatch, error) {
	if r.db == nil {
		return nil, nil
	}

	var batch model.ImportBatch
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&batch, id).Error
	if err != nil {
		return nil, err
	}

	return &batch, nil
}
