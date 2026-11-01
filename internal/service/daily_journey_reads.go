package service

import (
	"context"
	"sync"
)

type journeyReadsKey struct{}
type journeyReads struct {
	mu     sync.Mutex
	values map[any]any
}

// Each operation gets its own cache. Keys include the database transaction, so
// AI completion never reuses authorization from its earlier reservation.
func withJourneyReads(ctx context.Context) context.Context {
	return context.WithValue(ctx, journeyReadsKey{}, &journeyReads{values: map[any]any{}})
}

func journeyRead[T any](ctx context.Context, key any, read func() (T, error)) (T, error) {
	cache, _ := ctx.Value(journeyReadsKey{}).(*journeyReads)
	if cache == nil {
		return read()
	}
	cache.mu.Lock()
	value, exists := cache.values[key]
	cache.mu.Unlock()
	if exists {
		return value.(T), nil
	}
	result, err := read()
	if err == nil {
		cache.mu.Lock()
		cache.values[key] = result
		cache.mu.Unlock()
	}
	return result, err
}
