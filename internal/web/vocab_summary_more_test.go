package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestHandleVocabSummary_GuardsAndCourseFilter(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(424243)

	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "en"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/vocab/summary", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleVocabSummary(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST expected 405, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/vocab/summary", nil)
	w = httptest.NewRecorder()
	router.handleVocabSummary(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no auth expected 401, got %d", w.Code)
	}

	insertWord := func(word, course string, state string) int64 {
		t.Helper()
		var wordCardID int64
		if err := db.QueryRow(`
			INSERT INTO word_cards (word, definition, course_code)
			VALUES ($1, $2, $3) RETURNING id
		`, word, word, course).Scan(&wordCardID); err != nil {
			t.Fatalf("insert word_cards: %v", err)
		}
		var trainingCardID int64
		if err := db.QueryRow(`
			INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, course_code)
			VALUES ($1, $2, 0, $3, $4, $5) RETURNING id
		`, wordCardID, word, word+"_ru", word, course).Scan(&trainingCardID); err != nil {
			t.Fatalf("insert training_cards: %v", err)
		}
		if _, err := db.Exec(`
			INSERT INTO user_cards (user_id, training_card_id, direction, state)
			VALUES ($1, $2, 'ru_en', $3)
		`, user.ID, trainingCardID, state); err != nil {
			t.Fatalf("insert user_cards: %v", err)
		}
		return wordCardID
	}

	insertWord("englishonly", "en_ru", "review")
	insertWord("spanishonly", "es_ru", "new")

	// Multi-course DB: tag two course codes on training_cards.
	courseRepo := repository.NewCourseRepository(db, logger)
	if _, err := courseRepo.SelectCurrentCourse(t.Context(), user.ID, "en_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/vocab/summary", nil)
	req = setUserIDInContext(req, user.ID)
	w = httptest.NewRecorder()
	router.handleVocabSummary(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("summary status=%d body=%s", w.Code, w.Body.String())
	}
	var scoped map[string]int
	if err := json.Unmarshal(w.Body.Bytes(), &scoped); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if scoped["total"] != 1 {
		t.Fatalf("course-scoped total=%d want 1 (en_ru only), resp=%+v", scoped["total"], scoped)
	}

	// Direct queryVocabSummary without filter returns both words.
	all, err := router.queryVocabSummary(user.ID, "")
	if err != nil {
		t.Fatalf("queryVocabSummary all: %v", err)
	}
	if all.Total != 2 {
		t.Fatalf("unscoped total=%d want 2", all.Total)
	}
}

func TestQueryVocabSummary_NoRows(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	router := NewRouter(logger, &config.Config{}, db, nil, nil, service.NewOptionsService(repository.NewTrainingCardRepository(db, logger), logger, "en"), nil)
	userRepo := repository.NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(424244)

	summary, err := router.queryVocabSummary(user.ID, "en_ru")
	if err != nil {
		t.Fatalf("queryVocabSummary empty user: %v", err)
	}
	if summary.Total != 0 {
		t.Fatalf("expected zero summary, got %+v", summary)
	}
}

func TestVocabWordSummaryBucket_Clamp(t *testing.T) {
	if got := vocabWordSummaryBucket(-1, false); got != 0 {
		t.Fatalf("clamp low = %d", got)
	}
	if got := vocabWordSummaryBucket(99, false); got != 2 {
		t.Fatalf("clamp high = %d", got)
	}
}
