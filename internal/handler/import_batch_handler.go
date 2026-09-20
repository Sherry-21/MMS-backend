package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/service"
	"maiora-backend-service/pkg/utils"
)

// ImportBatchHandler handles HTTP requests for import batches
type ImportBatchHandler struct {
	service service.ImportBatchService
}

// NewImportBatchHandler creates a new ImportBatchHandler instance
func NewImportBatchHandler(svc service.ImportBatchService) *ImportBatchHandler {
	return &ImportBatchHandler{service: svc}
}

// CreateImportBatch godoc
// @Summary Register a new import batch
// @Description Creates an import batch, line items, and auto-populates inventory batches in a single transaction
// @Tags Import Batches
// @Accept json
// @Produce json
// @Param request body payloads.CreateImportBatchRequest true "Import Batch Registration Payload"
// @Success 201 {object} payloads.BaseResponse{data=payloads.ImportBatchResponse} "Import batch registered successfully"
// @Failure 400 {object} payloads.BaseResponse "Invalid request payload or validation error"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/import-batches [post]
func (h *ImportBatchHandler) CreateImportBatch(w http.ResponseWriter, r *http.Request) {
	var req payloads.CreateImportBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
		return
	}

	response, err := h.service.RegisterImportBatch(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidExchangeRate),
			errors.Is(err, service.ErrVariantNotFound),
			errors.Is(err, service.ErrBatchNumberRequired),
			errors.Is(err, service.ErrSupplierRequired),
			errors.Is(err, service.ErrCurrencyRequired),
			errors.Is(err, service.ErrItemsRequired),
			errors.Is(err, service.ErrInvalidQuantity),
			errors.Is(err, service.ErrInvalidUnitCost):
			utils.WriteBadRequest(w, err.Error(), nil)
		default:
			utils.WriteInternalServerError(w, "Failed to register import batch", err.Error())
		}
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, "Import batch registered successfully", response)
}
