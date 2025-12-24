package middleware

import (
	"context"
	"net/http"
	"postgresDB/internal/delivery/response"
	"postgresDB/internal/domain/entities"
	apperror "postgresDB/internal/domain/errors"
	"postgresDB/pkg/jwt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Context keys
type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
	TokenJTIKey contextKey = "token_jti"
	TokenExpKey contextKey = "token_exp"
)

// Auth Middleware validates authentication and authorization JWT tokens
func Auth(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, apperror.ErrUnauthorized)
				return
			}

			// Expecting header format: "Bearer <token>"
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				response.Error(w, apperror.ErrUnauthorized)
				return
			}

			// Validate token
			claims, err := jwtService.ValidateAccessToken(r.Context(), tokenString)
			if err != nil {
				response.Error(w, apperror.ErrInvalidToken)
				return
			}

			// Add claims to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, TokenJTIKey, claims.ID)
			ctx = context.WithValue(ctx, TokenExpKey, claims.ExpiresAt.Time)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRoles ...entities.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user role from context
			role, ok := r.Context().Value(UserRoleKey).(entities.Role)
			if !ok {
				response.Error(w, apperror.ErrUnauthorized)
				return
			}

			// Check if user role is in required roles
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					next.ServeHTTP(w, r)
					return
				}
			}

			// If role not authorized
			response.Error(w, apperror.ErrForbidden)
		})
	}
}

// GetUserIDFromContext retrieves user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, apperror.ErrUnauthorized
	}
	return userID, nil
}

// GetUserRoleFromContext retrieves user role from context
func GetUserRole(ctx context.Context) (entities.Role, error) {
	role, ok := ctx.Value(UserRoleKey).(entities.Role)
	if !ok {
		return "", apperror.ErrUnauthorized
	}
	return role, nil
}

// GetTokenJTIFromContext retrieves token JTI from context
func GetTokenJTIFromContext(ctx context.Context) (string, error) {
	jti, ok := ctx.Value(TokenJTIKey).(string)
	if !ok {
		return "", apperror.ErrUnauthorized
	}
	return jti, nil
}

// GetTokenExpFromContext retrieves token expiration time from context
func GetTokenExpFromContext(ctx context.Context) (time.Time, error) {
	exp, ok := ctx.Value(TokenExpKey).(time.Time)
	if !ok {
		return time.Time{}, apperror.ErrUnauthorized
	}
	return exp, nil
}
