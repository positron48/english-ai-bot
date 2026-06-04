package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestLinglowEventRepository_RecordGrammarTestAttempt(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(98765)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	courseRepo := NewCourseRepository(conn, logger)
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	if _, err := courseRepo.BackfillUserCoursesForLearning(context.Background(), lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
		SELECT c.id, d.id, l.id, 'grammar_section:test-events', 'grammar', 'Grammar Events', 'grammar_section', 'events.section', 1, 'published'
		FROM courses c
		JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
		JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
		WHERE c.code = 'es_ru'
	`); err != nil {
		t.Fatalf("insert module: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'grammar_chapter', 'grammar_chapter', 'events.chapter', 'Events Chapter', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'grammar_section:test-events'
	`); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}

	repo := NewLinglowEventRepository(conn)
	input := GrammarTestEventInput{
		UserID:          user.ID,
		AttemptID:       123,
		ScopeType:       "chapter",
		ScopeID:         "events.chapter",
		Score:           80,
		Passed:          true,
		TotalQuestions:  5,
		AnswersJSON:     `[{"question_id":"q1","answer":"a"}]`,
		ResultsJSON:     `[{"question_id":"q1","correct":true}]`,
		ClientAttemptID: "client-123",
		AnsweredAt:      time.Now(),
	}

	exerciseID, err := repo.RecordGrammarTestAttempt(context.Background(), lc, input)
	if err != nil {
		t.Fatalf("RecordGrammarTestAttempt: %v", err)
	}
	if exerciseID == 0 {
		t.Fatalf("expected exercise id")
	}

	secondID, err := repo.RecordGrammarTestAttempt(context.Background(), lc, input)
	if err != nil {
		t.Fatalf("second RecordGrammarTestAttempt: %v", err)
	}
	if secondID != exerciseID {
		t.Fatalf("idempotent exercise id = %d, want %d", secondID, exerciseID)
	}

	var exercises, events, mappedItems int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE source_table = 'grammar_test_attempts' AND source_pk = '123'`).Scan(&exercises); err != nil {
		t.Fatalf("count exercise attempts: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM learning_events WHERE source_table = 'grammar_test_attempts' AND source_pk = '123'`).Scan(&events); err != nil {
		t.Fatalf("count learning events: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE id = ? AND learning_item_id IS NOT NULL`, exerciseID).Scan(&mappedItems); err != nil {
		t.Fatalf("count mapped items: %v", err)
	}
	if exercises != 1 || events != 1 || mappedItems != 1 {
		t.Fatalf("unexpected dual-write counts: exercises=%d events=%d mappedItems=%d", exercises, events, mappedItems)
	}
}

func TestLinglowEventRepository_RecordGrammarTrainingAttempt(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(98766)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	courseRepo := NewCourseRepository(conn, logger)
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	if _, err := courseRepo.BackfillUserCoursesForLearning(context.Background(), lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
		SELECT c.id, d.id, l.id, 'grammar_section:test-training-events', 'grammar', 'Grammar Training Events', 'grammar_section', 'training.events.section', 1, 'published'
		FROM courses c
		JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
		JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
		WHERE c.code = 'es_ru'
	`); err != nil {
		t.Fatalf("insert module: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'grammar_chapter', 'grammar_chapter', 'training.events.chapter', 'Training Events Chapter', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'grammar_section:test-training-events'
	`); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}

	repo := NewLinglowEventRepository(conn)
	input := GrammarTrainingEventInput{
		UserID:          user.ID,
		AttemptID:       456,
		ChapterID:       "training.events.chapter",
		TheoryBlockID:   "block-1",
		ConceptID:       "concept-1",
		QuestionID:      "question-1",
		IsCorrect:       true,
		AnswerJSON:      `"A"`,
		CorrectJSON:     `"A"`,
		ClientAttemptID: "training-client-456",
		AnsweredAt:      time.Now(),
	}

	exerciseID, err := repo.RecordGrammarTrainingAttempt(context.Background(), lc, input)
	if err != nil {
		t.Fatalf("RecordGrammarTrainingAttempt: %v", err)
	}
	secondID, err := repo.RecordGrammarTrainingAttempt(context.Background(), lc, input)
	if err != nil {
		t.Fatalf("second RecordGrammarTrainingAttempt: %v", err)
	}
	if secondID != exerciseID {
		t.Fatalf("idempotent exercise id = %d, want %d", secondID, exerciseID)
	}

	var exercises, events, mappedItems int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE source_table = 'grammar_attempts' AND source_pk = '456'`).Scan(&exercises); err != nil {
		t.Fatalf("count exercise attempts: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM learning_events WHERE source_table = 'grammar_attempts' AND source_pk = '456'`).Scan(&events); err != nil {
		t.Fatalf("count learning events: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE id = ? AND learning_item_id IS NOT NULL AND mode = 'grammar_training'`, exerciseID).Scan(&mappedItems); err != nil {
		t.Fatalf("count mapped items: %v", err)
	}
	if exercises != 1 || events != 1 || mappedItems != 1 {
		t.Fatalf("unexpected grammar training dual-write counts: exercises=%d events=%d mappedItems=%d", exercises, events, mappedItems)
	}
}

