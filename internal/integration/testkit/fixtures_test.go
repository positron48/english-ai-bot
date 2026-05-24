package testkit

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserFixture_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	telegramID := int64(11111)
	// GetUserByTelegramID: no existing user
	mock.ExpectQuery("SELECT id, telegram_id.*FROM users WHERE telegram_id").
		WithArgs(telegramID).
		WillReturnError(sql.ErrNoRows)
	// InsertAndReturnID: INSERT ... RETURNING id
	mock.ExpectQuery("INSERT INTO users \\(telegram_id\\) VALUES").
		WithArgs(telegramID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	// GetUserByID: return user row
	mock.ExpectQuery("SELECT id, telegram_id.*FROM users WHERE id").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "telegram_id", "telegram_username", "timezone", "preferred_training_time", "settings_json", "subscription_tier", "created_at", "updated_at"}).
			AddRow(int64(1), telegramID, "", "", "", "", "free", "2024-01-01 12:00:00", "2024-01-01 12:00:00"))

	user := UserFixture(t, db, telegramID)
	if user == nil {
		t.Fatal("UserFixture: expected non-nil user")
	}
	if user.ID != 1 || user.TelegramID != telegramID {
		t.Errorf("UserFixture: got user id=%d telegram_id=%d, want id=1 telegram_id=%d", user.ID, user.TelegramID, telegramID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestGrammarPublishFixture_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sectionID := "sec-1"
	chapterIDs := []string{"ch-a", "ch-b"}

	for _, cid := range chapterIDs {
		mock.ExpectExec("INSERT INTO grammar_published_items").
			WithArgs("chapter", cid).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectExec("INSERT INTO grammar_published_items").
		WithArgs("section", sectionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	GrammarPublishFixture(t, db, sectionID, chapterIDs)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestGrammarPublishFixture_SectionOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sectionID := "sec-only"
	chapterIDs := []string{} // empty: only section insert

	mock.ExpectExec("INSERT INTO grammar_published_items").
		WithArgs("section", sectionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	GrammarPublishFixture(t, db, sectionID, chapterIDs)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestGrammarProgressFixture_PassedTrue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userID := int64(100)
	chapterID := "ch-1"
	bestScore := 80
	passed := true

	mock.ExpectExec("INSERT INTO grammar_progress").
		WithArgs(userID, chapterID, bestScore, sqlmock.AnyArg()). // passedAt is *time.Time
		WillReturnResult(sqlmock.NewResult(0, 1))

	GrammarProgressFixture(t, db, userID, chapterID, bestScore, passed)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestGrammarProgressFixture_PassedFalse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userID := int64(200)
	chapterID := "ch-2"
	bestScore := 30
	passed := false

	mock.ExpectExec("INSERT INTO grammar_progress").
		WithArgs(userID, chapterID, bestScore, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))

	GrammarProgressFixture(t, db, userID, chapterID, bestScore, passed)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestGrammarProgressFixture_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		chapterID string
		bestScore int
		passed    bool
	}{
		{"passed_true", 1, "c1", 100, true},
		{"passed_false", 2, "c2", 40, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			var fourthArg interface{} = nil
			if tt.passed {
				fourthArg = sqlmock.AnyArg()
			}
			mock.ExpectExec("INSERT INTO grammar_progress").
				WithArgs(tt.userID, tt.chapterID, tt.bestScore, fourthArg).
				WillReturnResult(sqlmock.NewResult(0, 1))

			GrammarProgressFixture(t, db, tt.userID, tt.chapterID, tt.bestScore, tt.passed)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations: %v", err)
			}
		})
	}
}

func TestGrammarPublishFixture_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		sectionID string
		chapters  []string
	}{
		{"one_chapter", "s1", []string{"ch1"}},
		{"two_chapters", "s2", []string{"ch1", "ch2"}},
		{"no_chapters", "s3", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			for _, cid := range tt.chapters {
				mock.ExpectExec("INSERT INTO grammar_published_items").
					WithArgs("chapter", cid).
					WillReturnResult(sqlmock.NewResult(0, 1))
			}
			mock.ExpectExec("INSERT INTO grammar_published_items").
				WithArgs("section", tt.sectionID).
				WillReturnResult(sqlmock.NewResult(0, 1))

			GrammarPublishFixture(t, db, tt.sectionID, tt.chapters)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations: %v", err)
			}
		})
	}
}

// TestUserFixture_GetOrCreateUser_ExistingUser covers the path when user already exists (GetUserByTelegramID returns a row).
func TestUserFixture_GetOrCreateUser_ExistingUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	telegramID := int64(22222)
	mock.ExpectQuery("SELECT id, telegram_id.*FROM users WHERE telegram_id").
		WithArgs(telegramID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "telegram_id", "telegram_username", "timezone", "preferred_training_time", "settings_json", "subscription_tier", "created_at", "updated_at"}).
			AddRow(int64(2), telegramID, "", "", "", "", "free", "2024-01-01 12:00:00", "2024-01-01 12:00:00"))

	user := UserFixture(t, db, telegramID)
	if user == nil {
		t.Fatal("UserFixture: expected non-nil user")
	}
	if user.ID != 2 || user.TelegramID != telegramID {
		t.Errorf("UserFixture: got user id=%d telegram_id=%d, want id=2 telegram_id=%d", user.ID, user.TelegramID, telegramID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}
