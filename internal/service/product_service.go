package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
)

var (
	ErrProductNameRequired  = errors.New("product name is required")
	ErrCategoryIDRequired   = errors.New("valid category_id is required")
	ErrSKURequired          = errors.New("sku is required")
	ErrVariantNameRequired  = errors.New("variant name is required")
	ErrSellingPriceNegative = errors.New("default selling price cannot be negative")
)

// ProductService defines business operations for products and variants
type ProductService interface {
	CreateProduct(ctx context.Context, req *payloads.CreateProductRequest) (*payloads.ProductResponse, error)
	CreateVariant(ctx context.Context, productID int64, req *payloads.CreateProductVariantRequest) (*payloads.ProductVariantResponse, error)
	GetProductByID(ctx context.Context, id int64) (*payloads.ProductResponse, error)
	ListProducts(ctx context.Context, search string, categoryID int64, page, limit int) ([]payloads.ProductResponse, int64, error)
	ListVariants(ctx context.Context, search string) ([]payloads.ProductVariantListItemResponse, error)
}

type productService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

// NewProductService creates a new ProductService instance
func NewProductService(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, req *payloads.CreateProductRequest) (*payloads.ProductResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrProductNameRequired
	}
	if req.CategoryID <= 0 {
		return nil, ErrCategoryIDRequired
	}

	// Validate category existence
	category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return nil, repository.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to verify category: %w", err)
	}

	// Validate variants & check SKU uniqueness
	skuSet := make(map[string]bool)
	variants := make([]model.ProductVariant, len(req.Variants))
	for i, v := range req.Variants {
		sku := strings.TrimSpace(v.SKU)
		if sku == "" {
			return nil, ErrSKURequired
		}
		if skuSet[sku] {
			return nil, fmt.Errorf("duplicate SKU %s provided in request", sku)
		}
		skuSet[sku] = true

		if v.DefaultSellingPrice.LessThan(decimal.Zero) {
			return nil, ErrSellingPriceNegative
		}

		// Check if SKU already exists in DB
		existing, err := s.productRepo.FindBySKU(ctx, sku)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("SKU '%s' is already in use by another product variant", sku)
		}

		variants[i] = model.ProductVariant{
			SKU:                 sku,
			VariantName:         strings.TrimSpace(v.VariantName),
			DefaultSellingPrice: v.DefaultSellingPrice,
		}
	}

	product := &model.Product{
		CategoryID:  req.CategoryID,
		Name:        name,
		Description: req.Description,
	}

	if err := s.productRepo.CreateProductWithVariants(ctx, product, variants); err != nil {
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	// Map response
	variantResponses := make([]payloads.ProductVariantResponse, len(product.Variants))
	for i, v := range product.Variants {
		variantResponses[i] = payloads.ProductVariantResponse{
			ID:                  v.ID,
			ProductID:           product.ID,
			SKU:                 v.SKU,
			VariantName:         v.VariantName,
			DefaultSellingPrice: v.DefaultSellingPrice,
			CreatedAt:           v.CreatedAt,
		}
	}

	return &payloads.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		Category: &payloads.CategoryResponse{
			ID:        category.ID,
			Name:      category.Name,
			Slug:      category.Slug,
			CreatedAt: category.CreatedAt,
		},
		Variants:  variantResponses,
		CreatedAt: product.CreatedAt,
	}, nil
}

