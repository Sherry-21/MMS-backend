package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// InventoryBatch tracks physical stock and unit cost for FIFO inventory tracking
type InventoryBatch struct {
	ID                int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ImportBatchItemID int64           `gorm:"not null;index" json:"import_batch_item_id"`
	ProductVariantID  int64           `gorm:"not null;index" json:"product_variant_id"`
	InitialQuantity   int             `gorm:"not null" json:"initial_quantity"`
	RemainingQuantity int             `gorm:"not null" json:"remaining_quantity"`
	UnitCostIDR       decimal.Decimal `gorm:"type:decimal(18,2);not null" json:"unit_cost_idr"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (InventoryBatch) TableName() string {
	return "inventory_batches"
}
