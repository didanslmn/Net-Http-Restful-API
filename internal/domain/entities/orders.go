package entities

import "github.com/google/uuid"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// isValid checks if the OrderStatus is valid
func (os OrderStatus) IsValid() bool {
	switch os {
	case OrderStatusPending, OrderStatusCompleted, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation of OrderStatus
func (os OrderStatus) String() string {
	return string(os)
}

type Order struct {
	ID          uuid.UUID   `db:"id"`
	CustomerID  uuid.UUID   `db:"customer_id"`
	Status      OrderStatus `db:"status"`
	TotalAmount float64     `db:"total_amount"`
	Items       []OrderItem `db:"items"`
	CreatedAt   string      `db:"created_at"`
	UpdatedAt   string      `db:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `db:"id"`
	OrderID   uuid.UUID `db:"order_id"`
	ProductID uuid.UUID `db:"product_id"`
	Quantity  int       `db:"quantity"`
	UnitPrice float64   `db:"price"`
	SubTotal  float64   `db:"sub_total"`
	CreatedAt string    `db:"created_at"`
}
