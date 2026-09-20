package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// ImportBatchItem represents an individual product line item inside an import batch
type ImportBatchItem struct {
	ID                       int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ImportBatchID            int64           `gorm:"not null;index" json:"import_batch_id"`
	ProductVariantID         int64           `gorm:"not null;index" json:"product_variant_id"`
	QuantityReceived         int             `gorm:"not null" json:"quantity_received"`
	UnitCostOriginalCurrency decimal.Decimal `gorm:"type:decimal(18,4);not null" json:"unit_cost_original_currency"`
	UnitCostIDR              decimal.Decimal `gorm:"type:decimal(18,2);not null" json:"unit_cost_idr"`
	CreatedAt                time.Time       `gorm:"autoCreateTime" json:"created_at"`
	ProductVariant           *ProductVariant `gorm:"foreignKey:ProductVariantID" json:"product_variant,omitempty"`
	InventoryBatch           *InventoryBatch `gorm:"foreignKey:ImportBatchItemID" json:"inventory_batch,omitempty"`
}

func (ImportBatchItem) TableName() string {
	return "import_batch_items"
}