func (s *productService) CreateVariant(ctx context.Context, productID int64, req *payloads.CreateProductVariantRequest) (*payloads.ProductVariantResponse, error) {
	if productID <= 0 {
		return nil, repository.ErrProductNotFound
	}

	// Verify product exists
	_, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	sku := strings.TrimSpace(req.SKU)
	if sku == "" {
		return nil, ErrSKURequired
	}
	if strings.TrimSpace(req.VariantName) == "" {
		return nil, ErrVariantNameRequired
	}
	if req.DefaultSellingPrice.LessThan(decimal.Zero) {
		return nil, ErrSellingPriceNegative
	}

	// Check SKU uniqueness
	existing, err := s.productRepo.FindBySKU(ctx, sku)
	if err == nil && existing != nil {
		return nil, repository.ErrSKUAlreadyExists
	}

	variant := &model.ProductVariant{
		ProductID:           productID,
		SKU:                 sku,
		VariantName:         req.VariantName,
		DefaultSellingPrice: req.DefaultSellingPrice,
	}

	if err := s.productRepo.CreateVariant(ctx, variant); err != nil {
		return nil, fmt.Errorf("failed to persist variant: %w", err)
	}

	return &payloads.ProductVariantResponse{
		ID:                  variant.ID,
		ProductID:           variant.ProductID,
		SKU:                 variant.SKU,
		VariantName:         variant.VariantName,
		DefaultSellingPrice: variant.DefaultSellingPrice,
		CreatedAt:           variant.CreatedAt,
	}, nil
}

func (s *productService) GetProductByID(ctx context.Context, id int64) (*payloads.ProductResponse, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var catResp *payloads.CategoryResponse
	if product.Category != nil {
		catResp = &payloads.CategoryResponse{
			ID:        product.Category.ID,
			Name:      product.Category.Name,
			Slug:      product.Category.Slug,
			CreatedAt: product.Category.CreatedAt,
		}
	}

	variantResponses := make([]payloads.ProductVariantResponse, len(product.Variants))
	for i, v := range product.Variants {
		variantResponses[i] = payloads.ProductVariantResponse{
			ID:                  v.ID,
			ProductID:           v.ProductID,
			SKU:                 v.SKU,
			VariantName:         v.VariantName,
			DefaultSellingPrice: v.DefaultSellingPrice,
			CreatedAt:           v.CreatedAt,
		}
	}

	return &payloads.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		Category:    catResp,
		Variants:    variantResponses,
		CreatedAt:   product.CreatedAt,
	}, nil
}

func (s *productService) ListProducts(ctx context.Context, search string, categoryID int64, page, limit int) ([]payloads.ProductResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	products, total, err := s.productRepo.ListProducts(ctx, search, categoryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]payloads.ProductResponse, len(products))
	for i, p := range products {
		var catResp *payloads.CategoryResponse
		if p.Category != nil {
			catResp = &payloads.CategoryResponse{
				ID:        p.Category.ID,
				Name:      p.Category.Name,
				Slug:      p.Category.Slug,
				CreatedAt: p.Category.CreatedAt,
			}
		}

		variants := make([]payloads.ProductVariantResponse, len(p.Variants))
		for j, v := range p.Variants {
			variants[j] = payloads.ProductVariantResponse{
				ID:                  v.ID,
				ProductID:           v.ProductID,
				SKU:                 v.SKU,
				VariantName:         v.VariantName,
				DefaultSellingPrice: v.DefaultSellingPrice,
				CreatedAt:           v.CreatedAt,
			}
		}

		result[i] = payloads.ProductResponse{
			ID:          p.ID,
			CategoryID:  p.CategoryID,
			Name:        p.Name,
			Description: p.Description,
			Category:    catResp,
			Variants:    variants,
			CreatedAt:   p.CreatedAt,
		}
	}

	return result, total, nil
}

func (s *productService) ListVariants(ctx context.Context, search string) ([]payloads.ProductVariantListItemResponse, error) {
	variants, err := s.productRepo.ListVariants(ctx, search)
	if err != nil {
		return nil, err
	}

	result := make([]payloads.ProductVariantListItemResponse, len(variants))
	for i, v := range variants {
		productName := ""
		categoryName := ""
		if v.Product != nil {
			productName = v.Product.Name
			if v.Product.Category != nil {
				categoryName = v.Product.Category.Name
			}
		}

		result[i] = payloads.ProductVariantListItemResponse{
			ID:                  v.ID,
			ProductID:           v.ProductID,
			ProductName:         productName,
			CategoryName:        categoryName,
			SKU:                 v.SKU,
			VariantName:         v.VariantName,
			DefaultSellingPrice: v.DefaultSellingPrice,
			CreatedAt:           v.CreatedAt,
		}
	}

	return result, nil
}
