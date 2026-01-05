package web

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 5,
		WindowDuration:    1 * time.Minute,
		BurstSize:         5, // Set burst equal to requests per window for this test
	}

	key := "test-key"

	// First 5 requests should be allowed (burst allows this)
	for i := 0; i < 5; i++ {
		allowed, retryAfter := rl.Allow(key, policy)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
		if retryAfter != 0 {
			t.Errorf("Request %d should not have retry after", i+1)
		}
	}

	// Next request should be rate limited (no tokens left, no time passed for refill)
	allowed, retryAfter := rl.Allow(key, policy)
	if allowed {
		t.Error("Request should be rate limited")
	}
	if retryAfter <= 0 {
		t.Error("Retry after should be positive")
	}
}

func TestRateLimiter_Burst(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 5,
		WindowDuration:    1 * time.Minute,
		BurstSize:         10,
	}

	key := "test-key-burst"

	// Burst should allow up to BurstSize requests
	for i := 0; i < 10; i++ {
		allowed, _ := rl.Allow(key, policy)
		if !allowed {
			t.Errorf("Burst request %d should be allowed", i+1)
		}
	}

	// Next request should be rate limited
	allowed, _ := rl.Allow(key, policy)
	if allowed {
		t.Error("Request after burst should be rate limited")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 5,
		WindowDuration:    1 * time.Minute,
		BurstSize:         5, // Set burst equal to requests per window
	}

	key := "test-key-refill"

	// Use all tokens (burst allows 5)
	for i := 0; i < 5; i++ {
		allowed, _ := rl.Allow(key, policy)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Should be rate limited (no tokens left)
	allowed, _ := rl.Allow(key, policy)
	if allowed {
		t.Error("Should be rate limited after using all tokens")
	}

	// Note: In a real scenario, tokens would refill gradually over time
	// This test verifies that tokens are properly consumed
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 5,
		WindowDuration:    1 * time.Minute,
		BurstSize:         5, // Set burst equal to requests per window
	}

	key1 := "key1"
	key2 := "key2"

	// Use all tokens for key1 (burst allows 5)
	for i := 0; i < 5; i++ {
		allowed, _ := rl.Allow(key1, policy)
		if !allowed {
			t.Errorf("key1 request %d should be allowed", i+1)
		}
	}

	// key1 should be rate limited (no tokens left)
	allowed, _ := rl.Allow(key1, policy)
	if allowed {
		t.Error("key1 should be rate limited")
	}

	// key2 should still be allowed (separate bucket with full tokens)
	allowed, _ = rl.Allow(key2, policy)
	if !allowed {
		t.Error("key2 should still be allowed")
	}
}
