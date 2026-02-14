//go:build integration

package user_flows

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/integration/testkit"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// runTrainingSession runs one full session: start -> for each card: reveal -> answer -> current (next)
// Returns total cards answered.
func runTrainingSession(t *testing.T, h *testkit.Harness, token string, correctFirstN int) int {
	t.Helper()

	startReq := httptest.NewRequest(http.MethodPost, "/api/training/start", nil)
	startReq.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, startReq)
	if w.Code != http.StatusOK {
		t.Fatalf("start: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var startResp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&startResp); err != nil {
		t.Fatalf("decode start: %v", err)
	}
	// No cards: start returns 400
	if startResp["complete"] == true {
		return 0
	}
	totalCards, _ := startResp["total_cards"].(float64)
	if totalCards == 0 {
		return 0
	}

	answered := 0
	for i := 0; i < int(totalCards); i++ {
		// Reveal options for current card
		revealReq := httptest.NewRequest(http.MethodPost, "/api/training/reveal", nil)
		revealReq.Header.Set("Authorization", token)
		w2 := httptest.NewRecorder()
		h.Router.ServeHTTP(w2, revealReq)
		if w2.Code != http.StatusOK {
			t.Fatalf("reveal card %d: expected 200, got %d: %s", i+1, w2.Code, w2.Body.String())
		}
		var revealResp struct {
			Options    []string `json:"options"`
			UserCardID float64  `json:"user_card_id"`
		}
		if err := json.NewDecoder(w2.Body).Decode(&revealResp); err != nil {
			t.Fatalf("decode reveal: %v", err)
		}
		if len(revealResp.Options) == 0 {
			t.Fatalf("reveal card %d: no options", i+1)
		}

		optionIndex := 0
		if correctFirstN >= 0 && i >= correctFirstN && len(revealResp.Options) >= 2 {
			optionIndex = 1 // wrong (different pattern in second cycle)
		}

		time.Sleep(50 * time.Millisecond) // ensure optionsShownAt delay + spread requests to avoid rate limit
		doAnswer := func() *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPost, "/api/training/answer",
				strings.NewReader(fmt.Sprintf("option_index=%d&user_card_id=%.0f", optionIndex, revealResp.UserCardID)))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Authorization", token)
			w := httptest.NewRecorder()
			h.Router.ServeHTTP(w, req)
			return w
		}
		w3 := doAnswer()
		for w3.Code == http.StatusTooManyRequests {
			time.Sleep(1500 * time.Millisecond)
			w3 = doAnswer()
		}
		if w3.Code != http.StatusOK {
			t.Fatalf("answer card %d: expected 200, got %d: %s", i+1, w3.Code, w3.Body.String())
		}

		var answerResp map[string]interface{}
		if err := json.NewDecoder(w3.Body).Decode(&answerResp); err != nil {
			t.Fatalf("decode answer: %v", err)
		}
		answered++

		// Session complete: answer returns feedback, next card comes from current
		if answerResp["complete"] == true {
			break
		}

		// Get next card (or session finished)
		if i+1 < int(totalCards) {
			currentReq := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
			currentReq.Header.Set("Authorization", token)
			w4 := httptest.NewRecorder()
			h.Router.ServeHTTP(w4, currentReq)
			if w4.Code != http.StatusOK {
				t.Fatalf("current after card %d: expected 200, got %d: %s", i+1, w4.Code, w4.Body.String())
			}
			var currentResp map[string]interface{}
			if err := json.NewDecoder(w4.Body).Decode(&currentResp); err != nil {
				t.Fatalf("decode current: %v", err)
			}
			if currentResp["complete"] == true {
				break
			}
		}
	}
	return answered
}

func TestTrainingTwoCycles_60Cards_SRSAndReviewEvents(t *testing.T) {
	h := testkit.NewHarness(t)
	conn := h.GetConnection().GetConnection()
	logger := zap.NewNop()

	telegramID := int64(88888)
	user := testkit.UserFixture(t, conn, telegramID)

	// Disable spell/type so we get only MCQ cards
	settings := models.UserSettings{
		SpellModeEnabled: ptrBool(false),
		TypeModeEnabled:  ptrBool(false),
	}
	settingsJSON, _ := json.Marshal(settings)
	userRepo := repository.NewUserRepository(conn, logger)
	if err := userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	// 30 words = 60 user_cards (2 directions each)
	testkit.TrainingDeckFixture(t, conn, user.ID, 30)

	token := h.AuthAsUser(telegramID)

	// Session 1: 30 cards (MaxCardsPerSession)
	n1 := runTrainingSession(t, h, token, -1)
	if n1 < 1 {
		t.Fatalf("session 1: expected at least 1 card, got %d", n1)
	}

	// Session 2: remaining or another batch
	n2 := runTrainingSession(t, h, token, -1)
	totalAnswered := n1 + n2

	// Assert review_events created
	testkit.AssertReviewEventsCount(t, conn, user.ID, totalAnswered)

	// Assert training_sessions exist
	var sessionCount int
	err := conn.QueryRow(`SELECT COUNT(*) FROM training_sessions WHERE user_id = $1`, user.ID).Scan(&sessionCount)
	if err != nil {
		t.Fatalf("session count: %v", err)
	}
	if sessionCount < 1 {
		t.Error("expected at least one training_session")
	}
}

func ptrBool(b bool) *bool {
	return &b
}
