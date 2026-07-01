// Package cache provides a minimal, fail-safe caching abstraction backed by Redis.
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache is a byte-oriented cache abstraction. Implementations must never fail a caller: a miss
// (including any backend error) is reported as ok=false so callers always have a safe fallback
// to the source of truth (the database).
type Cache interface {
	Get(ctx context.Context, key string) (value []byte, ok bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
	Delete(ctx context.Context, key string)
	// Incr atomically increments the integer counter at key and returns the new value. Used for
	// generation-based invalidation: bumping a per-user generation makes every cache entry whose
	// key embeds the old generation unreachable at once. On a no-op/unavailable cache it returns
	// ok=false so callers fall back to no caching.
	Incr(ctx context.Context, key string) (value int64, ok bool)
}

// RedisCache implements Cache over go-redis. A nil *RedisCache, or one built with a nil client,
// behaves as a no-op cache: every Get misses, Set/Delete do nothing. Callers can wire it in
// unconditionally even when Redis isn't provisioned in the current environment.
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisCache wraps client, which may be nil (see RedisCache doc).
func NewRedisCache(client *redis.Client, logger *zap.Logger) *RedisCache {
	return &RedisCache{client: client, logger: logger}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, bool) {
	if c == nil || c.client == nil {
		return nil, false
	}
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil && c.logger != nil {
			c.logger.Warn("cache get failed, falling back to source", zap.String("key", key), zap.Error(err))
		}
		return nil, false
	}
	return val, true
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	if c == nil || c.client == nil {
		return
	}
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil && c.logger != nil {
		c.logger.Warn("cache set failed", zap.String("key", key), zap.Error(err))
	}
}

func (c *RedisCache) Delete(ctx context.Context, key string) {
	if c == nil || c.client == nil {
		return
	}
	if err := c.client.Del(ctx, key).Err(); err != nil && c.logger != nil {
		c.logger.Warn("cache delete failed", zap.String("key", key), zap.Error(err))
	}
}

func (c *RedisCache) Incr(ctx context.Context, key string) (int64, bool) {
	if c == nil || c.client == nil {
		return 0, false
	}
	v, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("cache incr failed", zap.String("key", key), zap.Error(err))
		}
		return 0, false
	}
	return v, true
}
