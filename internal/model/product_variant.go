package model

import (
	"time"
)

// ProductVariant represents a specific SKU variant of a product
type ProductVariant struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID   int64     `gorm:"not null;index" json:"product_id"`
	SKU         string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"sku"`
	VariantName string    `gorm:"type:varchar(255);not null" json:"variant_name"`
	ImageURL    string    `gorm:"type:varchar(500)" json:"image_url,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	Product     *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (ProductVariant) TableName() string {
	return "product_variants"
}
