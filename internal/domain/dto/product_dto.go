package dto

import (
	"postgresDB/internal/domain/entities"
	"time"
)

// CreateProductRequest represents the payload for creating a new product
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Stock       int     `json:"stock" validate:"required,min=0"`
	Category    string  `json:"category" validate:"required"`
}

// UpdateProductRequest represents the payload for updating an existing product
type UpdateProductRequest struct {
	Name        *string  `json:"name" validate:"omitempty"`
	Description *string  `json:"description" validate:"omitempty"`
	Price       *float64 `json:"price" validate:"omitempty,min=0"`
	Stock       *int     `json:"stock" validate:"omitempty,min=0"`
	Category    *string  `json:"category" validate:"omitempty"`
}

// ProductResponse represents the product data returned in responses
type ProductResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ProductListRequest represents the query parameters for listing products
type ProductListRequest struct {
	Category string `json:"category" validate:"omitempty"`
	Limit    int    `json:"limit" validate:"omitempty,min=1,max=100"`
	Page     int    `json:"page" validate:"omitempty,min=1"`
	Search   string `json:"search" validate:"omitempty"`
}

// ToProductResponse converts a Product entity to ProductResponse DTO
func ToProductResponse(p *entities.Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Category:    p.Category,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

// ToProductResponseList converts a slice of Product entities to a slice of ProductResponse DTOs
func ToProductResponseList(products []*entities.Product) []ProductResponse {
	responses := make([]ProductResponse, len(products))
	for i, p := range products {
		responses[i] = ToProductResponse(p)
	}
	return responses
}
