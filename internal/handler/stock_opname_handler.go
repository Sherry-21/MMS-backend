package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"maiora-backend-service/internal/middleware"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/internal/service"
	"maiora-backend-service/pkg/utils"
)

// StockOpnameHandler handles HTTP requests for cycle count audits
type StockOpnameHandler struct {
	opnameService service.StockOpnameService
}

// NewStockOpnameHandler creates a new StockOpnameHandler instance
func NewStockOpnameHandler(opnameService service.StockOpnameService) *StockOpnameHandler {
	return &StockOpnameHandler{opnameService: opnameService}
}

// GenerateDailyOpname godoc
// @Summary Generate daily random cycle count assignment
// @Description Randomly samples product variants from inventory and generates a stock opname audit session for today
// @Tags Stock Opname
// @Accept json
// @Produce json
// @Param request body payloads.GenerateStockOpnameRequest false "Opname generation options"
// @Success 201 {object} payloads.BaseResponse{data=payloads.StockOpnameDetailResponse} "Stock opname generated or active retrieved"
// @Failure 400 {object} payloads.BaseResponse "Invalid request payload"
// @Failure 404 {object} payloads.BaseResponse "No product variants available"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/stock-opnames/generate-daily [post]
func (h *StockOpnameHandler) GenerateDailyOpname(w http.ResponseWriter, r *http.Request) {
	var req payloads.GenerateStockOpnameRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
			return
		}
	}

	var assignedBy *string
	if user, ok := middleware.GetUserFromContext(r.Context()); ok && user != nil {
		assignedBy = &user.ID
	}

	result, err := h.opnameService.GenerateDailyAssignment(r.Context(), &req, assignedBy)
	if err != nil {
		if errors.Is(err, repository.ErrNoVariantsAvailable) {
			utils.WriteNotFound(w, err.Error(), nil)
			return
		}
		utils.WriteInternalServerError(w, "Failed to generate daily stock opname", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, "Daily stock opname assignment ready", result)
}

// GetTodayOpname godoc
// @Summary Get today's active stock opname assignment
// @Description Retrieves the active stock opname cycle count session scheduled for today
// @Tags Stock Opname
// @Produce json
// @Success 200 {object} payloads.BaseResponse{data=payloads.StockOpnameDetailResponse} "Today's active stock opname"
// @Failure 404 {object} payloads.BaseResponse "No active stock opname found for today"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/stock-opnames/today [get]
func (h *StockOpnameHandler) GetTodayOpname(w http.ResponseWriter, r *http.Request) {
	result, err := h.opnameService.GetTodayAssignment(r.Context())
	if err != nil {
		if errors.Is(err, repository.ErrStockOpnameNotFound) {
			utils.WriteNotFound(w, "No active stock opname assignment found for today", nil)
			return
		}
		utils.WriteInternalServerError(w, "Failed to retrieve today's stock opname", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Today's stock opname retrieved successfully", result)
}

// ListOpnames godoc
// @Summary List stock opname record history
// @Description Retrieves paginated history of all stock opname audits with optional status and date filters
// @Tags Stock Opname
// @Produce json
// @Param page query int false "Page number (default 1)" default(1)
// @Param limit query int false "Items per page (default 20)" default(20)
// @Param status query string false "Filter by status (PENDING, IN_PROGRESS, COMPLETED, CANCELLED)"
// @Param from_date query string false "Start date filter (YYYY-MM-DD)"
// @Param to_date query string false "End date filter (YYYY-MM-DD)"
// @Success 200 {object} payloads.BaseResponse{data=[]payloads.StockOpnameResponse} "Stock opnames history"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/stock-opnames [get]
func (h *StockOpnameHandler) ListOpnames(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")
	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	results, meta, err := h.opnameService.ListHistory(r.Context(), status, fromDate, toDate, page, limit)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to query stock opname history", err.Error())
		return
	}

	utils.WritePaginatedSuccess(w, http.StatusOK, "Stock opname history retrieved successfully", results, *meta)
}

// GetOpnameByID godoc
// @Summary Get stock opname session by ID
// @Description Retrieves detailed information and all audited line items of a stock opname session
// @Tags Stock Opname
// @Produce json
// @Param id path int true "Stock Opname ID"
// @Success 200 {object} payloads.BaseResponse{data=payloads.StockOpnameDetailResponse} "Stock opname details"
// @Failure 400 {object} payloads.BaseResponse "Invalid ID"
// @Failure 404 {object} payloads.BaseResponse "Stock opname not found"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/stock-opnames/{id} [get]
func (h *StockOpnameHandler) GetOpnameByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.WriteBadRequest(w, "Invalid stock opname ID", nil)
		return
	}

	result, err := h.opnameService.GetOpnameByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrStockOpnameNotFound) {
			utils.WriteNotFound(w, "Stock opname not found", nil)
			return
		}
		utils.WriteInternalServerError(w, "Failed to retrieve stock opname", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Stock opname retrieved successfully", result)
}

