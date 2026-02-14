//go:build integration

package user_flows

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/integration/testkit"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestWordToTraining_LLMMock_CardCreatedAndHistory(t *testing.T) {
	wordJSON := testkit.DefaultWordJSON("intword", "интеграционное слово")
	mockLLM := testkit.StartMockLLMServer(t, wordJSON)

	h := testkit.NewHarness(t, testkit.WithAIService(mockLLM.URL, "test-key", "prompt"))
	telegramID := int64(99999)
	user := testkit.UserFixture(t, h.GetConnection().GetConnection(), telegramID)
	token := h.AuthAsUser(telegramID)

	// Request word via chat (triggers LLM -> word_card, word_request_history)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader("message=intword"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("chat: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	conn := h.GetConnection().GetConnection()

	// Assert word_request_history has entry for user
	var historyCount int
	err := conn.QueryRow(`SELECT COUNT(*) FROM word_request_history WHERE user_id = $1`, user.ID).Scan(&historyCount)
	if err != nil {
		t.Fatalf("history query: %v", err)
	}
	if historyCount == 0 {
		t.Error("expected at least one word_request_history entry for user")
	}

	// Assert word_cards has the lemma
	var wordCardID int64
	err = conn.QueryRow(`SELECT id FROM word_cards WHERE LOWER(word) = 'intword'`).Scan(&wordCardID)
	if err != nil || wordCardID == 0 {
		t.Fatalf("word_cards: expected 'intword', err=%v", err)
	}

	// Add training_card via repository (simulates worker); then re-request -> ensureUserCardsForWord
	logger := zap.NewNop()
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	_, err = tcRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "intword",
		SenseIndex: 0,
		WordRU:     "интеграционное слово",
		MeaningEN:  "integration word",
	})
	if err != nil {
		t.Fatalf("create training card: %v", err)
	}

	// Second request: word found in DB -> ensureUserCardsForWord creates user_cards
	req2 := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader("message=intword"))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Authorization", token)
	w2 := httptest.NewRecorder()
	h.Router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second chat: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var userCardsCount int
	err = conn.QueryRow(`
		SELECT COUNT(*) FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		WHERE uc.user_id = $1 AND tc.word_card_id = $2
	`, user.ID, wordCardID).Scan(&userCardsCount)
	if err != nil {
		t.Fatalf("user_cards query: %v", err)
	}
	if userCardsCount < 2 {
		t.Errorf("expected at least 2 user_cards (both directions), got %d", userCardsCount)
	}
}

func TestWordToTraining_RepeatRequest_Idempotent(t *testing.T) {
	wordJSON := testkit.DefaultWordJSON("idemword", "идемпотентное слово")
	mockLLM := testkit.StartMockLLMServer(t, wordJSON)

	h := testkit.NewHarness(t, testkit.WithAIService(mockLLM.URL, "test-key", "prompt"))
	telegramID := int64(99998)
	_ = testkit.UserFixture(t, h.GetConnection().GetConnection(), telegramID)
	token := h.AuthAsUser(telegramID)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader("message=idemword"))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1.Header.Set("Authorization", token)
	w1 := httptest.NewRecorder()
	h.Router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first chat: expected 200, got %d: %s", w1.Code, w1.Body.String())
	}

	// Second request - should not create duplicate user_cards (UNIQUE constraint)
	req2 := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader("message=idemword"))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Authorization", token)
	w2 := httptest.NewRecorder()
	h.Router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second chat: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	conn := h.GetConnection().GetConnection()
	var userCardsCount int
	err := conn.QueryRow(`
		SELECT COUNT(*) FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = $1 AND LOWER(wc.word) = 'idemword'
	`, telegramID).Scan(&userCardsCount)
	if err != nil {
		t.Fatalf("user_cards query: %v", err)
	}
	if userCardsCount > 2 {
		t.Errorf("expected at most 2 user_cards (both directions), got %d (idempotency violated)", userCardsCount)
	}
}
