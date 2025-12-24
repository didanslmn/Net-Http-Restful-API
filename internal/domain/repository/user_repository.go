package repository

import (
	"context"
	"postgresDB/internal/domain/entities"

	"github.com/google/uuid"
)

type UserRepository interface {
	// Define repository methods here
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	GetByEmailOrUsername(ctx context.Context, loginID string) (*entities.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
