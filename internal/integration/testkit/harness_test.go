//go:build integration

package testkit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHarness_Smoke(t *testing.T) {
	h := NewHarness(t)
	token := h.AuthAsUser(12345)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("dashboard: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHarness_WithAIService(t *testing.T) {
	wordJSON := DefaultWordJSON("mock", "мок")
	srv := StartMockLLMServer(t, wordJSON)
	h := NewHarness(t, WithAIService(srv.URL, "key", "Prompt"))
	_ = h
	// Harness created with mock AI; smoke check
	token := h.AuthAsUser(99999)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("dashboard: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHarness_FixturesAndAssertions(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()

	user := UserFixture(t, conn, 88888)
	if user == nil || user.ID == 0 {
		t.Fatal("UserFixture: expected non-nil user with ID")
	}

	wordCardID, trainingCardIDs := WordFixture(t, conn, "fixture-word", "definition", "слово")
	if wordCardID == 0 || len(trainingCardIDs) == 0 {
		t.Fatal("WordFixture: expected word card and training card IDs")
	}

	wordCardIDs, trainingCardIDs2 := TrainingDeckFixture(t, conn, user.ID, 2)
	if len(wordCardIDs) != 2 || len(trainingCardIDs2) != 2 {
		t.Errorf("TrainingDeckFixture: want 2 word cards and 2 training cards, got %d, %d", len(wordCardIDs), len(trainingCardIDs2))
	}

	AssertUserCardsState(t, conn, user.ID, map[string]string{"en_ru": "new", "ru_en": "new"})

	var userCardID int64
	err := conn.QueryRow(`SELECT id FROM user_cards WHERE user_id = $1 ORDER BY id LIMIT 1`, user.ID).Scan(&userCardID)
	if err != nil {
		t.Fatalf("query user_cards: %v", err)
	}
	AssertUserCardFields(t, conn, userCardID, map[string]interface{}{
		"state": "new", "ef": 2.5, "reps": 0, "interval_days": 0, "lapse_count": 0,
	})

	AssertReviewEventsCount(t, conn, user.ID, 0)

	GrammarPublishFixture(t, conn, "test-section", []string{"test-chapter"})
	GrammarProgressFixture(t, conn, user.ID, "test-chapter", 60, true)
	AssertGrammarChapterAccess(t, conn, user.ID, "test-chapter", true)

	// AssertNextDueWithin: set next_due_at then assert
	now := time.Now().UTC()
	_, err = conn.Exec(`UPDATE user_cards SET next_due_at = $1 WHERE id = $2`, now.Add(time.Minute), userCardID)
	if err != nil {
		t.Fatalf("update next_due_at: %v", err)
	}
	AssertNextDueWithin(t, conn, userCardID, now, now.Add(2*time.Minute))

	// Dashboard counters after fixtures (new cards only, no training yet)
	token := h.AuthAsUser(88888)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("dashboard: %d %s", w.Code, w.Body.String())
	}
	// 4 new user_cards from TrainingDeckFixture (2 words × 2 directions) for this user
	AssertDashboardCounters(t, w, 0, 4, 0, 0)
}

func TestHarness_AssertGrammarChapterAccess_NotPassed(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	user := UserFixture(t, conn, 77777)
	GrammarProgressFixture(t, conn, user.ID, "ch-low", 30, false)
	AssertGrammarChapterAccess(t, conn, user.ID, "ch-low", false)
}

func TestHarness_AssertReviewEventsCount_NonZero(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	user := UserFixture(t, conn, 66666)
	TrainingDeckFixture(t, conn, user.ID, 1) // creates user_cards
	var userCardID int64
	err := conn.QueryRow(`SELECT id FROM user_cards WHERE user_id = $1 LIMIT 1`, user.ID).Scan(&userCardID)
	if err != nil {
		t.Fatalf("get user_card id: %v", err)
	}
	_, err = conn.Exec(`INSERT INTO review_events (user_id, user_card_id, direction, answered_at, is_correct) VALUES ($1, $2, 'en_ru', NOW(), 1)`, user.ID, userCardID)
	if err != nil {
		t.Fatalf("insert review_event: %v", err)
	}
	AssertReviewEventsCount(t, conn, user.ID, 1)
}

// TestNewHarness_AndAuthAsUser covers NewHarness (including StartPostgres) and AuthAsUser.
func TestNewHarness_AndAuthAsUser(t *testing.T) {
	h := NewHarness(t)
	if h == nil || h.DB == nil || h.Router == nil {
		t.Fatal("NewHarness: expected non-nil Harness, DB and Router")
	}
	token := h.AuthAsUser(11111)
	if token == "" || token[:7] != "Bearer " {
		t.Errorf("AuthAsUser: expected Bearer token, got %q", token)
	}
}

// TestAuthAsUser_TableDriven covers AuthAsUser with multiple telegram IDs (integration).
func TestAuthAsUser_TableDriven(t *testing.T) {
	h := NewHarness(t)
	tests := []struct {
		name       string
		telegramID int64
	}{
		{"user_1", 90001},
		{"user_2", 90002},
		{"user_3", 90003},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := h.AuthAsUser(tt.telegramID)
			if token == "" {
				t.Error("AuthAsUser: expected non-empty token")
			}
			if len(token) < 7 || token[:7] != "Bearer " {
				t.Errorf("AuthAsUser: expected token to start with Bearer , got %q", token)
			}
		})
	}
}

