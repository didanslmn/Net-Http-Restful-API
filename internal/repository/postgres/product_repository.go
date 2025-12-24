package repository

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

type productRepository struct {
	// db connection or other dependencies can be added here
	db *pgxpool.Pool
}

// NewProductRepository untuk membuat instance baru dari ProductRepository
func NewProductRepository(db *pgxpool.Pool) repository.ProductRepository {
	return &productRepository{
		db: db,
	}
}

// Create produk baru
func (r *productRepository) Create(ctx context.Context, product *entities.Product) error {
	// implementasi pembuatan produk di database
	query := `
		INSERT INTO products (id, name, description, price, stock, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, product.ID, product.Name, product.Description, product.Price, product.Stock, product.Category)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	return nil
}

// GetByID mengambil produk berdasarkan ID
func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	// implementasi pengambilan produk dari database berdasarkan ID
	query := `SELECT id, name, description, price, stock, category, created_at, updated_at FROM products WHERE id = $1`

	// Scan the result into a Product entity
	var product entities.Product
	var description, category *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&description,
		&product.Price,
		&product.Stock,
		&category,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrProductNotFound
		}
		return nil, apperror.WrapInternal(err)
	}

	// Assign nullable fields
	if description != nil {
		product.Description = *description
	}
	if category != nil {
		product.Category = *category
	}

	return &product, nil
}

// List mengambil daftar produk dengan pagination dan filter
func (r *productRepository) List(ctx context.Context, limit, offset int, search, category string) ([]*entities.Product, int64, error) {
	//build count query
	countQuery := `SELECT COUNT(*) FROM products WHERE 1=1`
	args := make([]interface{}, 0)
	argIndex := 1

	if search != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}
	if category != "" {
		countQuery += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	// End count query
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, apperror.WrapInternal(err)
	}

	// Build main query
	query := `
		SELECT id, name, description, price, stock, category, created_at, updated_at
		FROM products
		WHERE 1=1
	`

	args = make([]interface{}, 0)
	argIndex = 1

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	if search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperror.WrapInternal(err)
	}
	defer rows.Close()

	products := make([]*entities.Product, 0, limit)
	for rows.Next() {
		var product entities.Product
		var description, categoryVal *string

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&description,
			&product.Price,
			&product.Stock,
			&categoryVal,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, 0, apperror.WrapInternal(err)
		}

		if description != nil {
			product.Description = *description
		}
		if categoryVal != nil {
			product.Category = *categoryVal
		}

		products = append(products, &product)
	}

	return products, total, nil
}

// Update mengupdate data produk
func (r *productRepository) Update(ctx context.Context, product *entities.Product) error {
	// implementasi update produk di database
	query := `UPDATE products SET name = $1, description = $2, price = $3, stock = $4, category = $5, updated_at = NOW() WHERE id = $6`

	// Execute the query
	res, err := r.db.Exec(ctx, query, product.Name, product.Description, product.Price, product.Stock, product.Category, product.ID)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrProductNotFound
	}
	return nil
}

// Delete menghapus produk berdasarkan ID
func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// implementasi penghapusan produk di database
	query := `DELETE FROM products WHERE id = $1`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrProductNotFound
	}
	return nil
}

// UpdateStock memperbarui stok produk
func (r *productRepository) UpdateStock(ctx context.Context, id uuid.UUID, newStock int) error {
	// implementasi pembaruan stok produk di database
	query := `UPDATE products SET stock = stock + $1, updated_at = NOW() WHERE id = $2 AND stock + $1 >= 0`

	res, err := r.db.Exec(ctx, query, newStock, id)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrInsufficientStock
	}
	return nil
}
