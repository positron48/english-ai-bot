package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"tgbot-skeleton/internal/i18n"

	"go.uber.org/zap"
)

// RateLimitMiddleware wraps a handler with rate limiting
type RateLimitMiddleware struct {
	limiter *RateLimiter
	logger  *zap.Logger
	policy  RateLimitPolicy
	keyFn   func(r *http.Request) string
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(
	limiter *RateLimiter,
	logger *zap.Logger,
	policy RateLimitPolicy,
	keyFn func(r *http.Request) string,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		logger:  logger,
		policy:  policy,
		keyFn:   keyFn,
	}
}

// Wrap wraps an http.HandlerFunc with rate limiting
func (m *RateLimitMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for OPTIONS requests (CORS preflight)
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}

		// Generate rate limit key
		key := m.keyFn(r)
		if key == "" {
			// If key is empty, allow the request but log a warning
			m.logger.Warn("rate limit key is empty, allowing request",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			next(w, r)
			return
		}

		// Check rate limit
		allowed, retryAfter := m.limiter.Allow(key, m.policy)
		if !allowed {
			m.logger.Warn("rate limit exceeded",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("key_type", "ip"), // Simplified for logging
				zap.Int("retry_after", retryAfter),
			)

			lang := i18n.DetectLanguageFromRequest(r)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			w.WriteHeader(http.StatusTooManyRequests)

			response := map[string]interface{}{
				"error":               "rate_limited",
				"message":             i18n.T(lang, "errors.tooManyRequests"),
				"retry_after_seconds": retryAfter,
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				m.logger.Error("failed to encode rate limit response", zap.Error(err))
			}
			return
		}

		// Request allowed, proceed
		next(w, r)
	}
}

// KeyFuncIP returns a key function that uses client IP
func KeyFuncIP(r *http.Request) string {
	return "ip:" + clientIP(r)
}

// KeyFuncIPAndUserID returns a key function that uses IP + user ID from form
func KeyFuncIPAndUserID(r *http.Request) string {
	ip := clientIP(r)
	// Try to parse form if not already parsed
	if r.Form == nil {
		_ = r.ParseForm()
	}
	userID := r.FormValue("user_id")
	if userID == "" {
		// Try to get from JSON body if form parsing failed
		var body struct {
			UserID string `json:"user_id"`
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
			userID = body.UserID
		}
	}
	if userID != "" {
		return "ip:" + ip + ":user:" + userID
	}
	return "ip:" + ip
}

// KeyFuncIPAndUsername returns a key function that uses IP + username from form
func KeyFuncIPAndUsername(r *http.Request) string {
	ip := clientIP(r)
	// Try to parse form if not already parsed
	if r.Form == nil {
		_ = r.ParseForm()
	}
	username := r.FormValue("username")
	if username == "" {
		// Try to get from JSON body if form parsing failed
		var body struct {
			Username string `json:"username"`
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
			username = body.Username
		}
	}
	if username != "" {
		return "ip:" + ip + ":user:" + username
	}
	return "ip:" + ip
}

// KeyFuncIPAndUserIDFromContext returns a key function that uses IP + user ID from context
func KeyFuncIPAndUserIDFromContext(r *http.Request) string {
	ip := clientIP(r)
	userID := getUserIDFromContext(r.Context())
	if userID > 0 {
		return fmt.Sprintf("ip:%s:user:%d", ip, userID)
	}
	return "ip:" + ip
}
