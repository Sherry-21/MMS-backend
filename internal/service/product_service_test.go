package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
)

type mockProductRepository struct {
	products map[int64]*model.Product
	variants map[string]*model.ProductVariant
}

func newMockProductRepository() *mockProductRepository {
	return &mockProductRepository{
		products: make(map[int64]*model.Product),
		variants: make(map[string]*model.ProductVariant),
	}
}

func (m *mockProductRepository) CreateProductWithVariants(ctx context.Context, product *model.Product, variants []model.ProductVariant) error {
	product.ID = int64(len(m.products) + 1)
	product.CreatedAt = time.Now()
	for i := range variants {
		variants[i].ID = int64(len(m.variants) + 1)
		variants[i].ProductID = product.ID
		variants[i].CreatedAt = time.Now()
		m.variants[variants[i].SKU] = &variants[i]
	}
	product.Variants = variants
	m.products[product.ID] = product
	return nil
}

func (m *mockProductRepository) CreateVariant(ctx context.Context, variant *model.ProductVariant) error {
	variant.ID = int64(len(m.variants) + 1)
	variant.CreatedAt = time.Now()
	m.variants[variant.SKU] = variant
	return nil
}

func (m *mockProductRepository) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	p, exists := m.products[id]
	if !exists {
		return nil, repository.ErrProductNotFound
	}
	return p, nil
}

func (m *mockProductRepository) FindBySKU(ctx context.Context, sku string) (*model.ProductVariant, error) {
	v, exists := m.variants[sku]
	if !exists {
		return nil, repository.ErrVariantNotFound
	}
	return v, nil
}

func (m *mockProductRepository) ListProducts(ctx context.Context, search string, categoryID int64, limit, offset int) ([]model.Product, int64, error) {
	list := make([]model.Product, 0, len(m.products))
	for _, p := range m.products {
		list = append(list, *p)
	}
	return list, int64(len(list)), nil
}

func (m *mockProductRepository) ListVariants(ctx context.Context, search string) ([]model.ProductVariant, error) {
	list := make([]model.ProductVariant, 0, len(m.variants))
	for _, v := range m.variants {
		list = append(list, *v)
	}
	return list, nil
}

type mockCategoryRepository struct {
	categories map[int64]*model.Category
}

func newMockCategoryRepository() *mockCategoryRepository {
	return &mockCategoryRepository{
		categories: make(map[int64]*model.Category),
	}
}

func (m *mockCategoryRepository) Create(ctx context.Context, c *model.Category) error {
	c.ID = int64(len(m.categories) + 1)
	m.categories[c.ID] = c
	return nil
}

func (m *mockCategoryRepository) FindByID(ctx context.Context, id int64) (*model.Category, error) {
	c, exists := m.categories[id]
	if !exists {
		return nil, repository.ErrCategoryNotFound
	}
	return c, nil
}

func (m *mockCategoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	for _, c := range m.categories {
		if c.Slug == slug {
			return c, nil
		}
	}
	return nil, repository.ErrCategoryNotFound
}

func (m *mockCategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	list := make([]model.Category, 0, len(m.categories))
	for _, c := range m.categories {
		list = append(list, *c)
	}
	return list, nil
}

func TestProductService_CreateProduct_Success(t *testing.T) {
	catRepo := newMockCategoryRepository()
	catRepo.categories[1] = &model.Category{ID: 1, Name: "Smartphones", Slug: "smartphones"}

	prodRepo := newMockProductRepository()
	svc := NewProductService(prodRepo, catRepo)

	req := &payloads.CreateProductRequest{
		CategoryID:  1,
		Name:        "iPhone 15 Pro",
		Description: "Titanium design",
		Variants: []payloads.CreateProductVariantRequest{
			{
				SKU:                 "IP15PRO-NAT-128",
				VariantName:         "Natural Titanium 128GB",
				DefaultSellingPrice: decimal.NewFromFloat(18999000),
			},
			{
				SKU:                 "IP15PRO-BLK-256",
				VariantName:         "Black Titanium 256GB",
				DefaultSellingPrice: decimal.NewFromFloat(21999000),
			},
		},
	}

	res, err := svc.CreateProduct(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.Name != "iPhone 15 Pro" {
		t.Errorf("expected product name 'iPhone 15 Pro', got %s", res.Name)
	}
	if len(res.Variants) != 2 {
		t.Errorf("expected 2 variants, got %d", len(res.Variants))
	}
}

func TestProductService_CreateProduct_DuplicateSKU(t *testing.T) {
	catRepo := newMockCategoryRepository()
	catRepo.categories[1] = &model.Category{ID: 1, Name: "Smartphones", Slug: "smartphones"}

	prodRepo := newMockProductRepository()
	svc := NewProductService(prodRepo, catRepo)

	req := &payloads.CreateProductRequest{
		CategoryID: 1,
		Name:       "iPhone 15",
		Variants: []payloads.CreateProductVariantRequest{
			{SKU: "SAME-SKU", VariantName: "V1", DefaultSellingPrice: decimal.NewFromFloat(1000)},
			{SKU: "SAME-SKU", VariantName: "V2", DefaultSellingPrice: decimal.NewFromFloat(2000)},
		},
	}

	_, err := svc.CreateProduct(context.Background(), req)
	if err == nil {
		t.Errorf("expected error for duplicate SKU in payload, got nil")
	}
}
