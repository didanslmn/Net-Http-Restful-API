package entities

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// IsValid checks if the order status is valid
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusPaid, OrderStatusShipped, OrderStatusCompleted, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation of OrderStatus
func (s OrderStatus) String() string {
	return string(s)
}

// CanTransitionTo checks if status can transition to the target status
func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:   {OrderStatusPaid, OrderStatusCancelled},
		OrderStatusPaid:      {OrderStatusShipped, OrderStatusCancelled},
		OrderStatusShipped:   {OrderStatusCompleted},
		OrderStatusCompleted: {},
		OrderStatusCancelled: {},
	}

	allowedTransitions, exists := transitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == target {
			return true
		}
	}
	return false
}

type Order struct {
	ID          uuid.UUID   `db:"id"`
	CustomerID  uuid.UUID   `db:"customer_id"`
	Status      OrderStatus `db:"status"`
	TotalAmount float64     `db:"total_amount"`
	Items       []OrderItem `db:"items"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `db:"id"`
	OrderID   uuid.UUID `db:"order_id"`
	ProductID uuid.UUID `db:"product_id"`
	Quantity  int       `db:"quantity"`
	UnitPrice float64   `db:"unit_price"`
	SubTotal  float64   `db:"subtotal"`
	CreatedAt time.Time `db:"created_at"`
}
