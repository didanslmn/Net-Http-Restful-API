package service

import (
	"context"
	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/entities"

	"github.com/google/uuid"
)

type OrderService interface {
	Create(ctx context.Context, customerID uuid.UUID, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetByID(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role) (*dto.OrderResponse, error)
	//GetByCustomerID(ctx context.Context, customerID uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role, req dto.OrderListRequest) ([]dto.OrderResponse, *dto.PaginationMeta, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role, req *dto.UpdateOrderRequest) (*dto.OrderResponse, error)
	ListAll(ctx context.Context, UserID uuid.UUID, requesterRole entities.Role, req dto.OrderListRequest) ([]dto.OrderResponse, *dto.PaginationMeta, error)
}
