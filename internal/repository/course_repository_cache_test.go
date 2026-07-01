package repository

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

// fakeCourseMapCache is an in-memory cache.Cache used to observe how CourseRepository interacts
// with its cache dependency (hit/miss, invalidation) without needing a real Redis instance.
type fakeCourseMapCache struct {
	store       map[string][]byte
	counters    map[string]int64
	setCalls    []string
	deleteCalls []string
}

func newFakeCourseMapCache() *fakeCourseMapCache {
	return &fakeCourseMapCache{store: map[string][]byte{}, counters: map[string]int64{}}
}

func (f *fakeCourseMapCache) Get(_ context.Context, key string) ([]byte, bool) {
	v, ok := f.store[key]
	return v, ok
}

func (f *fakeCourseMapCache) Set(_ context.Context, key string, value []byte, _ time.Duration) {
	f.store[key] = value
	f.setCalls = append(f.setCalls, key)
}

func (f *fakeCourseMapCache) Delete(_ context.Context, key string) {
	delete(f.store, key)
	f.deleteCalls = append(f.deleteCalls, key)
}

func (f *fakeCourseMapCache) Incr(_ context.Context, key string) (int64, bool) {
	f.counters[key]++
	return f.counters[key], true
}

// TestCourseRepository_GetCourseMap_CachesStructureNotUserCourse verifies that the expensive
// structural fetch (districts/locations/modules/items) is served from cache on a second call —
// while each user still gets their own live UserCourse, never a cached one belonging to someone
// else. This is the correctness nuance the caching design hinges on.
func TestCourseRepository_GetCourseMap_CachesStructureNotUserCourse(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)

	user1, err := userRepo.GetOrCreateUser(90001)
	if err != nil {
		t.Fatalf("create user1: %v", err)
	}
	user2, err := userRepo.GetOrCreateUser(90002)
	if err != nil {
		t.Fatalf("create user2: %v", err)
	}

	repo := NewCourseRepository(conn, logger)
	cache := newFakeCourseMapCache()
	repo.SetCourseMapCache(cache)

	if _, err := repo.BackfillUserCourses(context.Background(), "es_ru"); err != nil {
		t.Fatalf("BackfillUserCourses: %v", err)
	}

	map1, err := repo.GetCourseMap(context.Background(), "es_ru", user1.ID)
	if err != nil {
		t.Fatalf("GetCourseMap user1: %v", err)
	}
	if len(cache.setCalls) != 1 {
		t.Fatalf("expected exactly 1 cache Set after first (cold) call, got %d: %v", len(cache.setCalls), cache.setCalls)
	}
	if map1.UserCourse == nil {
		t.Fatalf("expected user1 UserCourse to be populated")
	}

	map2, err := repo.GetCourseMap(context.Background(), "es_ru", user2.ID)
	if err != nil {
		t.Fatalf("GetCourseMap user2: %v", err)
	}
	if len(cache.setCalls) != 1 {
		t.Fatalf("expected structural fetch to be served from cache (no additional Set) on second call, got %d: %v", len(cache.setCalls), cache.setCalls)
	}
	if map2.UserCourse == nil {
		t.Fatalf("expected user2 UserCourse to be populated")
	}
	if map2.UserCourse.ID == map1.UserCourse.ID {
		t.Fatalf("user1 and user2 must not share the same UserCourse row: got %d for both", map1.UserCourse.ID)
	}
	if map1.Totals.Districts != map2.Totals.Districts || map1.Totals.Locations != map2.Totals.Locations {
		t.Fatalf("structural totals should be identical across users: user1=%+v user2=%+v", map1.Totals, map2.Totals)
	}
}

// TestCourseRepository_MapLegacyContent_InvalidatesCache verifies that writing course structure
// invalidates the cached entry, so the next GetCourseMap call rebuilds from the database instead
// of serving stale content indefinitely.
func TestCourseRepository_MapLegacyContent_InvalidatesCache(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()

	repo := NewCourseRepository(conn, logger)
	cache := newFakeCourseMapCache()
	repo.SetCourseMapCache(cache)

	if _, err := repo.GetCourseMap(context.Background(), "es_ru", 0); err != nil {
		t.Fatalf("GetCourseMap (warm cache): %v", err)
	}
	if _, ok := cache.store[courseMapCacheKey("es_ru")]; !ok {
		t.Fatalf("expected cache to hold an entry for es_ru after GetCourseMap")
	}

	if _, err := repo.MapLegacyContent(context.Background(), "es_ru", "es"); err != nil {
		t.Fatalf("MapLegacyContent: %v", err)
	}

	if _, ok := cache.store[courseMapCacheKey("es_ru")]; ok {
		t.Fatalf("expected MapLegacyContent to invalidate the es_ru cache entry")
	}
	if len(cache.deleteCalls) != 1 || cache.deleteCalls[0] != courseMapCacheKey("es_ru") {
		t.Fatalf("expected exactly one Delete call for %q, got %v", courseMapCacheKey("es_ru"), cache.deleteCalls)
	}
}

// TestCourseRepository_GetCourseMap_NilCacheIsSafe verifies a CourseRepository without a cache
// wired in (the default) still works — mirrors bot bootstrap and cmd/import_learning_content,
// which never call SetCourseMapCache.
func TestCourseRepository_GetCourseMap_NilCacheIsSafe(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	repo := NewCourseRepository(conn, logger)

	if _, err := repo.GetCourseMap(context.Background(), "es_ru", 0); err != nil {
		t.Fatalf("GetCourseMap with nil cache should work: %v", err)
	}
}
