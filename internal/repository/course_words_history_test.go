package repository

import (
	"context"
	"strconv"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestCourseRepository_GetWordListAndHistory(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4901)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	if _, err := repo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var wordCardID int64
	if err := conn.QueryRow(`
		INSERT INTO word_cards (word, definition, display_en, definition_ru)
		VALUES ('casa', 'house', 'house', 'дом')
		RETURNING id
	`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}

	var itemID int64
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'word_set:wl-test', 'word_set', 'WL Test', 'word_set', 'wl-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', $1, 'casa', 'A0', 'published'
		FROM module
		RETURNING id
	`, strconv.FormatInt(wordCardID, 10)).Scan(&itemID); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}

	correct := true
	quality := 4
	if _, err := repo.RecordExerciseAttempt(context.Background(), ExerciseAttemptInput{
		UserID:          user.ID,
		DefaultCourse:   "es_ru",
		LearningItemID:  itemID,
		Mode:            "card",
		ClientAttemptID: "wl-attempt-1",
		IsCorrect:       &correct,
		Quality:         &quality,
		UpdateSRS:       true,
	}); err != nil {
		t.Fatalf("RecordExerciseAttempt: %v", err)
	}

	// Word list should surface the studied word with binding-correct queries.
	list, err := repo.GetWordListForUser(context.Background(), user.ID, "es_ru", "", WordListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("GetWordListForUser: %v", err)
	}
	if list.Total != 1 || len(list.Words) != 1 {
		t.Fatalf("word list total=%d len=%d words=%+v", list.Total, len(list.Words), list.Words)
	}
	w := list.Words[0]
	if w.LearningItemID != itemID {
		t.Fatalf("word learning_item_id=%d want %d", w.LearningItemID, itemID)
	}
	if w.Lemma != "casa" {
		t.Fatalf("word lemma=%q want casa", w.Lemma)
	}
	if w.Translation != "дом" {
		t.Fatalf("word translation=%q want дом", w.Translation)
	}

	// Search filter should match the translation.
	filtered, err := repo.GetWordListForUser(context.Background(), user.ID, "es_ru", "", WordListOptions{Limit: 50, Search: "дом"})
	if err != nil {
		t.Fatalf("GetWordListForUser search: %v", err)
	}
	if filtered.Total != 1 || len(filtered.Words) != 1 {
		t.Fatalf("filtered word list total=%d len=%d", filtered.Total, len(filtered.Words))
	}

	// A non-matching search returns nothing.
	none, err := repo.GetWordListForUser(context.Background(), user.ID, "es_ru", "", WordListOptions{Limit: 50, Search: "zzzznomatch"})
	if err != nil {
		t.Fatalf("GetWordListForUser empty search: %v", err)
	}
	if none.Total != 0 || len(none.Words) != 0 {
		t.Fatalf("empty search total=%d len=%d", none.Total, len(none.Words))
	}

	// History should reflect the attempt.
	hist, err := repo.GetHistoryForUser(context.Background(), user.ID, "es_ru", "", 7)
	if err != nil {
		t.Fatalf("GetHistoryForUser: %v", err)
	}
	if hist.TotalAttempts != 1 || hist.CorrectAttempts != 1 {
		t.Fatalf("history attempts=%d correct=%d", hist.TotalAttempts, hist.CorrectAttempts)
	}
	if hist.AccuracyPercent < 99.9 {
		t.Fatalf("history accuracy=%v want ~100", hist.AccuracyPercent)
	}
	if len(hist.WeeklyStats) == 0 {
		t.Fatalf("expected weekly stats")
	}
	if len(hist.WordsAddedStats) == 0 {
		t.Fatalf("expected words-added stats (srs item created)")
	}
	var foundCardMode bool
	for _, m := range hist.ByMode {
		if m.Mode == "card" && m.AttemptCount == 1 && m.CorrectCount == 1 {
			foundCardMode = true
		}
	}
	if !foundCardMode {
		t.Fatalf("expected card mode in by_mode, got %+v", hist.ByMode)
	}
}
