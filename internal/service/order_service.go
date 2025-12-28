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

type orderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository) service.OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *orderService) Create(ctx context.Context, customerID uuid.UUID, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Create order entity
	order := &entities.Order{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Status:      entities.OrderStatusPending,
		TotalAmount: 0,
		Items:       []entities.OrderItem{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	// Validate and create order items
	for _, itemReq := range req.Items {
		// Check product existence and stock
		product, err := s.productRepo.GetByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, err
		}

		if product.Stock < itemReq.Quantity {
			return nil, apperror.ErrInsufficientStock
		}
		// Create order item
		orderItem := entities.OrderItem{
			ID:        uuid.New(),
			OrderID:   order.ID,
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
			UnitPrice: product.Price,
			SubTotal:  product.Price * float64(itemReq.Quantity),
			CreatedAt: time.Now(),
		}

		// Append order item to order
		order.Items = append(order.Items, orderItem)
		order.TotalAmount += orderItem.SubTotal
		// Update product stock
		product.Stock -= itemReq.Quantity
		if err := s.productRepo.UpdateStock(ctx, product.ID, product.Stock); err != nil {
			return nil, err
		}

	}
	// Save order to repository

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	response := dto.ToOrderResponse(order)
	return &response, nil
}

func (s *orderService) GetByID(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role) (*dto.OrderResponse, error) {
	// Get order by ID
	order, err := s.orderRepo.GetByIDWithItems(ctx, id)
	if err != nil {
		return nil, err
	}
	// check authorization: admin or owner can view any, customers can view their own
	if requesterRole == entities.RoleUser && order.CustomerID != requesterID {
		return nil, apperror.ErrForbidden
	}
	response := dto.ToOrderResponse(order)
	return &response, nil
}

func (s *orderService) ListAll(ctx context.Context, UserID uuid.UUID, requesterRole entities.Role, req dto.OrderListRequest) ([]dto.OrderResponse, *dto.PaginationMeta, error) {
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

	// Fetch orders based on role
	var orders []*entities.Order
	var total int64
	var err error

	// Admin can see all orders, users can see their own orders
	if requesterRole == entities.RoleAdmin {
		orders, total, err = s.orderRepo.ListAll(ctx, limit, offset, req.Status)
	} else {
		orders, total, err = s.orderRepo.GetByCustomerID(ctx, UserID, limit, offset, req.Status)
	}
	if err != nil {
		return nil, nil, err
	}
	responseList := dto.ToOrderResponseList(orders)
	pagination := &dto.PaginationMeta{
		Total:      total,
		Limit:      limit,
		Page:       page,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}
	return responseList, pagination, nil
}

func (s *orderService) UpdateStatus(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role, req *dto.UpdateOrderRequest) (*dto.OrderResponse, error) {
	// Only admin can update order status
	if requesterRole != entities.RoleAdmin {
		return nil, apperror.ErrForbidden
	}

	// Get existing order
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// parse and validate new status
	newStatus := entities.OrderStatus(req.Status)
	if !newStatus.IsValid() {
		return nil, apperror.NewValidationError([]apperror.ValidationError{
			{Field: "Status", Message: "status tidak valid"},
		})
	}

	// check if transition is valid
	if !order.Status.CanTransitionTo(newStatus) {
		return nil, apperror.ErrInvalidStatusTransition
	}

	// if canceling, restock products
	if newStatus == entities.OrderStatusCancelled {
		for _, item := range order.Items {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				return nil, err
			}
			product.Stock += item.Quantity
			if err := s.productRepo.UpdateStock(ctx, product.ID, product.Stock); err != nil {
				return nil, err
			}
		}
	}

	// update order status
	if err := s.orderRepo.UpdateStatus(ctx, id, newStatus); err != nil {
		return nil, err
	}

	// get updated order with items
	updatedOrder, err := s.orderRepo.GetByIDWithItems(ctx, id)
	if err != nil {
		return nil, err
	}

	response := dto.ToOrderResponse(updatedOrder)
	return &response, nil
}
