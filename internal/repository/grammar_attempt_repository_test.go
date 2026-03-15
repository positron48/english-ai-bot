package repository

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupGrammarAttemptRepo(t *testing.T) *GrammarAttemptRepository {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	return NewGrammarAttemptRepository(conn, logger)
}

func TestGrammarAttemptRepository_CreateAttempt_WithNilFinishedAt(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(1)
	attempt := &TestAttempt{
		UserID:         user.ID,
		ScopeType:      "chapter",
		ScopeID:        "ch-nil-finished",
		StartedAt:      time.Now(),
		FinishedAt:     nil, // nil branch
		Score:          50,
		Passed:         true,
		TotalQuestions: 5,
		AnswersJSON:    "[]",
		ResultsJSON:    "[]",
	}
	id, err := repo.CreateAttempt(attempt)
	if err != nil {
		t.Fatalf("CreateAttempt(nil FinishedAt) error: %v", err)
	}
	if id == 0 {
		t.Error("CreateAttempt() should return non-zero ID")
	}
}

func TestGrammarAttemptRepository_CreateAndProgress(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(1)

	finished := time.Now()
	attempt := &TestAttempt{
		UserID:         user.ID,
		ScopeType:      "chapter",
		ScopeID:        "chapter-1",
		StartedAt:      time.Now(),
		FinishedAt:     &finished,
		Score:          80,
		Passed:         true,
		TotalQuestions: 10,
		AnswersJSON:    "[]",
		ResultsJSON:    "[]",
	}

	if _, err := repo.CreateAttempt(attempt); err != nil {
		t.Fatalf("CreateAttempt error: %v", err)
	}

	if err := repo.UpdateProgress(user.ID, "chapter-1", 80, true); err != nil {
		t.Fatalf("UpdateProgress error: %v", err)
	}

	progress, err := repo.GetChapterProgress(user.ID, "chapter-1")
	if err != nil {
		t.Fatalf("GetChapterProgress error: %v", err)
	}
	if progress.BestScore != 80 {
		t.Fatalf("expected best score 80, got %d", progress.BestScore)
	}
}

func TestGrammarAttemptRepository_CategoryProgress(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(2)

	finished := time.Now()
	attempt := &TestAttempt{
		UserID:         user.ID,
		ScopeType:      "category",
		ScopeID:        "section-1",
		StartedAt:      time.Now(),
		FinishedAt:     &finished,
		Score:          55,
		Passed:         true,
		TotalQuestions: 5,
		AnswersJSON:    "[]",
		ResultsJSON:    "[]",
	}

	if _, err := repo.CreateAttempt(attempt); err != nil {
		t.Fatalf("CreateAttempt error: %v", err)
	}

	hasAttempt, err := repo.HasCategoryTestAttempt(user.ID, "section-1")
	if err != nil {
		t.Fatalf("HasCategoryTestAttempt error: %v", err)
	}
	if !hasAttempt {
		t.Fatalf("expected category attempt")
	}

	passed, err := repo.GetCategoryTestProgress(user.ID, "section-1")
	if err != nil {
		t.Fatalf("GetCategoryTestProgress error: %v", err)
	}
	if !passed {
		t.Fatalf("expected category test to be passed")
	}

	bestScore, err := repo.GetCategoryTestBestScore(user.ID, "section-1")
	if err != nil {
		t.Fatalf("GetCategoryTestBestScore error: %v", err)
	}
	if bestScore != 55 {
		t.Fatalf("expected best score 55, got %d", bestScore)
	}

	// No attempts for other section — bestScore.Valid false, returns 0
	bestEmpty, err := repo.GetCategoryTestBestScore(user.ID, "section-none")
	if err != nil {
		t.Fatalf("GetCategoryTestBestScore(empty) error: %v", err)
	}
	if bestEmpty != 0 {
		t.Fatalf("expected 0 for no attempts, got %d", bestEmpty)
	}
}

func TestGrammarAttemptRepository_UpdateCategoryTestProgress(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(3)

	// UpdateCategoryTestProgress is a no-op (attempt is saved via CreateAttempt); just ensure it does not error
	if err := repo.UpdateCategoryTestProgress(user.ID, "cat-1", 70, true); err != nil {
		t.Fatalf("UpdateCategoryTestProgress error: %v", err)
	}
}