func TestLinglowEventRepository_RecordWordReviewEvent(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(98767)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	courseRepo := NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(context.Background(), lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}

	var wordCardID, trainingCardID, userCardID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition, display_en) VALUES ('casa', 'house', 'casa') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, 'casa', 0, 'дом', 'house') RETURNING id`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, 'es_ru', 'new', 2.5) RETURNING id`, user.ID, trainingCardID).Scan(&userCardID); err != nil {
		t.Fatalf("insert user card: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
		SELECT c.id, d.id, l.id, 'word_set:test-word-events', 'word_set', 'Word Events', 'word_set', 'test-word-set', 1, 'published'
		FROM courses c
		JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
		JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
		WHERE c.code = 'es_ru'
	`); err != nil {
		t.Fatalf("insert word module: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'word', 'word_card', ?, 'casa', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'word_set:test-word-events'
	`, fmt.Sprintf("%d", wordCardID)); err != nil {
		t.Fatalf("insert word learning item: %v", err)
	}

	repo := NewLinglowEventRepository(conn)
	input := WordReviewEventInput{
		UserID:          user.ID,
		ReviewEventID:   789,
		UserCardID:      userCardID,
		ClientAttemptID: "word-client-789",
		Direction:       "es_ru",
		IsCorrect:       true,
		Quality:         5,
		OptionsJSON:     `["дом","кот"]`,
		ChosenOption:    "дом",
		MetricsJSON:     `{"answer_time_ms":1200}`,
		SRSBeforeJSON:   `{"state":"new"}`,
		SRSAfterJSON:    `{"state":"learning"}`,
		AnsweredAt:      time.Now(),
	}

	exerciseID, err := repo.RecordWordReviewEvent(context.Background(), lc, input)
	if err != nil {
		t.Fatalf("RecordWordReviewEvent: %v", err)
	}
	secondID, err := repo.RecordWordReviewEvent(context.Background(), lc, input)
	if err != nil {
		t.Fatalf("second RecordWordReviewEvent: %v", err)
	}
	if secondID != exerciseID {
		t.Fatalf("idempotent exercise id = %d, want %d", secondID, exerciseID)
	}

	var exercises, events, mappedItems int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE source_table = 'review_events' AND source_pk = '789'`).Scan(&exercises); err != nil {
		t.Fatalf("count exercise attempts: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM learning_events WHERE source_table = 'review_events' AND source_pk = '789'`).Scan(&events); err != nil {
		t.Fatalf("count learning events: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE id = ? AND learning_item_id IS NOT NULL AND mode = 'word_training'`, exerciseID).Scan(&mappedItems); err != nil {
		t.Fatalf("count mapped items: %v", err)
	}
	if exercises != 1 || events != 1 || mappedItems != 1 {
		t.Fatalf("unexpected word review dual-write counts: exercises=%d events=%d mappedItems=%d", exercises, events, mappedItems)
	}
}
