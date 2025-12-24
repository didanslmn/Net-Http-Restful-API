package entities

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product entity in the system

type Product struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	Category    string    `db:"category"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
