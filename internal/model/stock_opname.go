package model

import (
	"time"
)

// Stock Opname Status Constants
const (
	OpnameStatusPending    = "PENDING"
	OpnameStatusInProgress = "IN_PROGRESS"
	OpnameStatusCompleted  = "COMPLETED"
	OpnameStatusCancelled  = "CANCELLED"

	OpnameItemStatusPending = "PENDING"
	OpnameItemStatusCounted = "COUNTED"
)

// StockOpname represents a cycle count / physical inventory audit session
type StockOpname struct {
	ID               int64             `gorm:"primaryKey;autoIncrement" json:"id"`
	OpnameNumber     string            `gorm:"type:varchar(100);uniqueIndex;not null" json:"opname_number"`
	OpnameDate       time.Time         `gorm:"type:date;not null;index" json:"opname_date"`
	Status           string            `gorm:"type:varchar(50);not null;default:'PENDING';index" json:"status"`
	TotalItems       int               `gorm:"not null;default:0" json:"total_items"`
	TotalDiscrepancy int               `gorm:"not null;default:0" json:"total_discrepancy"`
	Notes            string            `gorm:"type:text" json:"notes,omitempty"`
	AssignedBy       *string           `gorm:"type:varchar(36)" json:"assigned_by,omitempty"`
	ConductedBy      *string           `gorm:"type:varchar(255)" json:"conducted_by,omitempty"`
	CreatedAt        time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	CompletedAt      *time.Time        `json:"completed_at,omitempty"`
	Items            []StockOpnameItem `gorm:"foreignKey:StockOpnameID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (StockOpname) TableName() string {
	return "stock_opnames"
}

// StockOpnameItem represents an audited line item in a stock opname session
type StockOpnameItem struct {
	ID               int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	StockOpnameID    int64           `gorm:"not null;index" json:"stock_opname_id"`
	ProductVariantID int64           `gorm:"not null;index" json:"product_variant_id"`
	SystemQuantity   int             `gorm:"not null;default:0" json:"system_quantity"`
	PhysicalQuantity *int            `json:"physical_quantity"`
	Discrepancy      int             `gorm:"not null;default:0" json:"discrepancy"`
	Status           string          `gorm:"type:varchar(50);not null;default:'PENDING';index" json:"status"`
	Notes            string          `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt        time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	ProductVariant   *ProductVariant `gorm:"foreignKey:ProductVariantID" json:"product_variant,omitempty"`
}

func (StockOpnameItem) TableName() string {
	return "stock_opname_items"
}

// ProductVariantWithStock is a helper struct for cycle count selection
type ProductVariantWithStock struct {
	ProductVariantID int64  `json:"product_variant_id"`
	ProductID        int64  `json:"product_id"`
	SKU              string `json:"sku"`
	VariantName      string `json:"variant_name"`
	ProductName      string `json:"product_name"`
	ImageURL         string `json:"image_url,omitempty"`
	CurrentStock     int    `json:"current_stock"`
}
