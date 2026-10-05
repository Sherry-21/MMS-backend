package payloads

import (
	"time"
)

// CreateProductVariantRequest represents the payload for creating a single product variant/SKU
type CreateProductVariantRequest struct {
	SKU         string `json:"sku" example:"IPHONE15-BLK-128"`
	VariantName string `json:"variant_name" example:"Black 128GB"`
	ImageURL    string `json:"image_url,omitempty" example:"https://images.unsplash.com/photo-1510557880182-3d4d3cba35a5"`
}

// CreateProductRequest represents the payload for creating a new product catalog entry with optional inline variants
type CreateProductRequest struct {
	CategoryID  int64                         `json:"category_id" example:"1"`
	Name        string                        `json:"name" example:"iPhone 15"`
	Description string                        `json:"description" example:"Latest Apple smartphone"`
	Variants    []CreateProductVariantRequest `json:"variants,omitempty"`
}

// ProductVariantResponse represents the response DTO for a product variant
type ProductVariantResponse struct {
	ID          int64     `json:"id"`
	ProductID   int64     `json:"product_id"`
	SKU         string    `json:"sku"`
	VariantName string    `json:"variant_name"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProductResponse represents the response DTO for a product, including its category and variants
type ProductResponse struct {
	ID          int64                    `json:"id"`
	CategoryID  int64                    `json:"category_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Category    *CategoryResponse        `json:"category,omitempty"`
	Variants    []ProductVariantResponse `json:"variants,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
}

// ProductVariantListItemResponse represents flattened variant info with product and category name (ideal for import batch dropdowns)
type ProductVariantListItemResponse struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	ProductName  string    `json:"product_name"`
	CategoryName string    `json:"category_name,omitempty"`
	SKU          string    `json:"sku"`
	VariantName  string    `json:"variant_name"`
	ImageURL     string    `json:"image_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
