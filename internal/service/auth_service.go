package service

import (
	"context"
	"errors"
	"time"

	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/entities"
	apperror "postgresDB/internal/domain/errors"
	"postgresDB/internal/domain/repository"
	"postgresDB/internal/domain/service"
	"postgresDB/pkg/jwt"
	"postgresDB/pkg/utils"

	"github.com/google/uuid"
)

// AuthServiceImpl implements the AuthService interface
type authService struct {
	userRepo   repository.UserRepository
	jwtService *jwt.JWTService
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repository.UserRepository, jwtService *jwt.JWTService) service.AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Register handles user registration
func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Check if email or username already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrUserAlreadyExists
	}

	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, apperror.WrapInternal(err)
	}

	// Determine role - default to 'user' if not provided
	role := entities.RoleUser

	// Create user entity
	now := time.Now().UTC()
	newUser := &entities.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      role,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save user
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	// Response
	return &dto.RegisterResponse{
		Message: "User registered successfully",
		User:    dto.ToUserResponse(newUser),
	}, nil
}

// Login handles user authentication
func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email or username
	userEntity, err := s.userRepo.GetByEmailOrUsername(ctx, req.LoginID)
	if err != nil {
		if errors.Is(err, apperror.ErrUserNotFound) {
			return nil, apperror.ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !userEntity.IsActive {
		return nil, apperror.ErrUserInactive
	}

	// Verify password
	if err := utils.CheckPassword(req.Password, userEntity.Password); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	// Generate tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(ctx, userEntity.ID, userEntity.Role)
	if err != nil {
		return nil, apperror.WrapInternal(err)
	}

	// Create response
	return &dto.AuthResponse{
		User:         dto.ToUserResponse(userEntity),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken handles token refresh
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	// Validate refresh token
	tokenPair, err := s.jwtService.RefreshTokens(ctx, refreshToken)
	if err != nil {
		return nil, apperror.ErrInvalidToken.WithError(err)
	}

	// Get user
	claims, err := s.jwtService.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		return nil, apperror.ErrInvalidToken.WithError(err)
	}

	userEntity, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Check if user is active
	if !userEntity.IsActive {
		return nil, apperror.ErrUserInactive
	}

	// Create response
	return &dto.AuthResponse{
		User:         dto.ToUserResponse(userEntity),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RevokeAllSession revokes all sessions for a user
func (s *authService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	if err := s.jwtService.RevokeAllUserSessions(ctx, userID); err != nil {
		return apperror.WrapInternal(err)
	}
	return nil
}

// Logout handles user logout by blacklisting the access token and refresh token
func (s *authService) Logout(ctx context.Context, accessJTI string, accessExp time.Time, refreshToken string) error {
	accessTTL := time.Until(accessExp)
	// Blacklist access token
	if err := s.jwtService.BlacklistToken(ctx, accessJTI, accessTTL); err != nil {
		return apperror.WrapInternal(err)
	}

	// if refresh token is provided, blacklist it as well
	if refreshToken != "" {
		refreshClaims, err := s.jwtService.ValidateToken(refreshToken)
		if err != nil {
			return apperror.ErrInvalidToken.WithError(err)
		}
		refreshTTL := time.Until(refreshClaims.ExpiresAt.Time)
		if err := s.jwtService.BlacklistToken(ctx, refreshClaims.ID, refreshTTL); err != nil {
			return apperror.WrapInternal(err)
		}
	}

	return nil
}
