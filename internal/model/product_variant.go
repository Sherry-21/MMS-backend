package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// ProductVariant represents a specific SKU variant of a product
type ProductVariant struct {
	ID                  int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID           int64           `gorm:"not null;index" json:"product_id"`
	SKU                 string          `gorm:"type:varchar(100);uniqueIndex;not null" json:"sku"`
	VariantName         string          `gorm:"type:varchar(255);not null" json:"variant_name"`
	DefaultSellingPrice decimal.Decimal `gorm:"type:decimal(18,2);not null;default:0.00" json:"default_selling_price"`
	CreatedAt           time.Time       `gorm:"autoCreateTime" json:"created_at"`
	Product             *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (ProductVariant) TableName() string {
	return "product_variants"
}
