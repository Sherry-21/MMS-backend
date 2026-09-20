package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/internal/service"
	"maiora-backend-service/pkg/utils"
)

// CategoryHandler handles HTTP requests for product categories
type CategoryHandler struct {
	service service.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler instance
func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: svc}
}

// CreateCategory godoc
// @Summary Create a product category
// @Description Creates a new product category with automated or custom slug
// @Tags Categories
// @Accept json
// @Produce json
// @Param request body payloads.CreateCategoryRequest true "Category Creation Payload"
// @Success 201 {object} payloads.BaseResponse{data=payloads.CategoryResponse} "Category created successfully"
// @Failure 400 {object} payloads.BaseResponse "Validation error"
// @Failure 409 {object} payloads.BaseResponse "Slug already exists"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/categories [post]
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req payloads.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
		return
	}

	response, err := h.service.CreateCategory(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCategoryNameRequired):
			utils.WriteBadRequest(w, err.Error(), nil)
		case errors.Is(err, repository.ErrCategorySlugExists):
			utils.WriteError(w, http.StatusConflict, err.Error(), nil)
		default:
			utils.WriteInternalServerError(w, "Failed to create category", err.Error())
		}
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, "Category created successfully", response)
}

// ListCategories godoc
// @Summary List all product categories
// @Description Retrieves all product categories ordered alphabetically
// @Tags Categories
// @Produce json
// @Success 200 {object} payloads.BaseResponse{data=[]payloads.CategoryResponse} "Categories retrieved successfully"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/categories [get]
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to list categories", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Categories retrieved successfully", categories)
}
