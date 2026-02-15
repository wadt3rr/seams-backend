package services

import (
	"context"
	"fmt"

	"seams-backend/internal/models"

	"github.com/google/uuid"
)

// ProductService содержит бизнес-логику для работы с продуктами
type ProductService struct {
	storage Storage
}

// Storage интерфейс для работы с хранилищем
type Storage interface {
	GetProductsByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]*models.Product, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
	GetProducts(ctx context.Context, limit, offset int) ([]*models.Product, error)
	SearchProducts(ctx context.Context, query string, limit, offset int) ([]*models.Product, error)
}

// NewProductService создает новый сервис для работы с продуктами
func NewProductService(storage Storage) *ProductService {
	return &ProductService{storage: storage}
}

// ListProductsByCategorySlug возвращает список продуктов по слагу категории
func (s *ProductService) GetProductsByCategorySlug(
	ctx context.Context,
	slug string,
	page, limit int,
) ([]*models.Product, error) {
	category, err := s.storage.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	products, err := s.storage.GetProductsByCategoryID(ctx, category.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) GetProducts(
	ctx context.Context,
	search string,
	page, limit int,
) ([]*models.Product, error) {
	const op = "services.ProductService.GetProducts"
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	if search != "" {
		products, err := s.storage.SearchProducts(ctx, search, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return products, nil
	}

	products, err := s.storage.GetProducts(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return products, nil
}
