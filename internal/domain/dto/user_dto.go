package dto

import (
	"postgresDB/internal/domain/entities"
	"time"
)

// LoginRequest represents the payload for user login
type LoginRequest struct {
	LoginID  string `json:"login_id" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest represents the payload for user registration
type RegisterRequest struct {
	Username        string `json:"username" validate:"required,username"`
	Email           string `json:"email" validate:"required,customEmail"`
	Password        string `json:"password" validate:"required,strongPassword"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// UpdateUserRequest represents the payload for updating user information
type UpdateUserRequest struct {
	Username *string `json:"username" validate:"omitempty,username"`
	Email    *string `json:"email" validate:"omitempty,customEmail"`
	IsActive *bool   `json:"is_active" validate:"omitempty"`
}

// ChangePasswordRequest represents the payload for changing user password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=8"`
	NewPassword string `json:"new_password" validate:"required,strongPassword"`
}

// Response represents the user data returned in responses
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type AuthResponse struct {
	AccessToken  string       `json:"token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         UserResponse `json:"user"`
}

type RegisterResponse struct {
	Message      string       `json:"message"`
	Token        string       `json:"token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         UserResponse `json:"user"`
}

// Func Response creates a new Response instance
func ToUserResponse(u *entities.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func ToAuthResponse(access string, refresh string, u *entities.User) AuthResponse {
	return AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         ToUserResponse(u),
	}
}

func ToRegisterResponse(message, accessToken, refreshToken string, u *entities.User) RegisterResponse {
	return RegisterResponse{
		Message:      message,
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         ToUserResponse(u),
	}
}
