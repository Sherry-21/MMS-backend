package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
)

var (
	ErrCategoryNameRequired = errors.New("category name is required")
)

// CategoryService defines business logic for categories
type CategoryService interface {
	CreateCategory(ctx context.Context, req *payloads.CreateCategoryRequest) (*payloads.CategoryResponse, error)
	ListCategories(ctx context.Context) ([]payloads.CategoryResponse, error)
}

type categoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService creates a new CategoryService instance
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, req *payloads.CreateCategoryRequest) (*payloads.CategoryResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrCategoryNameRequired
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = generateSlug(name)
	} else {
		slug = generateSlug(slug)
	}

	// Check if slug exists
	existing, err := s.repo.FindBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, repository.ErrCategorySlugExists
	}
	if err != nil && !errors.Is(err, repository.ErrCategoryNotFound) {
		return nil, fmt.Errorf("failed to check category slug: %w", err)
	}

	category := &model.Category{
		Name: name,
		Slug: slug,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to persist category: %w", err)
	}

	return &payloads.CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		CreatedAt: category.CreatedAt,
	}, nil
}

func (s *categoryService) ListCategories(ctx context.Context) ([]payloads.CategoryResponse, error) {
	categories, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]payloads.CategoryResponse, len(categories))
	for i, c := range categories {
		result[i] = payloads.CategoryResponse{
			ID:        c.ID,
			Name:      c.Name,
			Slug:      c.Slug,
			CreatedAt: c.CreatedAt,
		}
	}

	return result, nil
}

// Helper to convert text to URL-friendly kebab-case slug
func generateSlug(input string) string {
	lower := strings.ToLower(input)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}