func TestGrammarAttemptRepository_GetUserAttempts(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(4)

	finished := time.Now()
	// started_at decreases so the last inserted has the latest started_at (first in ORDER BY started_at DESC)
	for i := 0; i < 3; i++ {
		attempt := &TestAttempt{
			UserID:         user.ID,
			ScopeType:      "chapter",
			ScopeID:        "ch-get",
			StartedAt:      time.Now().Add(-time.Duration(2-i) * time.Hour),
			FinishedAt:     &finished,
			Score:          60 + i*10,
			Passed:         true,
			TotalQuestions: 5,
			AnswersJSON:    "{}",
			ResultsJSON:    "[]",
		}
		if _, err := repo.CreateAttempt(attempt); err != nil {
			t.Fatalf("CreateAttempt error: %v", err)
		}
	}

	attempts, err := repo.GetUserAttempts(user.ID, "chapter", "ch-get", 10)
	if err != nil {
		t.Fatalf("GetUserAttempts error: %v", err)
	}
	if len(attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(attempts))
	}
	// Order DESC by started_at — latest is last inserted (score 80)
	if attempts[0].Score != 80 {
		t.Fatalf("expected latest score 80, got %d", attempts[0].Score)
	}

	// limit 2
	attempts2, err := repo.GetUserAttempts(user.ID, "chapter", "ch-get", 2)
	if err != nil {
		t.Fatalf("GetUserAttempts(limit=2) error: %v", err)
	}
	if len(attempts2) != 2 {
		t.Fatalf("expected 2 attempts with limit 2, got %d", len(attempts2))
	}
}

func TestGrammarAttemptRepository_GetBestScore(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(5)

	finished := time.Now()
	for _, score := range []int{40, 70, 55} {
		attempt := &TestAttempt{
			UserID:         user.ID,
			ScopeType:      "chapter",
			ScopeID:        "ch-best",
			StartedAt:      time.Now(),
			FinishedAt:     &finished,
			Score:          score,
			Passed:         score >= 50,
			TotalQuestions: 5,
			AnswersJSON:    "[]",
			ResultsJSON:    "[]",
		}
		if _, err := repo.CreateAttempt(attempt); err != nil {
			t.Fatalf("CreateAttempt error: %v", err)
		}
	}

	best, err := repo.GetBestScore(user.ID, "chapter", "ch-best")
	if err != nil {
		t.Fatalf("GetBestScore error: %v", err)
	}
	if best != 70 {
		t.Fatalf("expected best score 70, got %d", best)
	}

	// no attempts for other scope
	bestEmpty, err := repo.GetBestScore(user.ID, "chapter", "ch-none")
	if err != nil {
		t.Fatalf("GetBestScore(empty) error: %v", err)
	}
	if bestEmpty != 0 {
		t.Fatalf("expected 0 for no attempts, got %d", bestEmpty)
	}
}

func TestGrammarAttemptRepository_GetAverageTestScore(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(6)

	finished := time.Now()
	// two chapters, one attempt each
	for _, scopeID := range []string{"ch-a", "ch-b"} {
		attempt := &TestAttempt{
			UserID:         user.ID,
			ScopeType:      "chapter",
			ScopeID:        scopeID,
			StartedAt:      time.Now(),
			FinishedAt:     &finished,
			Score:          60,
			Passed:         true,
			TotalQuestions: 5,
			AnswersJSON:    "[]",
			ResultsJSON:    "[]",
		}
		if _, err := repo.CreateAttempt(attempt); err != nil {
			t.Fatalf("CreateAttempt error: %v", err)
		}
	}

	avg, err := repo.GetAverageTestScore(user.ID)
	if err != nil {
		t.Fatalf("GetAverageTestScore error: %v", err)
	}
	if avg != 60 {
		t.Fatalf("expected average 60, got %d", avg)
	}

	// user with no completed chapter tests
	user2, _ := userRepo.GetOrCreateUser(66)
	avgEmpty, err := repo.GetAverageTestScore(user2.ID)
	if err != nil {
		t.Fatalf("GetAverageTestScore(empty) error: %v", err)
	}
	if avgEmpty != 0 {
		t.Fatalf("expected 0 for no tests, got %d", avgEmpty)
	}
}

func TestParseAnswersJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantLen int
		wantErr bool
	}{
		{"empty string", "", 0, false},
		{"empty object", "{}", 0, false},
		{"object with keys", `{"q1":"a1","q2":2}`, 2, false},
		{"invalid JSON", "{", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAnswersJSON(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseAnswersJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Fatalf("len(got) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestParseResultsJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantLen int
		wantErr bool
	}{
		{"empty string", "", 0, false},
		{"empty array", "[]", 0, false},
		{"array with elements", `[1,true,"x"]`, 3, false},
		{"invalid JSON", "[", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseResultsJSON(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseResultsJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Fatalf("len(got) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestGrammarAttemptRepository_SaveAndGetPlacementTestResult(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(7)

	// no result yet
	got, err := repo.GetPlacementTestResult(user.ID)
	if err != nil {
		t.Fatalf("GetPlacementTestResult error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil when no result, got %+v", got)
	}

	// save first result
	if err := repo.SavePlacementTestResult(user.ID, 50, 10, []string{"s1", "s2"}); err != nil {
		t.Fatalf("SavePlacementTestResult error: %v", err)
	}
	got, err = repo.GetPlacementTestResult(user.ID)
	if err != nil {
		t.Fatalf("GetPlacementTestResult error: %v", err)
	}
	if got == nil || got.Score != 50 || got.TotalQuestions != 10 || len(got.OpenedSections) != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}

	// save higher score — should update
	if err := repo.SavePlacementTestResult(user.ID, 80, 10, []string{"s1", "s2", "s3"}); err != nil {
		t.Fatalf("SavePlacementTestResult(update) error: %v", err)
	}
	got, _ = repo.GetPlacementTestResult(user.ID)
	if got.Score != 80 || len(got.OpenedSections) != 3 {
		t.Fatalf("expected updated score 80 and 3 sections, got %+v", got)
	}

	// save lower score — should not update
	if err := repo.SavePlacementTestResult(user.ID, 30, 10, []string{"s1"}); err != nil {
		t.Fatalf("SavePlacementTestResult(lower) error: %v", err)
	}
	got, _ = repo.GetPlacementTestResult(user.ID)
	if got.Score != 80 {
		t.Fatalf("expected score unchanged 80, got %d", got.Score)
	}
}

// TestGrammarAttemptRepository_GetChapterProgress_NoRow covers GetChapterProgress when no progress row exists (ErrNoRows).
func TestGrammarAttemptRepository_GetChapterProgress_NoRow(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(90)

	progress, err := repo.GetChapterProgress(user.ID, "chapter-no-progress")
	if err != nil {
		t.Fatalf("GetChapterProgress error: %v", err)
	}
	if progress == nil {
		t.Fatal("GetChapterProgress(no row) should return non-nil zero progress")
	}
	if progress.BestScore != 0 || progress.Passed {
		t.Errorf("expected zero progress: BestScore=%d Passed=%v", progress.BestScore, progress.Passed)
	}
}

// TestGrammarAttemptRepository_UpdateProgress_NotPassed covers UpdateProgress when passed=false or score <= 50 (passedAt nil).
func TestGrammarAttemptRepository_UpdateProgress_NotPassed(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(8)
	if err := repo.UpdateProgress(user.ID, "ch-not-passed", 40, false); err != nil {
		t.Fatalf("UpdateProgress(40, false) error: %v", err)
	}
	if err := repo.UpdateProgress(user.ID, "ch-not-passed", 50, true); err != nil {
		t.Fatalf("UpdateProgress(50, true) error: %v", err)
	}
	progress, err := repo.GetChapterProgress(user.ID, "ch-not-passed")
	if err != nil {
		t.Fatalf("GetChapterProgress error: %v", err)
	}
	if progress.BestScore != 50 {
		t.Errorf("expected best score 50, got %d", progress.BestScore)
	}
}

// TestParseTimestampFlex covers all timestamp format branches in parseTimestampFlex.
func TestParseTimestampFlex(t *testing.T) {
	cases := []struct {
		name  string
		input string
		zero  bool
	}{
		{"space format", "2006-01-02 15:04:05", false},
		{"T with timezone", "2006-01-02T15:04:05+03:00", false},
		{"T without timezone", "2006-01-02T15:04:05", false},
		{"T with nanoseconds and timezone", "2006-01-02T15:04:05.123456789+03:00", false},
		{"space with nanoseconds and timezone", "2006-01-02 15:04:05.123456789+03:00", false},
		{"empty string", "", true},
		{"invalid", "not-a-date", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseTimestampFlex(tc.input)
			if tc.zero && !got.IsZero() {
				t.Errorf("expected zero time for %q, got %v", tc.input, got)
			}
			if !tc.zero && got.IsZero() {
				t.Errorf("expected non-zero time for %q, got zero", tc.input)
			}
		})
	}
}

// TestGrammarAttemptRepository_GetPlacementTestResult_InvalidJSON covers error when opened_sections_json is invalid.
func TestGrammarAttemptRepository_GetPlacementTestResult_InvalidJSON(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(9)
	_, err := conn.Exec(`INSERT INTO grammar_placement_test (user_id, score, total_questions, opened_sections_json, completed_at) VALUES ($1, 50, 10, 'invalid-json', CURRENT_TIMESTAMP)`, user.ID)
	if err != nil {
		t.Fatalf("INSERT placement test: %v", err)
	}
	got, err := repo.GetPlacementTestResult(user.ID)
	if err == nil {
		t.Error("GetPlacementTestResult() expected error for invalid JSON")
	}
	if got != nil {
		t.Error("GetPlacementTestResult() expected nil result on error")
	}
}
