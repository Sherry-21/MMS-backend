package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// ImportBatch represents a shipment/batch of imported goods
type ImportBatch struct {
	ID              int64             `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchNumber     string            `gorm:"type:varchar(100);uniqueIndex;not null" json:"batch_number"`
	ImportDate      time.Time         `gorm:"type:date;not null" json:"import_date"`
	SupplierName    string            `gorm:"type:varchar(255);not null" json:"supplier_name"`
	CurrencyCode    string            `gorm:"type:varchar(10);not null;default:'USD'" json:"currency_code"`
	ExchangeRate    decimal.Decimal   `gorm:"type:decimal(18,4);not null" json:"exchange_rate"`
	TotalImportCost decimal.Decimal   `gorm:"type:decimal(18,2);not null;default:0.00" json:"total_import_cost"`
	CreatedAt       time.Time         `gorm:"autoCreateTime" json:"created_at"`
	Items           []ImportBatchItem `gorm:"foreignKey:ImportBatchID" json:"items,omitempty"`
}

func (ImportBatch) TableName() string {
	return "import_batches"
}
