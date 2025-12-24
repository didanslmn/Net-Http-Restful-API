package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"postgresDB/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	blacklistPrefix    = "jwt:blacklist:"
	tokenFamilyPrefix  = "jwt:family:"
	userSessionsPrefix = "jwt:sessions:"
)

// tokenRepository implements repository.TokenRepository
type tokenRepository struct {
	client *redis.Client
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(client *redis.Client) repository.TokenRepository {
	return &tokenRepository{client: client}
}

// BlacklistToken adds a token JTI to the blacklist
func (r *tokenRepository) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil // Token already expired
	}
	key := blacklistPrefix + jti
	return r.client.Set(ctx, key, "1", ttl).Err()
}

// IsTokenBlacklisted checks if a token JTI is blacklisted
func (r *tokenRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := blacklistPrefix + jti
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetTokenFamily stores the current JTI for a token family
func (r *tokenRepository) SetTokenFamily(ctx context.Context, userID uuid.UUID, family, jti string, ttl time.Duration) error {
	key := fmt.Sprintf("%s%s:%s", tokenFamilyPrefix, userID.String(), family)
	return r.client.Set(ctx, key, jti, ttl).Err()
}

// GetTokenFamily gets the current JTI for a token family
func (r *tokenRepository) GetTokenFamily(ctx context.Context, userID uuid.UUID, family string) (string, error) {
	key := fmt.Sprintf("%s%s:%s", tokenFamilyPrefix, userID.String(), family)
	result, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return result, err
}

// TrackUserSession tracks a user's session (token family)
func (r *tokenRepository) TrackUserSession(ctx context.Context, userID uuid.UUID, family string, ttl time.Duration) error {
	key := userSessionsPrefix + userID.String()
	score := float64(time.Now().Add(ttl).Unix())
	return r.client.ZAdd(ctx, key, redis.Z{Score: score, Member: family}).Err()
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *tokenRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	// Get all session families for the user
	key := userSessionsPrefix + userID.String()
	families, err := r.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	// Delete all family tokens
	for _, family := range families {
		familyKey := fmt.Sprintf("%s%s:%s", tokenFamilyPrefix, userID.String(), family)
		r.client.Del(ctx, familyKey)
	}

	// Clear the sessions set
	r.client.Del(ctx, key)

	return nil
}
