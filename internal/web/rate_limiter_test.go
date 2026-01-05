package web

import (
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	_ = rl // Verify rate limiter is created
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 5,
		WindowDuration:    1 * time.Minute,
		BurstSize:         5,
	}

	// First requests should be allowed
	for i := 0; i < 5; i++ {
		allowed, _ := rl.Allow("test-key", policy)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should be rate limited
	allowed, _ := rl.Allow("test-key", policy)
	if allowed {
		t.Error("Request should be rate limited after limit")
	}
}

func TestRateLimiter_Allow_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 2,
		WindowDuration:    1 * time.Minute,
		BurstSize:         2,
	}

	// Different keys should have separate limits
	allowed1, _ := rl.Allow("key1", policy)
	allowed2, _ := rl.Allow("key2", policy)

	if !allowed1 || !allowed2 {
		t.Error("Different keys should have separate rate limits")
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)

	// Stop should not panic
	rl.Stop()
}
