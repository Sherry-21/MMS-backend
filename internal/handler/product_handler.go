package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/internal/service"
	"maiora-backend-service/pkg/utils"
)

// ProductHandler handles HTTP requests for products and variants
type ProductHandler struct {
	service service.ProductService
}

// NewProductHandler creates a new ProductHandler instance
func NewProductHandler(svc service.ProductService) *ProductHandler {
	return &ProductHandler{service: svc}
}

// CreateProduct godoc
// @Summary Create a product
// @Description Creates a new product catalog item with optional inline variants
// @Tags Products
// @Accept json
// @Produce json
// @Param request body payloads.CreateProductRequest true "Product Creation Payload"
// @Success 201 {object} payloads.BaseResponse{data=payloads.ProductResponse} "Product created successfully"
// @Failure 400 {object} payloads.BaseResponse "Validation error"
// @Failure 404 {object} payloads.BaseResponse "Category not found"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req payloads.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
		return
	}

	response, err := h.service.CreateProduct(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNameRequired),
			errors.Is(err, service.ErrCategoryIDRequired),
			errors.Is(err, service.ErrSKURequired),
			errors.Is(err, service.ErrVariantNameRequired):
			utils.WriteBadRequest(w, err.Error(), nil)
		case errors.Is(err, repository.ErrCategoryNotFound):
			utils.WriteNotFound(w, "Category not found", nil)
		default:
			utils.WriteBadRequest(w, err.Error(), nil)
		}
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, "Product created successfully", response)
}

// GetProductByID godoc
// @Summary Get product by ID
// @Description Retrieves full product details including category and variants
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} payloads.BaseResponse{data=payloads.ProductResponse} "Product retrieved successfully"
// @Failure 404 {object} payloads.BaseResponse "Product not found"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteBadRequest(w, "Invalid product ID", nil)
		return
	}

	response, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			utils.WriteNotFound(w, "Product not found", nil)
			return
		}
		utils.WriteInternalServerError(w, "Failed to retrieve product", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Product retrieved successfully", response)
}

// ListProducts godoc
// @Summary List products
// @Description Retrieves paginated list of products with optional search and category filter
// @Tags Products
// @Produce json
// @Param search query string false "Search name or description"
// @Param category_id query int false "Filter by category ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Page size limit (default: 20)"
// @Success 200 {object} payloads.BaseResponse{data=[]payloads.ProductResponse,meta=payloads.PaginationMeta} "Products retrieved successfully"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	categoryID, _ := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	products, total, err := h.service.ListProducts(r.Context(), search, categoryID, page, limit)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to list products", err.Error())
		return
	}

	meta := payloads.NewPaginationMeta(page, limit, total)
	utils.WritePaginatedSuccess(w, http.StatusOK, "Products retrieved successfully", meta, products)
}

// CreateVariant godoc
// @Summary Add a variant to a product
// @Description Creates a new SKU variant for an existing product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body payloads.CreateProductVariantRequest true "Product Variant Payload"
// @Success 201 {object} payloads.BaseResponse{data=payloads.ProductVariantResponse} "Variant created successfully"
// @Failure 400 {object} payloads.BaseResponse "Validation error"
// @Failure 404 {object} payloads.BaseResponse "Product not found"
// @Failure 409 {object} payloads.BaseResponse "SKU already exists"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/products/{id}/variants [post]
func (h *ProductHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteBadRequest(w, "Invalid product ID", nil)
		return
	}

	var req payloads.CreateProductVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
		return
	}

	response, err := h.service.CreateVariant(r.Context(), productID, &req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProductNotFound):
			utils.WriteNotFound(w, "Product not found", nil)
		case errors.Is(err, repository.ErrSKUAlreadyExists):
			utils.WriteError(w, http.StatusConflict, err.Error(), nil)
		case errors.Is(err, service.ErrSKURequired),
			errors.Is(err, service.ErrVariantNameRequired):
			utils.WriteBadRequest(w, err.Error(), nil)
		default:
			utils.WriteInternalServerError(w, "Failed to create variant", err.Error())
		}
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, "Product variant created successfully", response)
}

// ListVariants godoc
// @Summary List product variants (SKUs)
// @Description Retrieves list of all variants/SKUs (ideal for Import Batch search/dropdowns)
// @Tags Products
// @Produce json
// @Param search query string false "Search SKU or variant name"
// @Success 200 {object} payloads.BaseResponse{data=[]payloads.ProductVariantListItemResponse} "Variants retrieved successfully"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/product-variants [get]
func (h *ProductHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	variants, err := h.service.ListVariants(r.Context(), search)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to list variants", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Product variants retrieved successfully", variants)
}
