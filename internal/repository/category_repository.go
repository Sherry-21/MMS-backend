package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"maiora-backend-service/internal/model"
)

var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategorySlugExists  = errors.New("category slug already exists")
)

// CategoryRepository defines database operations for product categories
type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	FindByID(ctx context.Context, id int64) (*model.Category, error)
	FindBySlug(ctx context.Context, slug string) (*model.Category, error)
	List(ctx context.Context) ([]model.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new CategoryRepository instance
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *model.Category) error {
	if r.db == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id int64) (*model.Category, error) {
	if r.db == nil {
		return nil, ErrCategoryNotFound
	}
	var category model.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	if r.db == nil {
		return nil, ErrCategoryNotFound
	}
	var category model.Category
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) List(ctx context.Context) ([]model.Category, error) {
	if r.db == nil {
		return []model.Category{}, nil
	}
	var categories []model.Category
	err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}
