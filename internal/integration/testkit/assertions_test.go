package testkit

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestIntVal(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want int
	}{
		{"int", 42, 42},
		{"int64", int64(100), 100},
		{"float64", float64(7.9), 7},
		{"float64_whole", float64(12.0), 12},
		{"nil_other", "x", 0},
		{"bool", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intVal(tt.v)
			if got != tt.want {
				t.Errorf("intVal(%v) = %d, want %d", tt.v, got, tt.want)
			}
		})
	}
}

func TestAssertDashboardCounters(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		body         string
		wantDue      int
		wantNew      int
		wantLearning int
		wantReview   int
		shouldFail   bool
	}{
		{
			name:         "ok",
			code:         http.StatusOK,
			body:         `{"due_count":1,"new_count":2,"learning_count":3,"review_count":4}`,
			wantDue:      1,
			wantNew:      2,
			wantLearning: 3,
			wantReview:   4,
			shouldFail:   false,
		},
		{
			name:         "wrong_status",
			code:         http.StatusNotFound,
			body:         `{}`,
			wantDue:      0,
			wantNew:      0,
			wantLearning: 0,
			wantReview:   0,
			shouldFail:   true,
		},
		{
			name:         "zero_counts",
			code:         http.StatusOK,
			body:         `{"due_count":0,"new_count":0,"learning_count":0,"review_count":0}`,
			wantDue:      0,
			wantNew:      0,
			wantLearning: 0,
			wantReview:   0,
			shouldFail:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			w.WriteHeader(tt.code)
			_, _ = w.Write([]byte(tt.body))

			fakeT := &testing.T{}
			AssertDashboardCounters(fakeT, w, tt.wantDue, tt.wantNew, tt.wantLearning, tt.wantReview)
			failed := fakeT.Failed()
			if failed != tt.shouldFail {
				t.Errorf("AssertDashboardCounters: expected fail=%v, got fail=%v", tt.shouldFail, failed)
			}
		})
	}
}

func TestAssertDashboardCounters_ValidBody(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"due_count":10,"new_count":20,"learning_count":30,"review_count":40}`))
	AssertDashboardCounters(t, w, 10, 20, 30, 40)
}

func TestAssertUserCardsState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT direction, state FROM user_cards WHERE user_id = \$1 ORDER BY id`).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"direction", "state"}).
				AddRow("en_ru", "new").
				AddRow("ru_en", "new"))
		fakeT := &testing.T{}
		AssertUserCardsState(fakeT, db, 1, map[string]string{"en_ru": "new", "ru_en": "new"})
		if fakeT.Failed() {
			t.Error("AssertUserCardsState should not fail when expected matches")
		}
	})

	t.Run("wrong_state", func(t *testing.T) {
		mock.ExpectQuery(`SELECT direction, state FROM user_cards WHERE user_id = \$1 ORDER BY id`).
			WithArgs(int64(2)).
			WillReturnRows(sqlmock.NewRows([]string{"direction", "state"}).
				AddRow("en_ru", "new"))
		fakeT := &testing.T{}
		AssertUserCardsState(fakeT, db, 2, map[string]string{"en_ru": "review"})
		if !fakeT.Failed() {
			t.Error("AssertUserCardsState should fail when state does not match")
		}
	})
}

func TestAssertUserCardFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userCardCols := []string{"id", "user_id", "training_card_id", "direction", "state", "ef", "reps",
		"interval_days", "learning_step", "lapse_count", "next_due_at", "last_review_at", "last_quality",
		"last_options_json", "wrong_answers_json", "stats_json", "created_at", "updated_at"}
	ts := "2024-01-01 12:00:00"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, user_id, training_card_id, direction, state, ef, reps").
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows(userCardCols).
				AddRow(int64(10), int64(1), int64(1), "en_ru", "new", 2.5, 0, 0, 0, 0, nil, nil, nil, "", "", "", ts, ts))
		fakeT := &testing.T{}
		AssertUserCardFields(fakeT, db, 10, map[string]interface{}{"state": "new", "ef": 2.5, "reps": 0, "interval_days": 0, "lapse_count": 0})
		if fakeT.Failed() {
			t.Error("AssertUserCardFields should not fail when fields match")
		}
	})

	t.Run("wrong_state", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, user_id, training_card_id, direction, state, ef, reps").
			WithArgs(int64(11)).
			WillReturnRows(sqlmock.NewRows(userCardCols).
				AddRow(int64(11), int64(1), int64(1), "en_ru", "learning", 2.5, 1, 1, 0, 0, nil, nil, nil, "", "", "", ts, ts))
		fakeT := &testing.T{}
		AssertUserCardFields(fakeT, db, 11, map[string]interface{}{"state": "new"})
		if !fakeT.Failed() {
			t.Error("AssertUserCardFields should fail when state does not match")
		}
	})
}

func TestAssertReviewEventsCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM review_events WHERE user_id = \\$1").
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
		fakeT := &testing.T{}
		AssertReviewEventsCount(fakeT, db, 1, 3)
		if fakeT.Failed() {
			t.Error("AssertReviewEventsCount should not fail when count matches")
		}
	})

	t.Run("wrong_count", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM review_events WHERE user_id = \\$1").
			WithArgs(int64(2)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		fakeT := &testing.T{}
		AssertReviewEventsCount(fakeT, db, 2, 5)
		if !fakeT.Failed() {
			t.Error("AssertReviewEventsCount should fail when count does not match")
		}
	})
}

func TestAssertGrammarChapterAccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	t.Run("passed", func(t *testing.T) {
		now := time.Now()
		mock.ExpectQuery("SELECT passed_at, best_score FROM grammar_progress WHERE user_id = \\$1 AND chapter_id = \\$2").
			WithArgs(int64(1), "ch1").
			WillReturnRows(sqlmock.NewRows([]string{"passed_at", "best_score"}).AddRow(now, 60))
		fakeT := &testing.T{}
		AssertGrammarChapterAccess(fakeT, db, 1, "ch1", true)
		if fakeT.Failed() {
			t.Error("AssertGrammarChapterAccess should not fail when passed=true and progress exists with best_score>=50")
		}
	})

	t.Run("not_passed", func(t *testing.T) {
		mock.ExpectQuery("SELECT passed_at, best_score FROM grammar_progress WHERE user_id = \\$1 AND chapter_id = \\$2").
			WithArgs(int64(2), "ch2").
			WillReturnRows(sqlmock.NewRows([]string{"passed_at", "best_score"}).AddRow(nil, 30))
		fakeT := &testing.T{}
		AssertGrammarChapterAccess(fakeT, db, 2, "ch2", false)
		if fakeT.Failed() {
			t.Error("AssertGrammarChapterAccess should not fail when passed=false and best_score<50")
		}
	})

	t.Run("no_rows_expected_passed", func(t *testing.T) {
		mock.ExpectQuery("SELECT passed_at, best_score FROM grammar_progress WHERE user_id = \\$1 AND chapter_id = \\$2").
			WithArgs(int64(3), "ch3").
			WillReturnError(sql.ErrNoRows)
		fakeT := &testing.T{}
		AssertGrammarChapterAccess(fakeT, db, 3, "ch3", true)
		if !fakeT.Failed() {
			t.Error("AssertGrammarChapterAccess should fail when no row but wantPassed=true")
		}
	})
}

func TestAssertNextDueWithin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	from := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	within := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT next_due_at FROM user_cards WHERE id = \\$1").
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"next_due_at"}).AddRow(within))
		fakeT := &testing.T{}
		AssertNextDueWithin(fakeT, db, 1, from, to)
		if fakeT.Failed() {
			t.Error("AssertNextDueWithin should not fail when next_due_at is within window")
		}
	})

	t.Run("nil_next_due", func(t *testing.T) {
		mock.ExpectQuery("SELECT next_due_at FROM user_cards WHERE id = \\$1").
			WithArgs(int64(2)).
			WillReturnRows(sqlmock.NewRows([]string{"next_due_at"}).AddRow(nil))
		fakeT := &testing.T{}
		AssertNextDueWithin(fakeT, db, 2, from, to)
		if !fakeT.Failed() {
			t.Error("AssertNextDueWithin should fail when next_due_at is nil")
		}
	})

	t.Run("outside_window", func(t *testing.T) {
		outside := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		mock.ExpectQuery("SELECT next_due_at FROM user_cards WHERE id = \\$1").
			WithArgs(int64(3)).
			WillReturnRows(sqlmock.NewRows([]string{"next_due_at"}).AddRow(outside))
		fakeT := &testing.T{}
		AssertNextDueWithin(fakeT, db, 3, from, to)
		if !fakeT.Failed() {
			t.Error("AssertNextDueWithin should fail when next_due_at is outside window")
		}
	})
}
