package web

import (
	"sync"
	"testing"
	"time"
)

func TestMin(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a < b", 1, 2, 1},
		{"a > b", 2, 1, 1},
		{"a == b", 1, 1, 1},
		{"negative a < b", -2, -1, -2},
		{"zero", 0, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := min(tt.a, tt.b); got != tt.want {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 100*time.Millisecond)

	// Add some buckets
	rl.mu.Lock()
	oldTime := time.Now().Add(-200 * time.Millisecond) // Old, should be cleaned
	rl.buckets["key1"] = &tokenBucket{
		tokens:     10,
		lastRefill: oldTime,
		mu:         sync.Mutex{},
	}
	recentTime := time.Now() // Recent, should not be cleaned
	rl.buckets["key2"] = &tokenBucket{
		tokens:     10,
		lastRefill: recentTime,
		mu:         sync.Mutex{},
	}
	rl.mu.Unlock()

	// Wait for cleanup to run (interval 50ms; wait long enough on slow CI)
	time.Sleep(300 * time.Millisecond)
	rl.Stop()

	rl.mu.Lock()
	defer rl.mu.Unlock()
	if _, exists := rl.buckets["key1"]; exists {
		t.Error("Expected key1 bucket to be cleaned up")
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	tests := []struct {
		name           string
		key            string
		policy         RateLimitPolicy
		wantAllowed    bool
		wantRetryAfter int
	}{
		{
			name: "new key allowed with burst",
			key:  "user1",
			policy: RateLimitPolicy{
				RequestsPerWindow: 10,
				WindowDuration:    1 * time.Minute,
				BurstSize:         5,
			},
			wantAllowed:    true,
			wantRetryAfter: 0,
		},
		{
			name: "different key allowed",
			key:  "user2",
			policy: RateLimitPolicy{
				RequestsPerWindow: 1,
				WindowDuration:    1 * time.Second,
				BurstSize:         1,
			},
			wantAllowed:    true,
			wantRetryAfter: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, retryAfter := rl.Allow(tt.key, tt.policy)
			if allowed != tt.wantAllowed || retryAfter != tt.wantRetryAfter {
				t.Errorf("Allow() = (%v, %d), want (%v, %d)", allowed, retryAfter, tt.wantAllowed, tt.wantRetryAfter)
			}
		})
	}
}

func TestRateLimiter_Allow_Exhausted(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 10,
		WindowDuration:    1 * time.Minute,
		BurstSize:         1,
	}
	key := "exhausted"

	allowed, retryAfter := rl.Allow(key, policy)
	if !allowed || retryAfter != 0 {
		t.Errorf("first Allow() = (%v, %d), want (true, 0)", allowed, retryAfter)
	}

	allowed, retryAfter = rl.Allow(key, policy)
	if allowed {
		t.Error("second Allow() should be denied")
	}
	if retryAfter < 1 {
		t.Errorf("retryAfter = %d, want >= 1", retryAfter)
	}
}

func TestRateLimiter_Allow_RetryAfterClampedToOne(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	// Window 100ms, 10 req -> 10ms per token; int(0.01) = 0, code clamps to 1
	policy := RateLimitPolicy{
		RequestsPerWindow: 10,
		WindowDuration:    100 * time.Millisecond,
		BurstSize:         1,
	}
	key := "clamp"

	_, _ = rl.Allow(key, policy) // consume token
	_, retryAfter := rl.Allow(key, policy)
	if retryAfter != 1 {
		t.Errorf("retryAfter = %d, want 1 (clamped)", retryAfter)
	}
}

func TestRateLimiter_Allow_Refill(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	// 10 tokens per second, burst 1 -> refill 1 token per 100ms
	policy := RateLimitPolicy{
		RequestsPerWindow: 10,
		WindowDuration:    1 * time.Second,
		BurstSize:         1,
	}
	key := "refill"

	allowed, _ := rl.Allow(key, policy)
	if !allowed {
		t.Fatal("first Allow should succeed")
	}
	allowed, _ = rl.Allow(key, policy)
	if allowed {
		t.Fatal("second Allow should fail before refill")
	}

	time.Sleep(150 * time.Millisecond) // one token refill
	allowed, _ = rl.Allow(key, policy)
	if !allowed {
		t.Error("Allow after refill should succeed")
	}
}

func TestRateLimiter_Allow_ElapsedNoTokensAdded(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	// 10 req/s, burst 1. Sleep 50ms -> elapsed > 0 but tokensToAdd = int(0.5) = 0
	policy := RateLimitPolicy{
		RequestsPerWindow: 10,
		WindowDuration:    1 * time.Second,
		BurstSize:         1,
	}
	key := "no_add"

	_, _ = rl.Allow(key, policy)
	time.Sleep(50 * time.Millisecond)
	allowed, _ := rl.Allow(key, policy)
	if allowed {
		t.Error("Allow after short sleep should still be denied (no token refill)")
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 200*time.Millisecond)
	rl.Stop()
	// Stop must not panic; cleanup goroutine should exit
}
