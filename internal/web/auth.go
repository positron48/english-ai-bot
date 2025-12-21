package web

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// userIDKey is defined in context.go

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	userRepo      *repository.UserRepository
	sessionRepo   *repository.WebSessionRepository
	logger        *zap.Logger
	config        *config.Config
	botToken      string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(
	userRepo *repository.UserRepository,
	sessionRepo *repository.WebSessionRepository,
	logger *zap.Logger,
	cfg *config.Config,
	botToken string,
) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
		config:      cfg,
		botToken:    botToken,
	}
}

// RequireAuth wraps a handler to require authentication
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	// Log all cookies for debugging
	cookies := r.Cookies()
	var sessionCookieFound bool
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookieFound = true
			m.logger.Info("session cookie in request", 
				zap.String("path", r.URL.Path),
				zap.String("token_preview", maskToken(c.Value)),
				zap.Bool("secure", c.Secure),
				zap.String("domain", c.Domain))
			break
		}
	}
	if !sessionCookieFound {
		m.logger.Info("no session cookie in request", zap.String("path", r.URL.Path))
	}
		
		userID, err := m.GetUserFromSession(r)
		if err != nil || userID == 0 {
			// Log why authentication failed
			if err != nil {
				m.logger.Info("auth failed", 
					zap.String("path", r.URL.Path), 
					zap.Error(err),
					zap.Bool("cookie_found", sessionCookieFound))
			} else {
				m.logger.Info("auth failed: userID is 0", 
					zap.String("path", r.URL.Path),
					zap.Bool("cookie_found", sessionCookieFound))
			}
			
			// Redirect to login or return 401
			if strings.HasPrefix(r.URL.Path, "/app/") {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Update last seen
		_ = m.sessionRepo.UpdateLastSeen(r)

		// Add user ID to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// GetUserFromSession extracts user ID from session cookie (public method)
func (m *AuthMiddleware) GetUserFromSession(r *http.Request) (int64, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		m.logger.Info("no session cookie found", zap.String("path", r.URL.Path), zap.Error(err))
		return 0, err
	}

	m.logger.Info("session cookie found, looking up in database", 
		zap.String("token_preview", maskToken(cookie.Value)),
		zap.String("path", r.URL.Path),
		zap.String("token_length", fmt.Sprintf("%d", len(cookie.Value))))

	session, err := m.sessionRepo.GetSessionByToken(cookie.Value)
	if err != nil {
		m.logger.Warn("failed to get session from DB", 
			zap.String("token_preview", maskToken(cookie.Value)),
			zap.Error(err))
		return 0, err
	}

	if session == nil {
		m.logger.Warn("session not found in DB", 
			zap.String("token_preview", maskToken(cookie.Value)),
			zap.String("token_full", cookie.Value))
		return 0, fmt.Errorf("session not found")
	}

	// Check expiration - use UTC for consistency
	now := time.Now().UTC()
	if now.After(session.ExpiresAt) {
		m.logger.Warn("session expired", 
			zap.String("token_preview", maskToken(cookie.Value)),
			zap.Time("expires_at", session.ExpiresAt),
			zap.Time("now", now))
		_ = m.sessionRepo.DeleteSession(session.Token)
		return 0, fmt.Errorf("session expired")
	}

	m.logger.Info("session validated successfully", 
		zap.Int64("user_id", session.UserID),
		zap.String("token_preview", maskToken(cookie.Value)))

	return session.UserID, nil
}

// maskToken masks the session token for logging (shows first 8 and last 8 characters)
func maskToken(token string) string {
	if len(token) == 0 {
		return ""
	}
	if len(token) <= 16 {
		return strings.Repeat("*", len(token))
	}
	return token[:8] + "..." + token[len(token)-8:]
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

// CreateSession creates a new web session
func (m *AuthMiddleware) CreateSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, err := generateSessionToken()
	if err != nil {
		return err
	}

	ttl := time.Duration(m.config.WebApp.SessionTTLHours) * time.Hour
	// Use UTC for consistency with database storage
	expiresAt := time.Now().UTC().Add(ttl)

	session := &repository.WebSession{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := m.sessionRepo.CreateSession(session); err != nil {
		return err
	}

	// Set cookie - always use Secure=true since site works over HTTPS
	cookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	}
	http.SetCookie(w, cookie)

	m.logger.Info("session created and cookie set",
		zap.Int64("user_id", userID),
		zap.String("token_preview", maskToken(token)),
		zap.String("cookie_path", cookie.Path))

	return nil
}

// generateSessionToken generates a random session token
func generateSessionToken() (string, error) {
	// Use crypto/rand for secure token generation
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

