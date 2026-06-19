package repository

import (
	"context"
	"testing"

	"tgbot-skeleton/internal/testutil"
)

func TestLumiFactRepository_RotationAndFallback(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewLumiFactRepository(conn)
	ctx := context.Background()

	// Clear migration seed for a deterministic test.
	if _, err := conn.Exec(`DELETE FROM lumi_facts`); err != nil {
		t.Fatalf("clear: %v", err)
	}

	n, err := repo.BulkInsert(ctx, "es_ru", "general", "ru", []string{"fact one", "fact two"}, 0)
	if err != nil || n != 2 {
		t.Fatalf("BulkInsert: %d %v", n, err)
	}

	// First pick rotates to a fact; repeated picks the same day return the same fact.
	f1, err := repo.GetDailyFact(ctx, "es_ru", "general", "ru")
	if err != nil || f1 == nil {
		t.Fatalf("GetDailyFact: %v %v", f1, err)
	}
	f2, err := repo.GetDailyFact(ctx, "es_ru", "general", "ru")
	if err != nil || f2 == nil || f2.ID != f1.ID {
		t.Fatalf("expected stable fact of the day, got %v vs %v (%v)", f1, f2, err)
	}

	// Context fallback: grammar has no facts → general fact served.
	f3, err := repo.GetDailyFact(ctx, "es_ru", "grammar", "ru")
	if err != nil || f3 == nil {
		t.Fatalf("grammar fallback: %v %v", f3, err)
	}

	// Unknown course code → no facts, no error.
	f4, err := repo.GetDailyFact(ctx, "xx_yy", "general", "ru")
	if err != nil {
		t.Fatalf("unknown course: %v", err)
	}
	if f4 != nil {
		t.Fatalf("expected nil for unknown course without global facts, got %+v", f4)
	}

	// Archive both → nothing served.
	facts, _, err := repo.List(ctx, LumiFactFilter{CourseCode: "es_ru"})
	if err != nil || len(facts) != 2 {
		t.Fatalf("List: %d %v", len(facts), err)
	}
	for _, f := range facts {
		f.Status = "archived"
		if err := repo.Update(ctx, f); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	// Today's already-shown fact is archived too, so nothing comes back.
	f5, err := repo.GetDailyFact(ctx, "es_ru", "general", "ru")
	if err != nil {
		t.Fatalf("after archive: %v", err)
	}
	if f5 != nil {
		t.Fatalf("expected nil after archiving, got %+v", f5)
	}
}
