package repository

import (
	"context"
	"postgresDB/internal/domain/entities"

	"github.com/google/uuid"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)
	Update(ctx context.Context, product *entities.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int, search, category string) ([]*entities.Product, int64, error)
	UpdateStock(ctx context.Context, id uuid.UUID, newStock int) error
}
