package web

import (
	"sync"
	"time"
)

// RateLimitPolicy defines rate limiting rules
type RateLimitPolicy struct {
	// Requests per window
	RequestsPerWindow int
	// Window duration
	WindowDuration time.Duration
	// Burst size (max tokens in bucket)
	BurstSize int
}

// tokenBucket represents a token bucket for rate limiting
type tokenBucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// RateLimiter manages rate limiting with token bucket algorithm
type RateLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
	// Cleanup interval for stale buckets
	cleanupInterval time.Duration
	// TTL for buckets (if no activity, bucket is removed)
	bucketTTL time.Duration
	stopCleanup chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cleanupInterval, bucketTTL time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:         make(map[string]*tokenBucket),
		cleanupInterval: cleanupInterval,
		bucketTTL:       bucketTTL,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request is allowed under the given policy
// Returns (allowed, retryAfterSeconds)
func (rl *RateLimiter) Allow(key string, policy RateLimitPolicy) (bool, int) {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     policy.BurstSize, // Start with full burst capacity
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)

	// Refill tokens based on elapsed time
	if elapsed > 0 {
		// Calculate how many tokens should be available based on elapsed time
		tokensToAdd := int(elapsed.Seconds() * float64(policy.RequestsPerWindow) / policy.WindowDuration.Seconds())
		if tokensToAdd > 0 {
			bucket.tokens = min(bucket.tokens+tokensToAdd, policy.BurstSize)
			// Update lastRefill to now, but keep track of partial windows
			bucket.lastRefill = now
		}
	}

	// Check if we have tokens
	if bucket.tokens > 0 {
		bucket.tokens--
		return true, 0
	}

	// Calculate retry after - time until next token is available
	timeUntilNextToken := policy.WindowDuration / time.Duration(policy.RequestsPerWindow)
	retryAfter := int(timeUntilNextToken.Seconds())
	if retryAfter < 1 {
		retryAfter = 1
	}

	return false, retryAfter
}

// cleanup removes stale buckets periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, bucket := range rl.buckets {
				bucket.mu.Lock()
				lastActivity := bucket.lastRefill
				bucket.mu.Unlock()

				if now.Sub(lastActivity) > rl.bucketTTL {
					delete(rl.buckets, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCleanup:
			return
		}
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
