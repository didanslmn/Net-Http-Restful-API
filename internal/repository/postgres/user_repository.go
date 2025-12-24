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

type userRepository struct {
	// Add necessary fields here, e.g., database connections
	db *pgxpool.Pool
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{
		db: db,
	}
}

// CreateUser inserts a new user into the database
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (id, username, email, password, role, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, NOW(),NOW())
		`
	_, err := r.db.Exec(ctx, query, user.ID, user.Username, user.Email, user.Password, user.Role, user.IsActive)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.ErrEmailExists
		}
		return apperror.WrapInternal(err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {

	// Implement the logic to get a user by ID from the database
	query := `SELECT id, username, email, password, role, is_active, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	// Scan the result into a User entity
	var u entities.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, apperror.WrapInternal(err)
	}
	return &u, nil
}

// GetByEmail retrieves a user by their email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {

	// Implement the logic to get a user by email from the database
	query := `SELECT id, username, email, password, role, is_active, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)

	// Scan the result into a User entity
	var u entities.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, apperror.WrapInternal(err)
	}
	return &u, nil

}

// GetByUsername retrieves a user by their username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	// Implement the logic to get a user by username from the database
	query := `SELECT id, username, email, password, role, is_active, created_at, updated_at FROM users WHERE username = $1`
	row := r.db.QueryRow(ctx, query, username)
	// Scan the result into a User entity
	var u entities.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil

}

// GetByEmailOrUsername retrieves a user by their email or username
func (r *userRepository) GetByEmailOrUsername(ctx context.Context, loginID string) (*entities.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at
		FROM users 
		WHERE email = $1 OR username = $1
		LIMIT 1
	`

	var u entities.User
	err := r.db.QueryRow(ctx, query, loginID).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by login_id: %w", err)
	}
	return &u, nil
}

// ExistsByEmail
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, apperror.WrapInternal(err)
	}
	return exists, nil
}

// ExistsByUsername
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, apperror.WrapInternal(err)
	}
	return exists, nil
}

// UpdateUser updates an existing user in the database
func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	query := `UPDATE users SET username = $1, email = $2, role = $3, is_active = $4, updated_at = NOW() WHERE id = $5`
	res, err := r.db.Exec(ctx, query, user.Username, user.Email, user.Role, user.IsActive, user.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.ErrEmailExists
		}
		return apperror.WrapInternal(err)
	}

	if res.RowsAffected() == 0 {
		return apperror.ErrUserNotFound
	}
	return nil
}

// DeleteUser removes a user from the database by their ID
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return apperror.WrapInternal(err)
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrUserNotFound
	}
	return nil
}

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "23505") || contains(errStr, "unique")
}
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}
func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
