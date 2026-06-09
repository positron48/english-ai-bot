package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

// seedCourseScopedCard inserts a word_card + training_card + user_card with a given course tag
// and state, returning nothing (helper for the course-filter test).
func seedCourseScopedCard(t *testing.T, conn *sql.DB, userID int64, word, state string, due *time.Time, courseCode *string) {
	t.Helper()
	var wcID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES (?, 'd') RETURNING id`, word).Scan(&wcID); err != nil {
		t.Fatalf("insert word_card %s: %v", word, err)
	}
	var tcID int64
	if err := conn.QueryRow(`
		INSERT INTO training_cards (word_card_id, word_en, word_ru, meaning_en, sense_index, course_code)
		VALUES (?, ?, 'ru', 'en', 0, ?) RETURNING id
	`, wcID, word, courseCode).Scan(&tcID); err != nil {
		t.Fatalf("insert training_card %s: %v", word, err)
	}
	if _, err := conn.Exec(`
		INSERT INTO user_cards (user_id, training_card_id, direction, state, next_due_at, course_code)
		VALUES (?, ?, 'en_to_ru', ?, ?, ?)
	`, userID, tcID, state, due, courseCode); err != nil {
		t.Fatalf("insert user_card %s: %v", word, err)
	}
}

func TestUserCardRepository_CourseScopedQueries(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(7711)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewUserCardRepository(conn, zap.NewNop())

	past := time.Now().Add(-time.Hour)
	es := "es_ru"
	en := "en_ru"

	// Due cards: one es_ru, one en_ru, one still-untagged (NULL).
	seedCourseScopedCard(t, conn, user.ID, "casa", "review", &past, &es)
	seedCourseScopedCard(t, conn, user.ID, "house", "review", &past, &en)
	seedCourseScopedCard(t, conn, user.ID, "untagged", "review", &past, nil)
	// New cards: one es_ru, one en_ru.
	seedCourseScopedCard(t, conn, user.ID, "perro", "new", nil, &es)
	seedCourseScopedCard(t, conn, user.ID, "dog", "new", nil, &en)

	now := time.Now()

	// No course filter -> everything (3 due incl. the new ones with NULL due, + ...).
	allDue, err := repo.GetDueCardsForCourse(user.ID, "", now, 50)
	if err != nil {
		t.Fatalf("GetDueCardsForCourse(all): %v", err)
	}
	// "new" cards have NULL next_due_at, so they also count as due here. 3 review + 2 new = 5.
	if len(allDue) != 5 {
		t.Fatalf("all due = %d want 5", len(allDue))
	}

	// es_ru filter -> es_ru rows + untagged (NULL), excludes en_ru.
	esDue, err := repo.GetDueCardsForCourse(user.ID, "es_ru", now, 50)
	if err != nil {
		t.Fatalf("GetDueCardsForCourse(es): %v", err)
	}
	// casa(es) + untagged(NULL) + perro(es,new) = 3; house(en) + dog(en) excluded.
	if len(esDue) != 3 {
		t.Fatalf("es due = %d want 3", len(esDue))
	}

	esCount, err := repo.GetDueCountForCourse(user.ID, "es_ru", now)
	if err != nil {
		t.Fatalf("GetDueCountForCourse(es): %v", err)
	}
	if esCount != 3 {
		t.Fatalf("es due count = %d want 3", esCount)
	}

	// New cards scoped to es_ru -> only perro (es), excludes dog (en).
	esNew, err := repo.GetNewCardsForCourse(user.ID, "es_ru", 50)
	if err != nil {
		t.Fatalf("GetNewCardsForCourse(es): %v", err)
	}
	if len(esNew) != 1 {
		t.Fatalf("es new = %d want 1", len(esNew))
	}

	// en_ru filter excludes es_ru-only words but still includes untagged.
	enDue, err := repo.GetDueCardsForCourse(user.ID, "en_ru", now, 50)
	if err != nil {
		t.Fatalf("GetDueCardsForCourse(en): %v", err)
	}
	// house(en) + untagged(NULL) + dog(en,new) = 3.
	if len(enDue) != 3 {
		t.Fatalf("en due = %d want 3", len(enDue))
	}
}
