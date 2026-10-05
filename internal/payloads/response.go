package payloads

import "encoding/json"

// BaseResponse represents a standard envelope for API responses
type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Meta    interface{} `json:"meta"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// MarshalJSON customizes JSON serialization for BaseResponse to ensure Meta is never null
func (r BaseResponse) MarshalJSON() ([]byte, error) {
	type Alias BaseResponse
	copy := Alias(r)
	if copy.Meta == nil {
		copy.Meta = map[string]interface{}{}
	}
	return json.Marshal(copy)
}

// PaginationMeta contains pagination details for list queries
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginationMeta calculates and constructs PaginationMeta
func NewPaginationMeta(page, limit int, total int64) PaginationMeta {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// NewSuccessResponse creates a successful BaseResponse payload with default empty meta
func NewSuccessResponse(message string, data interface{}) BaseResponse {
	return BaseResponse{
		Success: true,
		Message: message,
		Meta:    map[string]interface{}{},
		Data:    data,
	}
}

// NewSuccessResponseWithMeta creates a successful BaseResponse payload with custom meta (e.g., pagination)
func NewSuccessResponseWithMeta(message string, meta interface{}, data interface{}) BaseResponse {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	return BaseResponse{
		Success: true,
		Message: message,
		Meta:    meta,
		Data:    data,
	}
}

// NewErrorResponse creates an error BaseResponse payload with default empty meta
func NewErrorResponse(message string, err interface{}) BaseResponse {
	return BaseResponse{
		Success: false,
		Message: message,
		Meta:    map[string]interface{}{},
		Error:   err,
	}
}
