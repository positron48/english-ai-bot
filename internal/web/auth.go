package web

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// userIDKey is defined in context.go

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	userRepo   *repository.UserRepository
	jwtService *JWTService
	logger     *zap.Logger
	config     *config.Config
	botToken   string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(
	userRepo *repository.UserRepository,
	jwtService *JWTService,
	logger *zap.Logger,
	cfg *config.Config,
	botToken string,
) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger,
		config:     cfg,
		botToken:   botToken,
	}
}

// RequireAuth wraps a handler to require authentication
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("authentication failed: missing Authorization header", 
				zap.String("path", r.URL.Path))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Unauthorized",
				"message": "Authentication required. Please provide a valid JWT token in Authorization header.",
			})
			return
		}

		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			m.logger.Warn("authentication failed: invalid Authorization header format", 
				zap.String("path", r.URL.Path),
				zap.Error(err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Unauthorized",
				"message": "Invalid Authorization header format. Expected: Bearer <token>",
			})
			return
		}

		userID, role, err := m.jwtService.ValidateToken(tokenString)
		if err != nil || userID == 0 {
			m.logger.Warn("authentication failed: invalid or expired token", 
				zap.String("path", r.URL.Path),
				zap.Error(err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Unauthorized",
				"message": "Invalid or expired token. Please authenticate again.",
			})
			return
		}

		m.logger.Info("JWT authentication successful", 
			zap.String("path", r.URL.Path),
			zap.Int64("user_id", userID),
			zap.String("role", role))
		
		// Add user ID and role to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		ctx = context.WithValue(ctx, userRoleKey, role)
		r = r.WithContext(ctx)
		next(w, r)
	}
}


// ValidateTelegramInitData validates Telegram WebApp initData
func (m *AuthMiddleware) ValidateTelegramInitData(initData string) (int64, error) {
	// Parse initData (format: key1=value1&key2=value2&hash=...)
	params := make(map[string]string)
	parts := strings.Split(initData, "&")
	var hash string

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]

		// URL-decode the value (Telegram sends URL-encoded values)
		if decodedValue, err := url.QueryUnescape(value); err == nil {
			value = decodedValue
		}
		// If decoding fails, use original value

		if key == "hash" {
			hash = value
		} else {
			params[key] = value
		}
	}

	if hash == "" {
		return 0, fmt.Errorf("hash not found in initData")
	}

	// Create secret_key = HMAC_SHA256("WebAppData", bot_token)
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(m.botToken))
	secretKeyBytes := secretKey.Sum(nil)

	// Create data_check_string: sort params by key and join as "key=value\n"
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	// Sort keys
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	var dataCheckParts []string
	for _, k := range keys {
		dataCheckParts = append(dataCheckParts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	dataCheckString := strings.Join(dataCheckParts, "\n")

	// Calculate HMAC_SHA256(secret_key, data_check_string)
	calculatedHash := hmac.New(sha256.New, secretKeyBytes)
	calculatedHash.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(calculatedHash.Sum(nil))

	if hash != expectedHash {
		return 0, fmt.Errorf("invalid hash")
	}

	// Extract user ID
	userStr, ok := params["user"]
	if !ok {
		return 0, fmt.Errorf("user not found in initData")
	}

	// Parse user JSON
	var userData struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userStr), &userData); err != nil {
		return 0, fmt.Errorf("failed to parse user data: %w", err)
	}

	return userData.ID, nil
}

// getUserRole determines the user's role based on their Telegram ID
func (m *AuthMiddleware) getUserRole(telegramID int64) string {
	if telegramID == int64(m.config.Admin.TelegramID) {
		return "admin"
	}
	return "user"
}

// GenerateJWTToken generates an access JWT token for a user
func (m *AuthMiddleware) GenerateJWTToken(userID int64, telegramID int64) (string, error) {
	role := m.getUserRole(telegramID)
	return m.jwtService.GenerateToken(userID, role)
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (m *AuthMiddleware) GenerateTokenPair(userID int64, telegramID int64) (accessToken, refreshToken string, err error) {
	role := m.getUserRole(telegramID)
	accessToken, err = m.jwtService.GenerateToken(userID, role)
	if err != nil {
		return "", "", err
	}
	
	refreshToken, err = m.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}
	
	return accessToken, refreshToken, nil
}
