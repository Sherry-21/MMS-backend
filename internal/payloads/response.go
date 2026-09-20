package payloads

import "time"

// BaseResponse represents a standard envelope for API responses
type BaseResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewSuccessResponse creates a successful BaseResponse payload
func NewSuccessResponse(message string, data interface{}) BaseResponse {
	return BaseResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// NewErrorResponse creates an erro r BaseResponse payload
func NewErrorResponse(message string, err interface{}) BaseResponse {
	return BaseResponse{
		Success:   false,
		Message:   message,
		Error:     err,
		Timestamp: time.Now().UTC(),
	}
}
