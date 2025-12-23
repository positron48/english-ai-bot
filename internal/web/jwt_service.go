package web

import (
	"errors"
	"fmt"
	"time"

	"tgbot-skeleton/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	secret       []byte
	ttlHours     int
	refreshHours int
	logger       *zap.Logger
}

// Claims represents JWT claims
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// RefreshClaims represents refresh token claims
type RefreshClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.Config, logger *zap.Logger) (*JWTService, error) {
	secret := cfg.WebApp.JWTSecret
	if secret == "" {
		// Fallback to session secret if JWT secret is not set (for backward compatibility)
		secret = cfg.WebApp.SessionSecret
	}
	if secret == "" {
		return nil, fmt.Errorf("JWT secret is required (set WEBAPP_JWT_SECRET or WEBAPP_SESSION_SECRET)")
	}

	ttlHours := cfg.WebApp.JWTTTLHours
	if ttlHours == 0 {
		// Fallback to session TTL if JWT TTL is not set (for backward compatibility)
		ttlHours = cfg.WebApp.SessionTTLHours
		if ttlHours == 0 {
			ttlHours = 24 // Default 24 hours for access token
		}
	}

	refreshHours := cfg.WebApp.RefreshTTLHours
	if refreshHours == 0 {
		refreshHours = 720 // Default 30 days for refresh token
	}

	return &JWTService{
		secret:       []byte(secret),
		ttlHours:     ttlHours,
		refreshHours: refreshHours,
		logger:       logger,
	}, nil
}

// GenerateToken generates a new access JWT token for a user
func (s *JWTService) GenerateToken(userID int64) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.ttlHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "english-bot",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		s.logger.Error("failed to generate JWT token", zap.Error(err), zap.Int64("user_id", userID))
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("JWT access token generated", zap.Int64("user_id", userID), zap.Time("expires_at", expiresAt))
	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh JWT token for a user
func (s *JWTService) GenerateRefreshToken(userID int64) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.refreshHours) * time.Hour)

	claims := &RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "english-bot",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err), zap.Int64("user_id", userID))
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	s.logger.Info("JWT refresh token generated", zap.Int64("user_id", userID), zap.Time("expires_at", expiresAt))
	return tokenString, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (s *JWTService) ValidateRefreshToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		s.logger.Warn("failed to parse refresh token", zap.Error(err))
		return 0, fmt.Errorf("invalid refresh token: %w", err)
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		s.logger.Warn("invalid refresh token claims")
		return 0, errors.New("invalid refresh token claims")
	}

	// Check expiration
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		s.logger.Warn("refresh token expired", zap.Int64("user_id", claims.UserID))
		return 0, errors.New("refresh token expired")
	}

	s.logger.Info("refresh token validated", zap.Int64("user_id", claims.UserID))
	return claims.UserID, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *JWTService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		s.logger.Warn("failed to parse JWT token", zap.Error(err))
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		s.logger.Warn("invalid JWT token claims")
		return 0, errors.New("invalid token claims")
	}

	// Check expiration
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		s.logger.Warn("JWT token expired", zap.Int64("user_id", claims.UserID))
		return 0, errors.New("token expired")
	}

	s.logger.Info("JWT token validated", zap.Int64("user_id", claims.UserID))
	return claims.UserID, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	// Check if it starts with "Bearer "
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}

