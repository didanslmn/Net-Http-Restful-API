package dto

import (
	"postgresDB/internal/domain/entities"
)

type CreateOrderRequest struct {
	CustomerID string             `json:"customer_id" validate:"required,uuid4"`
	Items      []OrderItemRequest `json:"items" validate:"required,dive,required"`
	Metadata   map[string]string  `json:"metadata" validate:"omitempty"`
}
type OrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid4"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

// UpdateOrderRequest represents the payload for updating an existing order
type UpdateOrderRequest struct {
	Status *string `json:"status" validate:"omitempty,oneof=pending completed cancelled"`
}

type OrderResponse struct {
	ID          string              `json:"id"`
	CustomerID  string              `json:"customer_id"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

type OrderItemResponse struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	SubTotal  float64 `json:"sub_total"`
	CreatedAt string  `json:"created_at"`
}

// OrderListRequest represents the query parameters for listing orders
type OrderListRequest struct {
	Status string `json:"status" validate:"omitempty,oneof=pending completed cancelled"`
	Limit  int    `json:"limit" validate:"omitempty,min=1,max=100"`
	Page   int    `json:"page" validate:"omitempty,min=1"`
}

// ToOrderResponse converts an Order entity to OrderResponse DTO
func ToOrderResponse(o *entities.Order) OrderResponse {
	items := make([]OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = OrderItemResponse{
			ID:        item.ID.String(),
			ProductID: item.ProductID.String(),
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			SubTotal:  item.SubTotal,
			CreatedAt: item.CreatedAt,
		}
	}

	return OrderResponse{
		ID:          o.ID.String(),
		CustomerID:  o.CustomerID.String(),
		Status:      string(o.Status),
		TotalAmount: o.TotalAmount,
		Items:       items,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

// ToOrderResponse converts a list of Order entities to responses
func ToOrderResponseList(orders []*entities.Order) []OrderResponse {
	responses := make([]OrderResponse, len(orders))
	for i, o := range orders {
		responses[i] = ToOrderResponse(o)
	}
	return responses
}
