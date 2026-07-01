package web

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// Per-user count queries (vocab summary, verb-forms pool) are expensive but change only when the
// user's own vocabulary/training data changes. We cache them per user+course behind a per-user
// generation counter: any data change bumps the generation (one INCR), which makes every cached
// entry whose key embeds the old generation unreachable at once — "dynamically invalidated"
// without hunting down every read key. A TTL backstop self-heals any mutation path we miss.
const (
	userCacheGenPrefix = "linglow:cachegen:"
	userCountCacheTTL  = 10 * time.Minute
)

// userCacheGen returns the user's current cache generation (0 when caching is disabled or unset).
func (r *Router) userCacheGen(ctx context.Context, userID int64) int64 {
	if r == nil || r.cache == nil || userID == 0 {
		return 0
	}
	b, ok := r.cache.Get(ctx, userCacheGenPrefix+strconv.FormatInt(userID, 10))
	if !ok {
		return 0
	}
	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// BumpUserCache invalidates all cached per-user count queries for userID by advancing its
// generation. Safe no-op when caching is disabled or userID is 0. Call after any change to the
// user's vocabulary/training/verb data.
func (r *Router) BumpUserCache(userID int64) {
	if r == nil || r.cache == nil || userID == 0 {
		return
	}
	r.cache.Incr(context.Background(), userCacheGenPrefix+strconv.FormatInt(userID, 10))
}

// cachedUserBytes returns the cached JSON bytes for a per-user, per-course query, computing and
// caching them on a miss. compute is the source-of-truth producer. When caching is disabled it
// simply calls compute. name namespaces the query (e.g. "vocabsummary", "verbupcoming").
func (r *Router) cachedUserBytes(ctx context.Context, userID int64, name, courseCode string, compute func() ([]byte, error)) ([]byte, error) {
	if r == nil || r.cache == nil {
		return compute()
	}
	gen := r.userCacheGen(ctx, userID)
	key := fmt.Sprintf("linglow:%s:%d:%d:%s", name, userID, gen, courseCode)
	if b, ok := r.cache.Get(ctx, key); ok {
		return b, nil
	}
	b, err := compute()
	if err != nil {
		return b, err
	}
	r.cache.Set(ctx, key, b, userCountCacheTTL)
	return b, nil
}
