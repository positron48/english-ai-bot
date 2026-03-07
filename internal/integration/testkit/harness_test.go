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
