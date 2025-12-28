package service

import (
	"context"
	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/entities"

	"github.com/google/uuid"
)

type ProductService interface {
	Create(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error)
	List(ctx context.Context, req dto.ProductListRequest) ([]dto.ProductResponse, *dto.PaginationMeta, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateProductRequest, userRole entities.Role) (*dto.ProductResponse, error)
	Delete(ctx context.Context, id uuid.UUID, userRole entities.Role) error
}
