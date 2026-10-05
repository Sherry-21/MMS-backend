package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
)

var (
	ErrStockOpnameNotFound     = errors.New("stock opname not found")
	ErrStockOpnameAlreadyDone  = errors.New("stock opname has already been completed or cancelled")
	ErrNoVariantsAvailable     = errors.New("no product variants found in system to audit")
	ErrOpnameItemNotFound      = errors.New("stock opname item not found in this session")
)

// StockOpnameRepository defines database access operations for stock opname
type StockOpnameRepository interface {
	CreateOpnameWithItems(ctx context.Context, opname *model.StockOpname, items []model.StockOpnameItem) error
	GetOpnameByID(ctx context.Context, id int64) (*model.StockOpname, error)
	GetTodayActiveOpname(ctx context.Context, today time.Time) (*model.StockOpname, error)
	ListOpnames(ctx context.Context, status string, fromDate, toDate string, limit, offset int) ([]model.StockOpname, int64, error)
	UpdateOpnameItemCounts(ctx context.Context, opnameID int64, updates []payloads.CountItemInput) error
	CompleteOpname(ctx context.Context, opnameID int64, conductedBy string, notes string) (*model.StockOpname, error)
	SelectRandomVariantsForOpname(ctx context.Context, sampleSize int) ([]model.ProductVariantWithStock, error)
	CountOpnamesForDate(ctx context.Context, date time.Time) (int64, error)
}

type stockOpnameRepository struct {
	db *gorm.DB
}

// NewStockOpnameRepository creates a new instance of StockOpnameRepository
func NewStockOpnameRepository(db *gorm.DB) StockOpnameRepository {
	return &stockOpnameRepository{db: db}
}

// CreateOpnameWithItems inserts header and all line items atomically in a single transaction
func (r *stockOpnameRepository) CreateOpnameWithItems(ctx context.Context, opname *model.StockOpname, items []model.StockOpnameItem) error {
	if r.db == nil {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		opname.TotalItems = len(items)
		if err := tx.Create(opname).Error; err != nil {
			return fmt.Errorf("failed to create stock opname header: %w", err)
		}

		for i := range items {
			items[i].StockOpnameID = opname.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return fmt.Errorf("failed to insert stock opname item: %w", err)
			}
		}

		opname.Items = items
		return nil
	})
}

// GetOpnameByID fetches a stock opname by ID preloading items and their variant/product relations
func (r *stockOpnameRepository) GetOpnameByID(ctx context.Context, id int64) (*model.StockOpname, error) {
	if r.db == nil {
		return nil, ErrStockOpnameNotFound
	}

	var opname model.StockOpname
	err := r.db.WithContext(ctx).
		Preload("Items.ProductVariant.Product").
		First(&opname, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStockOpnameNotFound
		}
		return nil, fmt.Errorf("failed to query stock opname: %w", err)
	}

	return &opname, nil
}

// GetTodayActiveOpname checks if an active (PENDING or IN_PROGRESS) opname already exists for the given date
func (r *stockOpnameRepository) GetTodayActiveOpname(ctx context.Context, today time.Time) (*model.StockOpname, error) {
	if r.db == nil {
		return nil, nil
	}

	dateStr := today.Format("2006-01-02")
	var opname model.StockOpname
	err := r.db.WithContext(ctx).
		Preload("Items.ProductVariant.Product").
		Where("opname_date = ? AND status IN (?, ?)", dateStr, model.OpnameStatusPending, model.OpnameStatusInProgress).
		Order("id DESC").
		First(&opname).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query today's active opname: %w", err)
	}

	return &opname, nil
}

