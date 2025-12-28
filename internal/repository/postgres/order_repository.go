package postgres

import (
	"context"
	"errors"
	"fmt"
	"postgresDB/internal/domain/entities"
	apperror "postgresDB/internal/domain/errors"
	"postgresDB/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type orderRepository struct {
	db *pgxpool.Pool
}

// NewOrderRepository creates a new OrderRepository instance
func NewOrderRepository(db *pgxpool.Pool) repository.OrderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) Create(ctx context.Context, order *entities.Order) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Insert order
	orderQuery := `
		INSERT INTO orders (id, customer_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.Exec(ctx, orderQuery,
		order.ID,
		order.CustomerID,
		order.Status,
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	// Insert order items
	itemQuery := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price,sub_total, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, itemQuery,
			item.ID,
			item.OrderID,
			item.ProductID,
			item.Quantity,
			item.UnitPrice,
			item.SubTotal,
			item.CreatedAt,
		)
		if err != nil {
			return apperror.WrapInternal(err)
		}
	}
	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	return nil
}

// GetByID retrieves an order by its ID (without items)
func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Order, error) {
	query := `SELECT id, customer_id, status, total_amount, created_at, updated_at FROM orders WHERE id = $1`

	var order entities.Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrOrderNotFound
		}
		return nil, apperror.WrapInternal(err)
	}

	return &order, nil
}

// GetByIDWithItems retrieves an order by its ID along with its items
func (r *orderRepository) GetByIDWithItems(ctx context.Context, id uuid.UUID) (*entities.Order, error) {
	order, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := r.GetOrderItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

// GetByCustomerID terima customerID dan mengembalikan daftar pesanan yang terkait dengan customer tersebut
func (r *orderRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int, status string) ([]*entities.Order, int64, error) {
	// Build count query
	countQuery := `SELECT COUNT(*) FROM orders WHERE customer_id = $1`
	args := []interface{}{customerID}
	argsIndex := 2

	// Add status filter if provided
	if status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", argsIndex)
		args = append(args, status)
		argsIndex++
	}

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, apperror.WrapInternal(err)
	}

	// Build main query
	query := `SELECT id, customer_id, status, total_amount, created_at, updated_at FROM orders WHERE customer_id = $1`
	args = []interface{}{customerID}
	argsIndex = 2

	// Add status filter if provided
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argsIndex)
		args = append(args, status)
		argsIndex++
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argsIndex, argsIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperror.WrapInternal(err)
	}
	defer rows.Close()

	orders := make([]*entities.Order, 0, limit)
	for rows.Next() {
		var order entities.Order
		if err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.Status,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, 0, apperror.WrapInternal(err)
		}
		orders = append(orders, &order)
	}

	return orders, total, nil
}

// ListAll retrieves all orders with pagination and optional status filtering
func (r *orderRepository) ListAll(ctx context.Context, limit, offset int, status string) ([]*entities.Order, int64, error) {
	count := `SELECT COUNT(*) FROM orders WHERE 1=1`
	args := make([]interface{}, 0)
	argIndex := 1

	if status != "" {
		count += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	var total int64
	if err := r.db.QueryRow(ctx, count, args...).Scan(&total); err != nil {
		return nil, 0, apperror.WrapInternal(err)
	}
	query := `SELECT id, customer_id, status, total_amount, created_at, updated_at FROM orders WHERE 1=1`
	args = make([]interface{}, 0)
	argIndex = 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperror.WrapInternal(err)
	}
	defer rows.Close()

	orders := make([]*entities.Order, 0, limit)
	for rows.Next() {
		var order entities.Order
		if err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.Status,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, 0, apperror.WrapInternal(err)
		}
		orders = append(orders, &order)
	}

	return orders, total, nil
}

// UpdateStatus updates the status of an order
func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus entities.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`

	res, err := r.db.Exec(ctx, query, newStatus, id)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrOrderNotFound
	}
	return nil
}

// CreateOrderItem buat item pesanan baru
func (r *orderRepository) CreateOrderItem(ctx context.Context, item *entities.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx,
		query,
		item.ID,
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.UnitPrice,
		item.SubTotal,
		item.CreatedAt,
	)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	return nil
}

// GetOrderItemsByOrderID mengambil item pesanan berdasarkan ID pesanan
func (r *orderRepository) GetOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]entities.OrderItem, error) {
	query := `SELECT id, order_id, product_id, quantity, unit_price, created_at, updated_at FROM order_items WHERE order_id = $1 ORDER BY created_at`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, apperror.WrapInternal(err)
	}
	defer rows.Close()

	items := make([]entities.OrderItem, 0)
	for rows.Next() {
		var item entities.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.SubTotal,
			&item.CreatedAt,
		); err != nil {
			return nil, apperror.WrapInternal(err)
		}
		items = append(items, item)
	}

	return items, nil
}
