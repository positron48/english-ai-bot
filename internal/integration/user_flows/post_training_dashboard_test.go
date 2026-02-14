//go:build integration

package user_flows

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/integration/testkit"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestPostTraining_DashboardCountersReflectState(t *testing.T) {
	h := testkit.NewHarness(t)
	conn := h.GetConnection().GetConnection() // *sql.DB for fixtures and assertions
	logger := zap.NewNop()

	telegramID := int64(77777)
	user := testkit.UserFixture(t, conn, telegramID)

	settings := models.UserSettings{
		SpellModeEnabled: ptrBool(false),
		TypeModeEnabled:  ptrBool(false),
	}
	settingsJSON, _ := json.Marshal(settings)
	userRepo := repository.NewUserRepository(conn, logger)
	if err := userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	// 10 words = 20 user_cards
	testkit.TrainingDeckFixture(t, conn, user.ID, 10)

	token := h.AuthAsUser(telegramID)

	// Dashboard before training: 20 new, 0 due/learning/review
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("dashboard before: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	testkit.AssertDashboardCounters(t, w, 0, 20, 0, 0)

	// Run one full session (20 cards)
	n := runTrainingSession(t, h, token, -1)
	if n < 10 {
		t.Fatalf("expected at least 10 cards answered, got %d", n)
	}

	// Dashboard after training: new reduced, learning/review increased
	req2 := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req2.Header.Set("Authorization", token)
	w2 := httptest.NewRecorder()
	h.Router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("dashboard after: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp struct {
		DueCount      int `json:"due_count"`
		NewCount      int `json:"new_count"`
		LearningCount int `json:"learning_count"`
		ReviewCount   int `json:"review_count"`
		TotalCards    int `json:"total_cards"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&resp); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	if resp.TotalCards != 20 {
		t.Errorf("total_cards: want 20, got %d", resp.TotalCards)
	}
	// After training: new + learning + review should sum to 20
	sum := resp.NewCount + resp.LearningCount + resp.ReviewCount
	if sum != 20 {
		t.Errorf("new+learning+review should equal 20, got %d+%d+%d=%d",
			resp.NewCount, resp.LearningCount, resp.ReviewCount, sum)
	}
	// Some cards should have moved out of new
	if resp.NewCount >= 20 {
		t.Error("expected some cards to leave new state after training")
	}

	testkit.AssertReviewEventsCount(t, conn, user.ID, n)
}
