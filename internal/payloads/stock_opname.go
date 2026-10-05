package payloads

import "time"

// GenerateStockOpnameRequest represents payload to generate a daily stock opname cycle count
type GenerateStockOpnameRequest struct {
	SampleSize int    `json:"sample_size" example:"5"`
	Notes      string `json:"notes" example:"Morning daily random cycle count"`
}

// CountItemInput represents a single audited item count submission
type CountItemInput struct {
	ItemID           int64  `json:"item_id" example:"1"`
	PhysicalQuantity int    `json:"physical_quantity" example:"48"`
	Notes            string `json:"notes" example:"2 units missing / damaged"`
}

// SubmitStockOpnameCountRequest represents payload to record physical counts for opname items
type SubmitStockOpnameCountRequest struct {
	Items []CountItemInput `json:"items"`
}

// CompleteStockOpnameRequest represents payload to finalize a stock opname audit
type CompleteStockOpnameRequest struct {
	ConductedBy string `json:"conducted_by" example:"Staff John"`
	Notes       string `json:"notes" example:"Cycle count completed and verified"`
}

// StockOpnameItemResponse represents item details within an opname response
type StockOpnameItemResponse struct {
	ID               int64     `json:"id"`
	StockOpnameID    int64     `json:"stock_opname_id"`
	ProductVariantID int64     `json:"product_variant_id"`
	SKU              string    `json:"sku"`
	VariantName      string    `json:"variant_name"`
	ProductName      string    `json:"product_name"`
	ImageURL         string    `json:"image_url,omitempty"`
	SystemQuantity   int       `json:"system_quantity"`
	PhysicalQuantity *int      `json:"physical_quantity"`
	Discrepancy      int       `json:"discrepancy"`
	Status           string    `json:"status"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// StockOpnameResponse represents summary data for a stock opname session in listings
type StockOpnameResponse struct {
	ID               int64      `json:"id"`
	OpnameNumber     string     `json:"opname_number"`
	OpnameDate       string     `json:"opname_date"`
	Status           string     `json:"status"`
	TotalItems       int        `json:"total_items"`
	TotalDiscrepancy int        `json:"total_discrepancy"`
	Notes            string     `json:"notes,omitempty"`
	AssignedBy       *string    `json:"assigned_by,omitempty"`
	ConductedBy      *string    `json:"conducted_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

// StockOpnameDetailResponse represents full data of a stock opname session including its line items
type StockOpnameDetailResponse struct {
	StockOpnameResponse
	Items []StockOpnameItemResponse `json:"items"`
}
