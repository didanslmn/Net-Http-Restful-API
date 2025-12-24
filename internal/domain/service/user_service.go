package service

import (
	"context"
	"time"

	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/entities"

	"github.com/google/uuid"
)

type UserService interface {
	// Define service methods here
	GetUser(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role) (*dto.UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, req dto.ChangePasswordRequest) error
	Delete(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role) error
}

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, accessJTI string, accessExp time.Time, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
}
