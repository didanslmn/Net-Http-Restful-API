package service

import (
	"context"
	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/entities"
	apperror "postgresDB/internal/domain/errors"
	"postgresDB/internal/domain/repository"
	"postgresDB/internal/domain/service"
	"time"

	"github.com/google/uuid"
)

type productService struct {
	productRepo repository.ProductRepository
}

// NewProductService creates a new ProductService instance
func NewProductService(productRepo repository.ProductRepository) service.ProductService {
	return &productService{
		productRepo: productRepo,
	}
}

// Create creates a new product
func (s *productService) Create(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Create product entity
	product := &entities.Product{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save product to repository
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	// Return response DTO
	response := dto.ToProductResponse(product)
	return &response, nil
}

// GetByID retrieves a product by its ID
func (s *productService) GetByID(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := dto.ToProductResponse(product)
	return &response, nil
}

// List retrieves a list of products with pagination and optional filtering
func (s *productService) List(ctx context.Context, req dto.ProductListRequest) ([]dto.ProductResponse, *dto.PaginationMeta, error) {
	// set default pagination values
	page := req.Page
	if page <= 1 {
		page = 1
	}

	limit := req.Limit
	if limit <= 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	products, total, err := s.productRepo.List(ctx, limit, offset, req.Search, req.Category)
	if err != nil {
		return nil, nil, err
	}

	responseList := dto.ToProductResponseList(products)
	pagination := &dto.PaginationMeta{
		Total:      total,
		Limit:      limit,
		Page:       page,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	return responseList, pagination, nil
}

// Update updates an existing product
func (s *productService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateProductRequest, UserID entities.Role) (*dto.ProductResponse, error) {
	// get existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// check authorization hanya admin yang bisa update product
	if UserID != entities.RoleAdmin {
		return nil, apperror.ErrUnauthorized
	}

	// Update fields if provided
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.Category != nil {
		product.Category = *req.Category
	}

	// Save updated product
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	// Return response DTO
	response := dto.ToProductResponse(product)
	return &response, nil
}

// Delete deletes a product by its ID
func (s *productService) Delete(ctx context.Context, id uuid.UUID, userRole entities.Role) error {
	// Get existing product
	_, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check authorization: only admin can delete
	if userRole != entities.RoleAdmin {
		return apperror.ErrUnauthorized
	}

	// Delete product

	return s.productRepo.Delete(ctx, id)
}
