package repository

import (
	"context"
	"postgresDB/internal/domain/entities"

	"github.com/google/uuid"
)

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, order *entities.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Order, error)
	GetByIDWithItems(ctx context.Context, id uuid.UUID) (*entities.Order, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int, status string) ([]*entities.Order, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, newStatus entities.OrderStatus) error
	CreateOrderItem(ctx context.Context, item *entities.OrderItem) error
	GetOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]entities.OrderItem, error)
	ListAll(ctx context.Context, limit, offset int, status string) ([]*entities.Order, int64, error)
}
