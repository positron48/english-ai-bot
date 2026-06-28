package repository

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupCourseWordsFixture(t *testing.T) (*sql.DB, *CourseRepository, int64, int64) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	user, err := NewUserRepository(conn, logger).GetOrCreateUser(90001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	if _, err := repo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var rootCatID int64
	if err := conn.QueryRow(`
		INSERT INTO word_set_categories (parent_id, name, description, level_code, is_published, sort_order, course_code)
		VALUES (NULL, 'Level A0', 'A0 root', 'A0', 1, 1, 'es_ru')
		RETURNING id`).Scan(&rootCatID); err != nil {
		t.Fatalf("insert root category: %v", err)
	}
	var subCatID int64
	if err := conn.QueryRow(`
		INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order, course_code)
		VALUES (?, 'Basics', 'Basics', 1, 1, 'es_ru')
		RETURNING id`, rootCatID).Scan(&subCatID); err != nil {
		t.Fatalf("insert sub category: %v", err)
	}

	var wordCardKnown, wordCardStudied int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition, display_en, definition_ru) VALUES ('perro', 'dog', 'perro', 'собака') RETURNING id`).Scan(&wordCardKnown); err != nil {
		t.Fatalf("insert known word: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition, display_en, definition_ru) VALUES ('gato', 'cat', 'gato', 'кот') RETURNING id`).Scan(&wordCardStudied); err != nil {
		t.Fatalf("insert studied word: %v", err)
	}

	var wordSetID int64
	if err := conn.QueryRow(`
		INSERT INTO word_sets (category_id, title, description, is_published, sort_order, course_code)
		VALUES (?, 'Animals', 'Animals', 1, 1, 'es_ru')
		RETURNING id`, subCatID).Scan(&wordSetID); err != nil {
		t.Fatalf("insert word set: %v", err)
	}
	for _, wcID := range []int64{wordCardKnown, wordCardStudied} {
		if _, err := conn.Exec(`INSERT INTO word_set_items (word_set_id, word_card_id, sort_order) VALUES (?, ?, 1)`, wordSetID, wcID); err != nil {
			t.Fatalf("insert word_set_item: %v", err)
		}
	}

	if _, err := conn.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES (?, ?, 'known')`, user.ID, wordCardKnown); err != nil {
		t.Fatalf("insert known status: %v", err)
	}

	var trainingCardID int64
	if err := conn.QueryRow(`
		INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en)
		VALUES (?, 'gato', 0, 'кот', 'cat')
		RETURNING id`, wordCardStudied).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training card: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, 'es_ru', 'review', 2.5)`, user.ID, trainingCardID); err != nil {
		t.Fatalf("insert user card: %v", err)
	}

	return conn, repo, user.ID, wordCardStudied
}

func TestCourseRepository_GetWordLevelProgressForCourse(t *testing.T) {
	_, repo, userID, _ := setupCourseWordsFixture(t)

	empty, err := repo.GetWordLevelProgressForCourse(context.Background(), 0, "es_ru")
	if err != nil {
		t.Fatalf("empty user progress: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty map for user 0, got %+v", empty)
	}

	progress, err := repo.GetWordLevelProgressForCourse(context.Background(), userID, "es_ru")
	if err != nil {
		t.Fatalf("GetWordLevelProgressForCourse: %v", err)
	}
	a0, ok := progress["A0"]
	if !ok {
		t.Fatalf("expected A0 progress, got %+v", progress)
	}
	if a0.Total != 2 || a0.Mastered != 2 {
		t.Fatalf("A0 progress = %+v, want total=2 mastered=2", a0)
	}
}

func TestCourseRepository_GetWordListForUser_FiltersAndSort(t *testing.T) {
	conn, repo, userID, wordCardStudied := setupCourseWordsFixture(t)
	ctx := context.Background()

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
			SELECT course_id, district_id, location_id, 'word_set:cw-test', 'word_set', 'CW Test', 'word_set', 'cw-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', $1, 'gato', 'A0', 'published'
		FROM module
		RETURNING id
	`, strconv.FormatInt(wordCardStudied, 10)).Scan(&itemID); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}

	correct := true
	quality := 4
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID:          userID,
		DefaultCourse:   "es_ru",
		LearningItemID:  itemID,
		Mode:            "card",
		ClientAttemptID: "cw-attempt-1",
		IsCorrect:       &correct,
		Quality:         &quality,
		UpdateSRS:       true,
	}); err != nil {
		t.Fatalf("RecordExerciseAttempt: %v", err)
	}

	list, err := repo.GetWordListForUser(ctx, userID, "es_ru", "", WordListOptions{Limit: 50, Sort: "word_desc"})
	if err != nil {
		t.Fatalf("GetWordListForUser: %v", err)
	}
	if list.Total != 1 || len(list.Words) != 1 {
		t.Fatalf("list = total %d len %d", list.Total, len(list.Words))
	}
	if list.Words[0].Lemma != "gato" {
		t.Fatalf("word = %+v", list.Words[0])
	}

	reviewList, err := repo.GetWordListForUser(ctx, userID, "es_ru", "", WordListOptions{Limit: 50, Status: "learning"})
	if err != nil {
		t.Fatalf("GetWordListForUser learning: %v", err)
	}
	if reviewList.Total != 1 {
		t.Fatalf("learning filter total=%d", reviewList.Total)
	}

	newList, err := repo.GetWordListForUser(ctx, userID, "es_ru", "", WordListOptions{Limit: 50, Status: "new"})
	if err != nil {
		t.Fatalf("GetWordListForUser new: %v", err)
	}
	if newList.Total != 0 {
		t.Fatalf("new filter total=%d, want 0", newList.Total)
	}

	if _, err := repo.GetWordListForUser(ctx, 0, "es_ru", "", WordListOptions{}); err == nil {
		t.Fatal("expected error for user id 0")
	}
}

func TestCourseRepository_GetWordListForUser_MasterySort(t *testing.T) {
	conn, repo, userID, wordCardStudied := setupCourseWordsFixture(t)
	ctx := context.Background()

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
			SELECT course_id, district_id, location_id, 'word_set:cw-sort', 'word_set', 'CW Sort', 'word_set', 'cw-sort', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', $1, 'gato', 'A0', 'published'
		FROM module
		RETURNING id
	`, strconv.FormatInt(wordCardStudied, 10)).Scan(&itemID); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}

	correct := true
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID:          userID,
		DefaultCourse:   "es_ru",
		LearningItemID:  itemID,
		Mode:            "card",
		ClientAttemptID: "cw-sort-1",
		IsCorrect:       &correct,
		UpdateSRS:       true,
	}); err != nil {
		t.Fatalf("RecordExerciseAttempt: %v", err)
	}

	for _, sort := range []string{"mastery_asc", "mastery_desc", "added_at"} {
		list, err := repo.GetWordListForUser(ctx, userID, "es_ru", "", WordListOptions{Limit: 50, Sort: sort})
		if err != nil {
			t.Fatalf("GetWordListForUser sort=%s: %v", sort, err)
		}
		if list.Total != 1 {
			t.Fatalf("sort=%s total=%d", sort, list.Total)
		}
	}
}
