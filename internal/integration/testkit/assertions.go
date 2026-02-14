package testkit

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// AssertUserCardsState checks user_cards states for given user
func AssertUserCardsState(t *testing.T, conn *sql.DB, userID int64, expectedByDirection map[string]string) {
	t.Helper()
	rows, err := conn.Query(
		`SELECT direction, state FROM user_cards WHERE user_id = $1 ORDER BY id`,
		userID,
	)
	if err != nil {
		t.Fatalf("AssertUserCardsState query: %v", err)
	}
	defer rows.Close()
	got := make(map[string]string)
	for rows.Next() {
		var dir, state string
		if err := rows.Scan(&dir, &state); err != nil {
			t.Fatalf("AssertUserCardsState scan: %v", err)
		}
		got[dir] = state
	}
	for k, want := range expectedByDirection {
		if g, ok := got[k]; !ok || g != want {
			t.Errorf("AssertUserCardsState direction=%s: want %q, got %q", k, want, g)
		}
	}
}

// AssertUserCardFields checks key SRS fields of a user card
func AssertUserCardFields(t *testing.T, conn *sql.DB, userCardID int64, want map[string]interface{}) {
	t.Helper()
	logger := zap.NewNop()
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	uc, err := userCardRepo.GetUserCard(userCardID)
	if err != nil {
		t.Fatalf("AssertUserCardFields get: %v", err)
	}
	if uc == nil {
		t.Fatalf("AssertUserCardFields: user_card %d not found", userCardID)
	}
	if s, ok := want["state"]; ok && string(uc.State) != s.(string) {
		t.Errorf("user_card %d state: want %q, got %q", userCardID, s, uc.State)
	}
	if v, ok := want["ef"]; ok {
		var expected float64
		switch x := v.(type) {
		case float64:
			expected = x
		case int:
			expected = float64(x)
		default:
			expected = v.(float64)
		}
		if uc.EF != expected {
			t.Errorf("user_card %d ef: want %v, got %v", userCardID, expected, uc.EF)
		}
	}
	if v, ok := want["reps"]; ok {
		expected := intVal(v)
		if uc.Reps != expected {
			t.Errorf("user_card %d reps: want %v, got %v", userCardID, expected, uc.Reps)
		}
	}
	if v, ok := want["interval_days"]; ok {
		expected := intVal(v)
		if uc.IntervalDays != expected {
			t.Errorf("user_card %d interval_days: want %v, got %v", userCardID, expected, uc.IntervalDays)
		}
	}
	if v, ok := want["lapse_count"]; ok {
		expected := intVal(v)
		if uc.LapseCount != expected {
			t.Errorf("user_card %d lapse_count: want %v, got %v", userCardID, expected, uc.LapseCount)
		}
	}
}

func intVal(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	default:
		return 0
	}
}

// AssertReviewEventsCount checks number of review_events for user
func AssertReviewEventsCount(t *testing.T, conn *sql.DB, userID int64, want int) {
	t.Helper()
	var got int
	err := conn.QueryRow(`SELECT COUNT(*) FROM review_events WHERE user_id = $1`, userID).Scan(&got)
	if err != nil {
		t.Fatalf("AssertReviewEventsCount: %v", err)
	}
	if got != want {
		t.Errorf("AssertReviewEventsCount: want %d, got %d", want, got)
	}
}

// AssertGrammarChapterAccess checks grammar_progress for user+chapter (passed = best_score>=50 and passed_at not null)
func AssertGrammarChapterAccess(t *testing.T, conn *sql.DB, userID int64, chapterID string, wantPassed bool) {
	t.Helper()
	var passedAt sql.NullTime
	var bestScore int
	err := conn.QueryRow(
		`SELECT passed_at, best_score FROM grammar_progress WHERE user_id = $1 AND chapter_id = $2`,
		userID, chapterID,
	).Scan(&passedAt, &bestScore)
	if err == sql.ErrNoRows {
		if wantPassed {
			t.Errorf("AssertGrammarChapterAccess: chapter %s not in progress, expected passed", chapterID)
		}
		return
	}
	if err != nil {
		t.Fatalf("AssertGrammarChapterAccess: %v", err)
	}
	passed := passedAt.Valid && bestScore >= 50
	if passed != wantPassed {
		t.Errorf("AssertGrammarChapterAccess chapter=%s: want passed=%v, got %v (passed_at_valid=%v, best_score=%d)", chapterID, wantPassed, passed, passedAt.Valid, bestScore)
	}
}

// AssertDashboardCounters checks dashboard response counters
func AssertDashboardCounters(t *testing.T, w *httptest.ResponseRecorder, wantDue, wantNew, wantLearning, wantReview int) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Errorf("AssertDashboardCounters: expected 200, got %d: %s", w.Code, w.Body.String())
		return
	}
	var resp struct {
		DueCount      int `json:"due_count"`
		NewCount      int `json:"new_count"`
		LearningCount int `json:"learning_count"`
		ReviewCount   int `json:"review_count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("AssertDashboardCounters decode: %v", err)
	}
	if resp.DueCount != wantDue {
		t.Errorf("dashboard due_count: want %d, got %d", wantDue, resp.DueCount)
	}
	if resp.NewCount != wantNew {
		t.Errorf("dashboard new_count: want %d, got %d", wantNew, resp.NewCount)
	}
	if resp.LearningCount != wantLearning {
		t.Errorf("dashboard learning_count: want %d, got %d", wantLearning, resp.LearningCount)
	}
	if resp.ReviewCount != wantReview {
		t.Errorf("dashboard review_count: want %d, got %d", wantReview, resp.ReviewCount)
	}
}

// AssertNextDueWithin checks next_due_at is within a window
func AssertNextDueWithin(t *testing.T, conn *sql.DB, userCardID int64, from, to time.Time) {
	t.Helper()
	var nextDue *time.Time
	err := conn.QueryRow(`SELECT next_due_at FROM user_cards WHERE id = $1`, userCardID).Scan(&nextDue)
	if err != nil {
		t.Fatalf("AssertNextDueWithin: %v", err)
	}
	if nextDue == nil {
		t.Errorf("user_card %d: next_due_at is nil", userCardID)
		return
	}
	if nextDue.Before(from) || nextDue.After(to) {
		t.Errorf("user_card %d next_due_at %v outside [%v, %v]", userCardID, *nextDue, from, to)
	}
}