// TestTrainingDeckFixture_TableDriven covers TrainingDeckFixture with different counts (integration).
func TestTrainingDeckFixture_TableDriven(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	user := UserFixture(t, conn, 44001)
	tests := []struct {
		name  string
		count int
	}{
		{"one_word", 1},
		{"two_words", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wordCardIDs, trainingCardIDs := TrainingDeckFixture(t, conn, user.ID, tt.count)
			if len(wordCardIDs) != tt.count {
				t.Errorf("TrainingDeckFixture(%d): wordCardIDs len = %d, want %d", tt.count, len(wordCardIDs), tt.count)
			}
			if len(trainingCardIDs) != tt.count {
				t.Errorf("TrainingDeckFixture(%d): trainingCardIDs len = %d, want %d", tt.count, len(trainingCardIDs), tt.count)
			}
		})
	}
}

// TestWordFixture_TableDriven covers WordFixture with different inputs (integration).
func TestWordFixture_TableDriven(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	tests := []struct {
		name       string
		word       string
		definition string
		wordRu     string
	}{
		{"word_1", "table-word-a", "def a", "слово а"},
		{"word_2", "table-word-b", "def b", "слово б"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wordCardID, trainingCardIDs := WordFixture(t, conn, tt.word, tt.definition, tt.wordRu)
			if wordCardID == 0 {
				t.Error("WordFixture: expected non-zero word_card_id")
			}
			if len(trainingCardIDs) == 0 {
				t.Error("WordFixture: expected at least one training_card_id")
			}
			var wordEN string
			err := conn.QueryRow(`SELECT word FROM word_cards WHERE id = $1`, wordCardID).Scan(&wordEN)
			if err != nil {
				t.Fatalf("WordFixture: word_cards row: %v", err)
			}
			if wordEN != tt.word {
				t.Errorf("WordFixture: word = %q, want %q", wordEN, tt.word)
			}
		})
	}
}

func TestUserFixture_CreatesOrReturnsUser(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	u := UserFixture(t, conn, 55555)
	if u == nil || u.ID == 0 {
		t.Fatal("UserFixture: expected non-nil user with ID")
	}
	if u.TelegramID != 55555 {
		t.Errorf("UserFixture: telegram_id want 55555, got %d", u.TelegramID)
	}
	// idempotent: same telegram ID returns same user
	u2 := UserFixture(t, conn, 55555)
	if u2.ID != u.ID {
		t.Errorf("UserFixture: same telegram_id should return same user, got id %d vs %d", u2.ID, u.ID)
	}
}

func TestWordFixture_CreatesWordAndTrainingCard(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	wordCardID, trainingCardIDs := WordFixture(t, conn, "unit-word", "unit def", "юнит-слово")
	if wordCardID == 0 {
		t.Fatal("WordFixture: expected non-zero word_card_id")
	}
	if len(trainingCardIDs) == 0 {
		t.Fatal("WordFixture: expected at least one training_card_id")
	}
	var wordEN string
	err := conn.QueryRow(`SELECT word FROM word_cards WHERE id = $1`, wordCardID).Scan(&wordEN)
	if err != nil {
		t.Fatalf("WordFixture: word_cards row: %v", err)
	}
	if wordEN != "unit-word" {
		t.Errorf("WordFixture: word want unit-word, got %q", wordEN)
	}
}

func TestTrainingDeckFixture_CreatesDeck(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	user := UserFixture(t, conn, 44444)
	wordCardIDs, trainingCardIDs := TrainingDeckFixture(t, conn, user.ID, 3)
	if len(wordCardIDs) != 3 || len(trainingCardIDs) != 3 {
		t.Errorf("TrainingDeckFixture(3): want 3 word cards and 3 training cards, got %d, %d", len(wordCardIDs), len(trainingCardIDs))
	}
	var count int
	err := conn.QueryRow(`SELECT COUNT(*) FROM user_cards WHERE user_id = $1`, user.ID).Scan(&count)
	if err != nil {
		t.Fatalf("TrainingDeckFixture: count user_cards: %v", err)
	}
	// 3 words × 2 directions = 6 user_cards
	if count < 6 {
		t.Errorf("TrainingDeckFixture: want at least 6 user_cards, got %d", count)
	}
}

func TestGrammarPublishFixture_PublishesChapters(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	GrammarPublishFixture(t, conn, "sec-1", []string{"ch-a", "ch-b"})
	var n int
	err := conn.QueryRow(`SELECT COUNT(*) FROM grammar_published_items WHERE item_type = 'chapter' AND item_id IN ('ch-a','ch-b') AND is_published = 1`).Scan(&n)
	if err != nil {
		t.Fatalf("GrammarPublishFixture: query: %v", err)
	}
	if n != 2 {
		t.Errorf("GrammarPublishFixture: want 2 published chapters, got %d", n)
	}
	err = conn.QueryRow(`SELECT COUNT(*) FROM grammar_published_items WHERE item_type = 'section' AND item_id = 'sec-1' AND is_published = 1`).Scan(&n)
	if err != nil {
		t.Fatalf("GrammarPublishFixture section: %v", err)
	}
	if n != 1 {
		t.Errorf("GrammarPublishFixture: want 1 published section, got %d", n)
	}
}

func TestGrammarProgressFixture_CreatesProgress(t *testing.T) {
	h := NewHarness(t)
	conn := h.GetConnection().GetConnection()
	user := UserFixture(t, conn, 33333)
	GrammarProgressFixture(t, conn, user.ID, "ch-progress", 75, true)
	var bestScore int
	var passedAt interface{}
	err := conn.QueryRow(`SELECT best_score, passed_at FROM grammar_progress WHERE user_id = $1 AND chapter_id = 'ch-progress'`, user.ID).Scan(&bestScore, &passedAt)
	if err != nil {
		t.Fatalf("GrammarProgressFixture: query: %v", err)
	}
	if bestScore != 75 {
		t.Errorf("GrammarProgressFixture: best_score want 75, got %d", bestScore)
	}
	if passedAt == nil {
		t.Error("GrammarProgressFixture: passed_at should be set when passed=true")
	}
}
