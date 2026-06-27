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

func TestHandleVocabSummary_BreakdownCounts(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(424242)

	cfg := &config.Config{}
	cfg.WebApp.JWTSecret = "test-secret"
	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	conn := db.GetConnection()
	insertWordWithState := func(word, state string) {
		t.Helper()
		var wordCardID int64
		if err := conn.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", word, word).Scan(&wordCardID); err != nil {
			t.Fatalf("insert word_cards: %v", err)
		}
		var trainingCardID int64
		if err := conn.QueryRow(
			`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en)
			 VALUES ($1, $2, 0, $3, $4) RETURNING id`,
			wordCardID, word, word+"_ru", word,
		).Scan(&trainingCardID); err != nil {
			t.Fatalf("insert training_cards: %v", err)
		}
		if _, err := conn.Exec(
			`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'ru_en', $3)`,
			user.ID, trainingCardID, state,
		); err != nil {
			t.Fatalf("insert user_cards: %v", err)
		}
	}

	insertWordWithState("newword", "new")
	insertWordWithState("learningword", "learning")
	insertWordWithState("reviewword", "review")

	var masteredWordCardID int64
	if err := conn.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "masteredword", "masteredword").Scan(&masteredWordCardID); err != nil {
		t.Fatalf("insert mastered word_cards: %v", err)
	}
	if _, err := conn.Exec(
		`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES ($1, $2, 'known')`,
		user.ID, masteredWordCardID,
	); err != nil {
		t.Fatalf("insert user_word_knowledge: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/vocab/summary", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleVocabSummary(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]int
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["total"] != 3 {
		t.Fatalf("total = %d, want 3", resp["total"])
	}
	if resp["new"] != 1 || resp["learning"] != 1 || resp["review"] != 1 {
		t.Fatalf("breakdown = %+v, want 1 new/learning/review each", resp)
	}
	if resp["mastered"] != 1 || resp["known"] != 1 {
		t.Fatalf("mastered/known = %+v, want 1 mastered", resp)
	}
}
