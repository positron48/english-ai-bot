package repository

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupCoursePlacementAccess(t *testing.T) (*sql.DB, *GrammarAttemptRepository, int64, int64, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	user, err := NewUserRepository(db, zap.NewNop()).GetOrCreateUser(970001)
	if err != nil {
		t.Fatal(err)
	}
	courses := NewCourseRepository(db, zap.NewNop())
	en, err := courses.EnsureUserCourse(context.Background(), user.ID, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	es, err := courses.EnsureUserCourse(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatal(err)
	}
	return db, NewGrammarAttemptRepository(db, zap.NewNop()), user.ID, en.ID, es.ID
}

func saveDiagnosticAccess(t *testing.T, db *sql.DB, userCourseID int64, score int, sections ...string) {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err := SaveDiagnosticPlacementAccessTx(context.Background(), tx, userCourseID, score, 30, sections); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestCoursePlacementAccess_IsolationAndLegacyPreservation(t *testing.T) {
	db, repo, userID, enID, _ := setupCoursePlacementAccess(t)
	if err := repo.UpsertPlacementByAdmin(userID, 85, 25, []string{"en.grammar.b1", "es.grammar.a2"}); err != nil {
		t.Fatal(err)
	}
	en := repo.ForCourse("en_ru")
	es := repo.ForCourse("es_ru")
	old, err := en.GetPlacementTestResult(userID)
	if err != nil || old == nil || !reflect.DeepEqual(old.OpenedSections, []string{"en.grammar.b1"}) || old.Source != "legacy" || !old.AdminOverride {
		t.Fatalf("attributed legacy = %#v, %v", old, err)
	}
	saveDiagnosticAccess(t, db, enID, 30, "en.grammar.a1")
	result, err := en.GetPlacementTestResult(userID)
	if err != nil || result == nil || !reflect.DeepEqual(result.OpenedSections, []string{"en.grammar.a1", "en.grammar.b1"}) || result.Source != "diagnostic" || !result.AdminOverride {
		t.Fatalf("preserved access = %#v, %v", result, err)
	}
	saveDiagnosticAccess(t, db, enID, 0)
	weaker, err := en.GetPlacementTestResult(userID)
	if err != nil || !reflect.DeepEqual(weaker.OpenedSections, result.OpenedSections) || weaker.Score != 0 {
		t.Fatalf("weaker attempt narrowed access: %#v, %v", weaker, err)
	}
	spanish, err := es.GetPlacementTestResult(userID)
	if err != nil || spanish == nil || !reflect.DeepEqual(spanish.OpenedSections, []string{"es.grammar.a2"}) || spanish.Source != "legacy" {
		t.Fatalf("Spanish legacy changed: %#v, %v", spanish, err)
	}
	if repo.CourseCode() != "" {
		t.Fatal("ForCourse mutated the shared repository")
	}
}

func TestCoursePlacementAccess_AdminResetDoesNotResurrectLegacy(t *testing.T) {
	db, repo, userID, enID, esID := setupCoursePlacementAccess(t)
	if err := repo.SavePlacementTestResult(userID, 80, 25, []string{"en.grammar.b2"}); err != nil {
		t.Fatal(err)
	}
	saveDiagnosticAccess(t, db, esID, 70, "es.grammar.a2")
	en := repo.ForCourse("en_ru")
	if err := en.DeletePlacementTestResult(userID); err != nil {
		t.Fatal(err)
	}
	if got, err := en.GetPlacementTestResult(userID); err != nil || got != nil {
		t.Fatalf("cleared legacy reappeared: %#v, %v", got, err)
	}
	saveDiagnosticAccess(t, db, enID, 10, "en.grammar.a1")
	got, err := en.GetPlacementTestResult(userID)
	if err != nil || got == nil || !reflect.DeepEqual(got.OpenedSections, []string{"en.grammar.a1"}) || got.AdminOverride {
		t.Fatalf("diagnostic after reset: %#v, %v", got, err)
	}
	spanish, err := repo.ForCourse("es_ru").GetPlacementTestResult(userID)
	if err != nil || spanish == nil || !reflect.DeepEqual(spanish.OpenedSections, []string{"es.grammar.a2"}) {
		t.Fatalf("reset affected Spanish: %#v, %v", spanish, err)
	}
	if err := en.UpsertPlacementByAdmin(userID, 0, 0, []string{}); err != nil {
		t.Fatal(err)
	}
	below, err := en.GetPlacementTestResult(userID)
	if err != nil || below == nil || !below.AdminOverride || below.Source != "admin" || len(below.OpenedSections) != 0 {
		t.Fatalf("explicit below A1 must remain visible: %#v, %v", below, err)
	}
}

func TestCoursePlacementAccess_RollbackAndUnattributableLegacy(t *testing.T) {
	db, repo, userID, enID, _ := setupCoursePlacementAccess(t)
	if err := repo.UpsertPlacementByAdmin(userID, 0, 0, nil); err != nil {
		t.Fatal(err)
	}
	for _, course := range []string{"en_ru", "es_ru"} {
		if result, err := repo.ForCourse(course).GetPlacementTestResult(userID); err != nil || result != nil {
			t.Fatalf("unattributable empty legacy for %s: %#v, %v", course, result, err)
		}
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveDiagnosticPlacementAccessTx(context.Background(), tx, enID, 70, 30, []string{"en.grammar.a2"}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if result, err := repo.ForCourse("en_ru").GetPlacementTestResult(userID); err != nil || result != nil {
		t.Fatalf("rolled-back access persisted: %#v, %v", result, err)
	}
}

func TestCoursePlacementAccess_MalformedLegacyDoesNotBlockNewResult(t *testing.T) {
	for _, raw := range []string{"not-json", `["en.grammar.old",42]`} {
		t.Run(raw, func(t *testing.T) {
			db, repo, userID, enID, _ := setupCoursePlacementAccess(t)
			if err := repo.UpsertPlacementByAdmin(userID, 80, 25, []string{"en.grammar.old"}); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`UPDATE grammar_placement_test SET opened_sections_json=? WHERE user_id=?`, raw, userID); err != nil {
				t.Fatal(err)
			}
			result, err := repo.ForCourse("en_ru").GetPlacementTestResult(userID)
			if err != nil || result != nil {
				t.Fatalf("malformed legacy should not grant access: %+v %v", result, err)
			}
			saveDiagnosticAccess(t, db, enID, 75, "en.grammar.new")
			result, err = repo.ForCourse("en_ru").GetPlacementTestResult(userID)
			if err != nil || result == nil || !reflect.DeepEqual(result.OpenedSections, []string{"en.grammar.new"}) {
				t.Fatalf("new result: %+v %v", result, err)
			}
			var archived string
			if err = db.QueryRow(`SELECT opened_sections_json FROM grammar_placement_test WHERE user_id=?`, userID).Scan(&archived); err != nil || archived != raw {
				t.Fatalf("archive changed: %q %v", archived, err)
			}
		})
	}
}
