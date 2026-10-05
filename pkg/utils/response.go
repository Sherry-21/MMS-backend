package utils

import (
	"encoding/json"
	"net/http"

	"maiora-backend-service/internal/payloads"
)

// WriteJSON sends a JSON response with the given status code and payload
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteSuccess sends a standardized 200 OK or custom 2xx JSON response with default empty meta
func WriteSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	response := payloads.NewSuccessResponse(message, data)
	WriteJSON(w, status, response)
}

// WritePaginatedSuccess sends a standardized 2xx JSON response with pagination metadata in meta and items in data
func WritePaginatedSuccess(w http.ResponseWriter, status int, message string, meta interface{}, data interface{}) {
	response := payloads.NewSuccessResponseWithMeta(message, meta, data)
	WriteJSON(w, status, response)
}

// WriteError sends a standardized error JSON response with default empty meta
func WriteError(w http.ResponseWriter, status int, message string, err interface{}) {
	response := payloads.NewErrorResponse(message, err)
	WriteJSON(w, status, response)
}

// WriteBadRequest sends a 400 Bad Request standardized JSON response
func WriteBadRequest(w http.ResponseWriter, message string, err interface{}) {
	WriteError(w, http.StatusBadRequest, message, err)
}

// WriteInternalServerError sends a 500 Internal Server Error standardized JSON response
func WriteInternalServerError(w http.ResponseWriter, message string, err interface{}) {
	WriteError(w, http.StatusInternalServerError, message, err)
}

// WriteNotFound sends a 404 Not Found standardized JSON response
func WriteNotFound(w http.ResponseWriter, message string, err interface{}) {
	WriteError(w, http.StatusNotFound, message, err)
}