// SubmitItemCounts godoc
// @Summary Submit physical counts for stock opname items
// @Description Records physical inventory counts for items in an active stock opname session and calculates discrepancies
// @Tags Stock Opname
// @Accept json
// @Produce json
// @Param id path int true "Stock Opname ID"
// @Param request body payloads.SubmitStockOpnameCountRequest true "Audited item counts"
// @Success 200 {object} payloads.BaseResponse{data=payloads.StockOpnameDetailResponse} "Counts recorded successfully"
// @Failure 400 {object} payloads.BaseResponse "Invalid payload or parameters"
// @Failure 404 {object} payloads.BaseResponse "Stock opname or item not found"
// @Failure 409 {object} payloads.BaseResponse "Stock opname already completed"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/stock-opnames/{id}/count [post]
func (h *StockOpnameHandler) SubmitItemCounts(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.WriteBadRequest(w, "Invalid stock opname ID", nil)
		return
	}

	var req payloads.SubmitStockOpnameCountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
		return
	}

	result, err := h.opnameService.SubmitItemCounts(r.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyCountSubmission):
			utils.WriteBadRequest(w, err.Error(), nil)
		case errors.Is(err, repository.ErrStockOpnameNotFound), errors.Is(err, repository.ErrOpnameItemNotFound):
			utils.WriteNotFound(w, err.Error(), nil)
		case errors.Is(err, repository.ErrStockOpnameAlreadyDone):
			utils.WriteError(w, http.StatusConflict, err.Error(), nil)
		default:
			utils.WriteInternalServerError(w, "Failed to submit item counts", err.Error())
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Item counts recorded successfully", result)
}

// CompleteOpname godoc
// @Summary Finalize and complete stock opname session
// @Description Marks stock opname session as COMPLETED, calculates total discrepancy, and records completion timestamp
// @Tags Stock Opname
// @Accept json
// @Produce json
// @Param id path int true "Stock Opname ID"
// @Param request body payloads.CompleteStockOpnameRequest false "Completion details"
// @Success 200 {object} payloads.BaseResponse{data=payloads.StockOpnameDetailResponse} "Stock opname completed successfully"
// @Failure 400 {object} payloads.BaseResponse "Invalid ID"
// @Failure 404 {object} payloads.BaseResponse "Stock opname not found"
// @Failure 409 {object} payloads.BaseResponse "Stock opname already completed"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/stock-opnames/{id}/complete [post]
func (h *StockOpnameHandler) CompleteOpname(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.WriteBadRequest(w, "Invalid stock opname ID", nil)
		return
	}

	var req payloads.CompleteStockOpnameRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.ConductedBy == "" {
		if user, ok := middleware.GetUserFromContext(r.Context()); ok && user != nil {
			req.ConductedBy = user.Username
		}
	}

	result, err := h.opnameService.CompleteOpname(r.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrStockOpnameNotFound):
			utils.WriteNotFound(w, err.Error(), nil)
		case errors.Is(err, repository.ErrStockOpnameAlreadyDone):
			utils.WriteError(w, http.StatusConflict, err.Error(), nil)
		default:
			utils.WriteInternalServerError(w, "Failed to complete stock opname", err.Error())
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Stock opname finalized successfully", result)
}
