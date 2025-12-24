// pkg/jwt/manager.go
package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"postgresDB/config"
	"postgresDB/internal/domain/entities"
	"postgresDB/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType untuk membedakan access & refresh
const (
	// TokenTypeAccess represents access token type
	TokenTypeAccess = "access"
	// TokenTypeRefresh represents refresh token type
	TokenTypeRefresh = "refresh"
)

// Claims custom
type Claims struct {
	UserID      uuid.UUID     `json:"user_id"`
	Role        entities.Role `json:"role"`
	TokenType   string        `json:"token_type"`
	TokenFamily string        `json:"token_family,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair Represent access and refresh token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessJTI    string
	RefreshJTI   string
	TokenFamily  string
	ExpiresAt    time.Time
}

// JWTManager – immutable & thread-safe
type JWTService struct {
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
	audience        string
	tokenRepo       repository.TokenRepository
}

func NewService(cfg *config.JWTConfig, tokenRepo repository.TokenRepository) (*JWTService, error) {
	privateKey, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey, err := loadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &JWTService{
		privateKey:      privateKey,
		publicKey:       publicKey,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		issuer:          cfg.Issuer,
		audience:        cfg.Audience,
		tokenRepo:       tokenRepo,
	}, nil
}

// GenerateTokenPair generates new access and refresh tokens
func (s *JWTService) GenerateTokenPair(ctx context.Context, userID uuid.UUID, role entities.Role) (*TokenPair, error) {
	tokenFamily := uuid.New().String()
	return s.generateTokenPairWithFamily(ctx, userID, role, tokenFamily)
}

// generate Token
// generateTokenPairWithFamily generates tokens with a specific family
func (s *JWTService) generateTokenPairWithFamily(ctx context.Context, userID uuid.UUID, role entities.Role, tokenFamily string) (*TokenPair, error) {
	now := time.Now()
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	// Generate access token
	accessClaims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			Subject:   userID.String(),
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken, err := s.signToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := Claims{
		UserID:      userID,
		Role:        role,
		TokenType:   TokenTypeRefresh,
		TokenFamily: tokenFamily,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   userID.String(),
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken, err := s.signToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store current refresh JTI for this token family
	if err := s.tokenRepo.SetTokenFamily(ctx, userID, tokenFamily, refreshJTI, s.refreshTokenTTL); err != nil {
		return nil, err
	}

	// Track this session
	if err := s.tokenRepo.TrackUserSession(ctx, userID, tokenFamily, s.refreshTokenTTL); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessJTI:    accessJTI,
		RefreshJTI:   refreshJTI,
		TokenFamily:  tokenFamily,
		ExpiresAt:    now.Add(s.accessTokenTTL),
	}, nil
}

// signToken signs a token with the private key
func (s *JWTService) signToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// RefreshTokens validates refresh token and generates new token pair
func (s *JWTService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("invalid token type")
	}

	isBlacklisted, err := s.tokenRepo.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	// Check for token reuse
	currentJTI, err := s.tokenRepo.GetTokenFamily(ctx, claims.UserID, claims.TokenFamily)
	if err != nil {
		return nil, err
	}

	if currentJTI != "" && currentJTI != claims.ID {
		// Token reuse detected! Revoke all sessions
		if err := s.tokenRepo.RevokeAllUserSessions(ctx, claims.UserID); err != nil {
			return nil, err
		}
		return nil, errors.New("refresh token reuse detected, all sessions revoked")
	}

	// Blacklist the old refresh token
	if err := s.tokenRepo.BlacklistToken(ctx, claims.ID, time.Until(claims.ExpiresAt.Time)); err != nil {
		return nil, err
	}

	return s.generateTokenPairWithFamily(ctx, claims.UserID, claims.Role, claims.TokenFamily)
}

// ValidateToken validates a token and returns its claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing algorithm: %s", token.Method.Alg())
		}
		return s.publicKey, nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithLeeway(30*time.Second))

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token and checks blacklist
func (s *JWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, errors.New("invalid token type")
	}

	isBlacklisted, err := s.tokenRepo.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

// BlacklistToken adds a token JTI to the blacklist
func (s *JWTService) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	return s.tokenRepo.BlacklistToken(ctx, jti, ttl)
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *JWTService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return s.tokenRepo.RevokeAllUserSessions(ctx, userID)
}

// GetAccessTokenTTL returns the access token TTL
func (s *JWTService) GetAccessTokenTTL() time.Duration {
	return s.accessTokenTTL
}

// GetRefreshTokenTTL returns the refresh token TTL
func (s *JWTService) GetRefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}

// loadPrivateKey loads an RSA private key from a PEM file
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}

	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := keyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not RSA private key")
	}

	return rsaKey, nil
}

// loadPublicKey loads an RSA public key from a PEM file
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	keyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		rsaKey, ok := keyInterface.(*rsa.PublicKey)
		if ok {
			return rsaKey, nil
		}
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err == nil {
		rsaKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if ok {
			return rsaKey, nil
		}
	}

	return nil, errors.New("failed to parse public key")
}
