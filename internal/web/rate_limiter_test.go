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
