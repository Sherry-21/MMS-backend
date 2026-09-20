package payloads

import "time"

// CreateCategoryRequest represents the payload for creating a product category
type CreateCategoryRequest struct {
	Name string `json:"name" example:"Electronics"`
	Slug string `json:"slug,omitempty" example:"electronics"`
}

// CategoryResponse represents the response DTO for a category
type CategoryResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}