// ListOpnames retrieves paginated stock opname records with optional status and date filters
func (r *stockOpnameRepository) ListOpnames(ctx context.Context, status string, fromDate, toDate string, limit, offset int) ([]model.StockOpname, int64, error) {
	if r.db == nil {
		return []model.StockOpname{}, 0, nil
	}

	query := r.db.WithContext(ctx).Model(&model.StockOpname{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if fromDate != "" {
		query = query.Where("opname_date >= ?", fromDate)
	}
	if toDate != "" {
		query = query.Where("opname_date <= ?", toDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count stock opnames: %w", err)
	}

	var opnames []model.StockOpname
	err := query.
		Order("opname_date DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&opnames).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list stock opnames: %w", err)
	}

	return opnames, total, nil
}

// UpdateOpnameItemCounts updates physical counts, notes, and discrepancies for items in a session
func (r *stockOpnameRepository) UpdateOpnameItemCounts(ctx context.Context, opnameID int64, updates []payloads.CountItemInput) error {
	if r.db == nil {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var opname model.StockOpname
		if err := tx.First(&opname, opnameID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrStockOpnameNotFound
			}
			return err
		}

		if opname.Status == model.OpnameStatusCompleted || opname.Status == model.OpnameStatusCancelled {
			return ErrStockOpnameAlreadyDone
		}

		for _, itemUpdate := range updates {
			var item model.StockOpnameItem
			if err := tx.Where("id = ? AND stock_opname_id = ?", itemUpdate.ItemID, opnameID).First(&item).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("%w: item id %d", ErrOpnameItemNotFound, itemUpdate.ItemID)
				}
				return err
			}

			qty := itemUpdate.PhysicalQuantity
			discrepancy := qty - item.SystemQuantity

			err := tx.Model(&item).Updates(map[string]interface{}{
				"physical_quantity": qty,
				"discrepancy":       discrepancy,
				"status":            model.OpnameItemStatusCounted,
				"notes":             itemUpdate.Notes,
				"updated_at":        time.Now().UTC(),
			}).Error
			if err != nil {
				return fmt.Errorf("failed to update item count: %w", err)
			}
		}

		// Update opname status to IN_PROGRESS if currently PENDING
		if opname.Status == model.OpnameStatusPending {
			if err := tx.Model(&opname).Update("status", model.OpnameStatusInProgress).Error; err != nil {
				return fmt.Errorf("failed to update opname status: %w", err)
			}
		}

		return nil
	})
}

// CompleteOpname finalizes the opname, calculates total discrepancy, and marks as COMPLETED
func (r *stockOpnameRepository) CompleteOpname(ctx context.Context, opnameID int64, conductedBy string, notes string) (*model.StockOpname, error) {
	if r.db == nil {
		return nil, ErrStockOpnameNotFound
	}

	var completedOpname model.StockOpname
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var opname model.StockOpname
		if err := tx.First(&opname, opnameID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrStockOpnameNotFound
			}
			return err
		}

		if opname.Status == model.OpnameStatusCompleted || opname.Status == model.OpnameStatusCancelled {
			return ErrStockOpnameAlreadyDone
		}

		// Calculate total discrepancy across all counted items
		var totalDiscrepancy int64
		err := tx.Model(&model.StockOpnameItem{}).
			Where("stock_opname_id = ?", opnameID).
			Select("COALESCE(SUM(discrepancy), 0)").
			Scan(&totalDiscrepancy).Error
		if err != nil {
			return fmt.Errorf("failed to aggregate discrepancies: %w", err)
		}

		now := time.Now().UTC()
		updates := map[string]interface{}{
			"status":            model.OpnameStatusCompleted,
			"total_discrepancy": int(totalDiscrepancy),
			"completed_at":      now,
			"updated_at":        now,
		}
		if conductedBy != "" {
			updates["conducted_by"] = conductedBy
		}
		if notes != "" {
			updates["notes"] = notes
		}

		if err := tx.Model(&opname).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to finalize stock opname: %w", err)
		}

		// Load preloaded completed session
		return tx.Preload("Items.ProductVariant.Product").First(&completedOpname, opnameID).Error
	})

	if err != nil {
		return nil, err
	}

	return &completedOpname, nil
}

// SelectRandomVariantsForOpname picks random product variants along with their current on-hand stock
func (r *stockOpnameRepository) SelectRandomVariantsForOpname(ctx context.Context, sampleSize int) ([]model.ProductVariantWithStock, error) {
	if r.db == nil {
		return nil, nil
	}

	if sampleSize <= 0 {
		sampleSize = 5
	}

	var results []model.ProductVariantWithStock
	query := `
		SELECT 
			pv.id AS product_variant_id,
			pv.product_id,
			pv.sku,
			pv.variant_name,
			pv.image_url,
			p.name AS product_name,
			COALESCE(SUM(ib.remaining_quantity), 0) AS current_stock
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		LEFT JOIN inventory_batches ib ON ib.product_variant_id = pv.id
		GROUP BY pv.id, pv.product_id, pv.sku, pv.variant_name, pv.image_url, p.name
		ORDER BY RANDOM()
		LIMIT ?
	`

	if err := r.db.WithContext(ctx).Raw(query, sampleSize).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to select random variants for opname: %w", err)
	}

	return results, nil
}

// CountOpnamesForDate counts how many opnames were created on a given date to generate sequence numbers
func (r *stockOpnameRepository) CountOpnamesForDate(ctx context.Context, date time.Time) (int64, error) {
	if r.db == nil {
		return 0, nil
	}

	dateStr := date.Format("2006-01-02")
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.StockOpname{}).
		Where("opname_date = ?", dateStr).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count opnames for date: %w", err)
	}

	return count, nil
}
