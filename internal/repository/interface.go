package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TokenRepository defines token management methods (Redis)
type TokenRepository interface {
	// BlacklistToken adds a token JTI to the blacklist
	BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error
	// IsTokenBlacklisted checks if a token JTI is blacklisted
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	// SetTokenFamily stores the current JTI for a token family
	SetTokenFamily(ctx context.Context, userID uuid.UUID, family, jti string, ttl time.Duration) error
	// GetTokenFamily gets the current JTI for a token family
	GetTokenFamily(ctx context.Context, userID uuid.UUID, family string) (string, error)
	// TrackUserSession tracks a user's session (token family)
	TrackUserSession(ctx context.Context, userID uuid.UUID, family string, ttl time.Duration) error
	// RevokeAllUserSessions revokes all sessions for a user
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
}
