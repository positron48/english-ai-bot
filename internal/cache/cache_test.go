package cache

import (
	"context"
	"testing"
	"time"
)

// TestRedisCache_NilClientIsNoOp verifies the core safety guarantee: when Redis isn't configured
// or unreachable, callers can still wire in a RedisCache unconditionally and every operation
// degrades to a safe no-op instead of panicking or returning a bogus hit.
func TestRedisCache_NilClientIsNoOp(t *testing.T) {
	c := NewRedisCache(nil, nil)
	ctx := context.Background()

	if val, ok := c.Get(ctx, "some-key"); ok || val != nil {
		t.Fatalf("Get on nil-client cache should always miss, got ok=%v val=%v", ok, val)
	}

	// Must not panic.
	c.Set(ctx, "some-key", []byte("value"), time.Minute)
	c.Delete(ctx, "some-key")

	if val, ok := c.Get(ctx, "some-key"); ok || val != nil {
		t.Fatalf("Get after Set/Delete on nil-client cache should still miss, got ok=%v val=%v", ok, val)
	}
}

// TestRedisCache_NilPointerIsNoOp verifies a nil *RedisCache (not just a nil inner client) is
// also a safe no-op, since callers may hold a nil pointer before construction in some paths.
func TestRedisCache_NilPointerIsNoOp(t *testing.T) {
	var c *RedisCache
	ctx := context.Background()

	if val, ok := c.Get(ctx, "some-key"); ok || val != nil {
		t.Fatalf("Get on nil *RedisCache should always miss, got ok=%v val=%v", ok, val)
	}
	c.Set(ctx, "some-key", []byte("value"), time.Minute) // must not panic
	c.Delete(ctx, "some-key")                            // must not panic
}
