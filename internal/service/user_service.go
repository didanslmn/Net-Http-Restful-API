package service

import (
	"context"
	"time"

	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/entities"
	apperror "postgresDB/internal/domain/errors"
	"postgresDB/internal/domain/repository"
	"postgresDB/internal/domain/service"
	"postgresDB/pkg/utils"

	"github.com/google/uuid"
)

// UserServiceImpl implements the UserService interface
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepository) service.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role) (*dto.UserResponse, error) {
	// Additional authorization: User can get own profile, Admin can get any profile
	if requesterRole != entities.RoleAdmin && requesterID != id {
		return nil, apperror.ErrUnauthorized
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	response := dto.ToUserResponse(user)
	return &response, nil
}

// Update updates user information
func (s *userService) Update(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// Authorization: Only Admin or the user themselves can update
	if requesterRole != entities.RoleAdmin && requesterID != id {
		return nil, apperror.ErrUnauthorized
	}

	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}

	// only admin can update role and active status
	// if req.Role != nil {
	// 	if requesterRole != entities.RoleAdmin {
	// 		return nil, apperror.ErrUnauthorized
	// 	}
	// 	existingUser.Role = entities.Role(*req.Role)
	// } karena disini saya pakai role hanya admin dan user saja maka saya hapus atau koemntarkan bagian ini
	if req.IsActive != nil {
		if requesterRole != entities.RoleAdmin {
			return nil, apperror.ErrUnauthorized
		}
		existingUser.IsActive = *req.IsActive
	}

	// Check for uniqueness if email is being updated
	if req.Email != nil && *req.Email != existingUser.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrUserAlreadyExists
		}
		existingUser.Email = *req.Email
	}

	// Check for uniqueness if username is being updated
	if req.Username != nil && *req.Username != existingUser.Username {
		exists, err := s.userRepo.ExistsByUsername(ctx, *req.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrUserAlreadyExists
		}
		existingUser.Username = *req.Username
	}

	// Update timestamp
	existingUser.UpdatedAt = time.Now()

	// Save changes
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, err
	}

	// Return updated user response
	response := dto.ToUserResponse(existingUser)
	return &response, nil
}

// ChangePassword changes user password
func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, req dto.ChangePasswordRequest) error {
	// Authoritaion: User can change only their own password
	if id != requesterID {
		return apperror.ErrUnauthorized
	}

	// Get user
	userEntity, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if userEntity == nil {
		return apperror.ErrUserNotFound
	}

	// Verify old password
	if err := utils.CheckPassword(req.OldPassword, userEntity.Password); err != nil {
		return apperror.ErrPasswordMismatch
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	if hashedPassword == "" {
		return apperror.ErrPasswordTooWeak
	}

	// Update password
	userEntity.Password = hashedPassword
	userEntity.UpdatedAt = time.Now()

	// Save changes
	return s.userRepo.Update(ctx, userEntity)
}

// Delete deactivates a user account
func (s *userService) Delete(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole entities.Role) error {
	// Authorization: Only Admin or the user themselves can delete
	if requesterRole != entities.RoleAdmin && requesterID != id {
		return apperror.ErrUnauthorized
	}
	// Get user
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Save changes
	return s.userRepo.Delete(ctx, id)
}
