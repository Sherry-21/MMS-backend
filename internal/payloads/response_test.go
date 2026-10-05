package payloads

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBaseResponse_JSONSerialization(t *testing.T) {
	t.Run("success response without pagination has default empty meta object", func(t *testing.T) {
		resp := NewSuccessResponse("Category created successfully", map[string]interface{}{
			"id":   1,
			"name": "Electronics",
		})

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("unexpected error marshaling: %v", err)
		}

		jsonStr := string(data)

		if !strings.Contains(jsonStr, `"meta":{}`) {
			t.Errorf("expected json to contain '\"meta\":{}', got: %s", jsonStr)
		}

		var parsed struct {
			Success bool                   `json:"success"`
			Message string                 `json:"message"`
			Meta    map[string]interface{} `json:"meta"`
			Data    map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if !parsed.Success {
			t.Errorf("expected success to be true")
		}
		if parsed.Message != "Category created successfully" {
			t.Errorf("expected message 'Category created successfully', got %q", parsed.Message)
		}
		if len(parsed.Meta) != 0 {
			t.Errorf("expected meta to be empty map, got %v", parsed.Meta)
		}
		if parsed.Data["name"] != "Electronics" {
			t.Errorf("expected data name to be Electronics, got %v", parsed.Data["name"])
		}
	})

	t.Run("success response with pagination has populated meta and data items", func(t *testing.T) {
		meta := NewPaginationMeta(1, 10, 25)
		items := []string{"item1", "item2"}

		resp := NewSuccessResponseWithMeta("Products retrieved successfully", meta, items)

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("unexpected error marshaling: %v", err)
		}

		var parsed struct {
			Success bool           `json:"success"`
			Message string         `json:"message"`
			Meta    PaginationMeta `json:"meta"`
			Data    []string       `json:"data"`
		}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if parsed.Meta.Page != 1 || parsed.Meta.Limit != 10 || parsed.Meta.Total != 25 || parsed.Meta.TotalPages != 3 {
			t.Errorf("expected pagination meta {1 10 25 3}, got %+v", parsed.Meta)
		}
		if len(parsed.Data) != 2 || parsed.Data[0] != "item1" {
			t.Errorf("expected data items [item1 item2], got %v", parsed.Data)
		}
	})

	t.Run("error response has default empty meta object and error details", func(t *testing.T) {
		resp := NewErrorResponse("Validation failed", "name is required")

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("unexpected error marshaling: %v", err)
		}

		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"meta":{}`) {
			t.Errorf("expected json to contain '\"meta\":{}', got: %s", jsonStr)
		}

		var parsed struct {
			Success bool                   `json:"success"`
			Message string                 `json:"message"`
			Meta    map[string]interface{} `json:"meta"`
			Error   string                 `json:"error"`
		}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if parsed.Success {
			t.Errorf("expected success to be false")
		}
		if parsed.Error != "name is required" {
			t.Errorf("expected error 'name is required', got %q", parsed.Error)
		}
		if len(parsed.Meta) != 0 {
			t.Errorf("expected empty meta, got %v", parsed.Meta)
		}
	})

	t.Run("nil Meta defaults to empty object in JSON", func(t *testing.T) {
		resp := BaseResponse{
			Success: true,
			Message: "OK",
			Meta:    nil,
			Data:    "test",
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("unexpected error marshaling: %v", err)
		}

		if !strings.Contains(string(data), `"meta":{}`) {
			t.Errorf("expected nil Meta to serialize as {}, got %s", string(data))
		}
	})
}

func TestNewPaginationMeta(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		limit         int
		total         int64
		expectedPages int
	}{
		{"zero total items", 1, 10, 0, 0},
		{"exact page boundary", 1, 10, 20, 2},
		{"non-boundary items", 1, 10, 25, 3},
		{"zero limit guard", 1, 0, 10, 0},
		{"single item", 1, 20, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewPaginationMeta(tt.page, tt.limit, tt.total)
			if meta.TotalPages != tt.expectedPages {
				t.Errorf("expected total_pages %d, got %d", tt.expectedPages, meta.TotalPages)
			}
			if meta.Page != tt.page {
				t.Errorf("expected page %d, got %d", tt.page, meta.Page)
			}
			if meta.Limit != tt.limit {
				t.Errorf("expected limit %d, got %d", tt.limit, meta.Limit)
			}
			if meta.Total != tt.total {
				t.Errorf("expected total %d, got %d", tt.total, meta.Total)
			}
		})
	}
}
