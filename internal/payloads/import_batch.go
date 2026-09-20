package payloads

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreateImportBatchItemRequest represents payload for an item within an import batch
type CreateImportBatchItemRequest struct {
	ProductVariantID         int64           `json:"product_variant_id" example:"1"`
	QuantityReceived         int             `json:"quantity_received" example:"100"`
	UnitCostOriginalCurrency decimal.Decimal `json:"unit_cost_original_currency" example:"12.5000"`
}

// CreateImportBatchRequest represents payload for registering an entire import batch
type CreateImportBatchRequest struct {
	BatchNumber  string                         `json:"batch_number" example:"IMP-2026-09-001"`
	ImportDate   time.Time                      `json:"import_date" example:"2026-09-20T00:00:00Z"`
	SupplierName string                         `json:"supplier_name" example:"Shenzhen Global Tech Ltd"`
	CurrencyCode string                         `json:"currency_code" example:"USD"`
	ExchangeRate decimal.Decimal                `json:"exchange_rate" example:"15850.5000"`
	Items        []CreateImportBatchItemRequest `json:"items"`
}

// ImportBatchItemResponse represents an item returned in the response
type ImportBatchItemResponse struct {
	ID                       int64           `json:"id"`
	ProductVariantID         int64           `json:"product_variant_id"`
	QuantityReceived         int             `json:"quantity_received"`
	UnitCostOriginalCurrency decimal.Decimal `json:"unit_cost_original_currency"`
	UnitCostIDR              decimal.Decimal `json:"unit_cost_idr"`
}

// ImportBatchResponse represents the full response of a registered import batch
type ImportBatchResponse struct {
	ID              int64                     `json:"id"`
	BatchNumber     string                    `json:"batch_number"`
	ImportDate      string                    `json:"import_date"`
	SupplierName    string                    `json:"supplier_name"`
	CurrencyCode    string                    `json:"currency_code"`
	ExchangeRate    decimal.Decimal           `json:"exchange_rate"`
	TotalImportCost decimal.Decimal           `json:"total_import_cost"`
	Items           []ImportBatchItemResponse `json:"items"`
	CreatedAt       time.Time                 `json:"created_at"`
}
