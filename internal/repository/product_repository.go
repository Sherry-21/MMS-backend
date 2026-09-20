package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"maiora-backend-service/internal/model"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrVariantNotFound = errors.New("product variant not found")
	ErrSKUAlreadyExists = errors.New("product variant with this SKU already exists")
)

// ProductRepository defines database operations for products and variants
type ProductRepository interface {
	CreateProductWithVariants(ctx context.Context, product *model.Product, variants []model.ProductVariant) error
	CreateVariant(ctx context.Context, variant *model.ProductVariant) error
	FindByID(ctx context.Context, id int64) (*model.Product, error)
	FindBySKU(ctx context.Context, sku string) (*model.ProductVariant, error)
	ListProducts(ctx context.Context, search string, categoryID int64, limit, offset int) ([]model.Product, int64, error)
	ListVariants(ctx context.Context, search string) ([]model.ProductVariant, error)
}

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository instance
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// CreateProductWithVariants creates a product and its initial variants in a single transaction
func (r *productRepository) CreateProductWithVariants(ctx context.Context, product *model.Product, variants []model.ProductVariant) error {
	if r.db == nil {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(product).Error; err != nil {
			return fmt.Errorf("failed to insert product: %w", err)
		}

		for i := range variants {
			variants[i].ProductID = product.ID
			if err := tx.Create(&variants[i]).Error; err != nil {
				return fmt.Errorf("failed to insert product variant %s: %w", variants[i].SKU, err)
			}
		}

		product.Variants = variants
		return nil
	})
}

func (r *productRepository) CreateVariant(ctx context.Context, variant *model.ProductVariant) error {
	if r.db == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(variant).Error; err != nil {
		return fmt.Errorf("failed to create variant: %w", err)
	}
	return nil
}

func (r *productRepository) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	if r.db == nil {
		return nil, ErrProductNotFound
	}
	var product model.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Variants").
		First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindBySKU(ctx context.Context, sku string) (*model.ProductVariant, error) {
	if r.db == nil {
		return nil, ErrVariantNotFound
	}
	var variant model.ProductVariant
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&variant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVariantNotFound
		}
		return nil, err
	}
	return &variant, nil
}

func (r *productRepository) ListProducts(ctx context.Context, search string, categoryID int64, limit, offset int) ([]model.Product, int64, error) {
	if r.db == nil {
		return []model.Product{}, 0, nil
	}

	query := r.db.WithContext(ctx).Model(&model.Product{})

	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}

	var products []model.Product
	err := query.
		Preload("Category").
		Preload("Variants").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}

	return products, total, nil
}

func (r *productRepository) ListVariants(ctx context.Context, search string) ([]model.ProductVariant, error) {
	if r.db == nil {
		return []model.ProductVariant{}, nil
	}

	query := r.db.WithContext(ctx).
		Preload("Product.Category").
		Model(&model.ProductVariant{})

	if search != "" {
		query = query.Where("sku LIKE ? OR variant_name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var variants []model.ProductVariant
	err := query.Order("id DESC").Find(&variants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query variants: %w", err)
	}

	return variants, nil
}
