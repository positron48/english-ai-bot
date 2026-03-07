package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const selectTrainingCardsQuery = `SELECT id, word_card_id, word_en, word_ru, meaning_en, example_en, pos, display_word 
			  FROM training_cards 
			  WHERE pos IS NULL OR pos = '' OR display_word IS NULL OR display_word = ''
			  ORDER BY id`

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a < b", 1, 2, 1},
		{"a > b", 3, 2, 2},
		{"a == b", 5, 5, 5},
		{"negative a < b", -1, 0, -1},
		{"negative a > b", 0, -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestRunMigration_ConfigLoadFailure(t *testing.T) {
	ctx := context.Background()
	load := func() (*config.Config, error) { return nil, errors.New("injected config error") }
	msg, err := runMigrationWithLoader(ctx, load)
	if err == nil {
		t.Fatal("expected error from runMigrationWithLoader")
	}
	if msg != "Failed to load config" {
		t.Errorf("exitMsg = %q, want %q", msg, "Failed to load config")
	}
}

func TestRunMigrationWithLoader_DBInitFailure(t *testing.T) {
	// Use a loader that returns valid config in memory so we reach DB init, not config load.
	load := func() (*config.Config, error) {
		return &config.Config{
			Logging:  config.LoggingConfig{Level: "info"},
			Database: config.DatabaseConfig{Driver: "postgres", URL: "postgres://nonexistent:5432/nodb?sslmode=disable"},
			AI:       config.AIConfig{URL: "http://localhost:9999", Model: "x", APIKey: "x", Prompt: "p"},
			WebApp:   config.WebAppConfig{JWTSecret: "x"},
		}, nil
	}
	msg, err := runMigrationWithLoader(context.Background(), load)
	if err == nil {
		t.Fatal("expected error from runMigrationWithLoader (db init)")
	}
	if msg != "Failed to initialize database" {
		t.Errorf("exitMsg = %q, want %q", msg, "Failed to initialize database")
	}
}

func TestRunCLI_DBInitFailure_Returns1(t *testing.T) {
	old := configLoader
	configLoader = func() (*config.Config, error) {
		return &config.Config{
			Logging:  config.LoggingConfig{Level: "info"},
			Database: config.DatabaseConfig{Driver: "postgres", URL: "postgres://nonexistent:5432/nodb?sslmode=disable"},
			AI:       config.AIConfig{URL: "http://localhost:9999", Model: "x", APIKey: "x", Prompt: "p"},
			WebApp:   config.WebAppConfig{JWTSecret: "x"},
		}, nil
	}
	defer func() { configLoader = old }()
	code := runCLI()
	if code != 1 {
		t.Errorf("runCLI() = %d, want 1", code)
	}
}

func TestRunMigrationWithLoader_RunReturnsError(t *testing.T) {
	dsn := testutil.GetTestDSN(t)
	// Use a loader that returns valid config in memory so we reach run(), not config load.
	load := func() (*config.Config, error) {
		return &config.Config{
			Logging:  config.LoggingConfig{Level: "info"},
			Database: config.DatabaseConfig{Driver: "postgres", URL: dsn},
			AI:       config.AIConfig{URL: "http://localhost:9999", Model: "x", APIKey: "x", Prompt: "p"},
			WebApp:   config.WebAppConfig{JWTSecret: "x"},
		}, nil
	}
	oldRun := runFuncForTests
	runFuncForTests = func(context.Context, dbConn, wordCardGetter, aiResponder, *zap.Logger) error {
		return errors.New("injected run error")
	}
	defer func() { runFuncForTests = oldRun }()
	msg, err := runMigrationWithLoader(context.Background(), load)
	if err == nil {
		t.Fatal("expected error from runMigrationWithLoader (run)")
	}
	if msg != "Failed to query training cards" {
		t.Errorf("exitMsg = %q, want %q", msg, "Failed to query training cards")
	}
}

func TestRunCLI_ConfigFailure_Returns1(t *testing.T) {
	old := configLoader
	configLoader = func() (*config.Config, error) { return nil, errors.New("injected") }
	defer func() { configLoader = old }()
	code := runCLI()
	if code != 1 {
		t.Errorf("runCLI() = %d, want 1", code)
	}
}

func TestRunMigration_Success_EmptyDB(t *testing.T) {
	dsn := testutil.GetTestDSN(t)
	oldLoader := configLoader
	configLoader = func() (*config.Config, error) {
		return &config.Config{
			Logging:  config.LoggingConfig{Level: "info"},
			Database: config.DatabaseConfig{Driver: "postgres", URL: dsn},
			AI:       config.AIConfig{URL: "http://localhost:9999", Model: "x", APIKey: "test-key", Prompt: "test prompt"},
			WebApp:   config.WebAppConfig{JWTSecret: "test-jwt-secret"},
		}, nil
	}
	defer func() { configLoader = oldLoader }()
	msg, err := runMigration(context.Background())
	if err != nil {
		t.Fatalf("runMigration: %v (msg=%q)", err, msg)
	}
	if msg != "" {
		t.Errorf("runMigration exitMsg = %q, want empty", msg)
	}
}

func TestRunCLI_Success_Returns0(t *testing.T) {
	dsn := testutil.GetTestDSN(t)
	oldLoader := configLoader
	configLoader = func() (*config.Config, error) {
		return &config.Config{
			Logging:  config.LoggingConfig{Level: "info"},
			Database: config.DatabaseConfig{Driver: "postgres", URL: dsn},
			AI:       config.AIConfig{URL: "http://localhost:9999", Model: "x", APIKey: "test-key", Prompt: "test prompt"},
			WebApp:   config.WebAppConfig{JWTSecret: "test-jwt-secret"},
		}, nil
	}
	defer func() { configLoader = oldLoader }()
	code := runCLI()
	if code != 0 {
		t.Errorf("runCLI() = %d, want 0", code)
	}
}

func TestRun_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnError(errors.New("query failed"))

	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, &fakeWordRepo{}, &fakeAI{}, log)
	if err == nil {
		t.Fatal("expected run to return query error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"})
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, &fakeWordRepo{}, &fakeAI{}, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Row with wrong type for one column (e.g. id as string) causes Scan to fail
	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow("not_an_int", int64(1), "word", "слово", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, &fakeWordRepo{}, &fakeAI{}, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_AlreadyMigrated_Skipped(t *testing.T) {
	migrationDelay = 0
	defer func() { migrationDelay = 500 * time.Millisecond }()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", "noun", "spy")
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, &fakeWordRepo{}, &fakeAI{}, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_GetWordCardError(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	wordRepo := &fakeWordRepo{err: errors.New("db error")}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, &fakeAI{}, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_WordCardNotFound(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	wordRepo := &fakeWordRepo{card: nil}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, &fakeAI{}, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_AIResponseError(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	pos := "noun"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", POS: &pos}}
	aiSvc := &fakeAI{err: errors.New("AI error")}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_InvalidJSONFromAI(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	pos := "noun"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", POS: &pos}}
	aiSvc := &fakeAI{response: "not valid json"}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_LLMRejected_RealErrorKeyword(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "xyz", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	pos := "noun"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "xyz", POS: &pos}}
	aiSvc := &fakeAI{response: `{"error": "gibberish word", "pos": "", "display_word": ""}`}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_LLMRejected_UnknownErrorString(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	pos := "noun"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", POS: &pos}}
	aiSvc := &fakeAI{response: `{"error": "some unknown message", "pos": "", "display_word": ""}`}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_NonErrorStringInErrorField_Accepted(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)
	mock.ExpectExec("UPDATE training_cards SET pos = ?, display_word = ? WHERE id = ?").WithArgs("noun", "spy", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	pos := "noun"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", POS: &pos}}
	aiSvc := &fakeAI{response: `{"error": "none", "pos": "noun", "display_word": "spy"}`}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_POSEmpty_UseWordCardFallback(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)
	mock.ExpectExec("UPDATE training_cards SET pos = ?, display_word = ? WHERE id = ?").WithArgs("noun", "spy", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	pos := "noun"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", POS: &pos}}
	aiSvc := &fakeAI{response: `{"error": "", "pos": "", "display_word": "spy"}`}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_POSEmpty_NoFallback_Error(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)

	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", POS: nil}}
	aiSvc := &fakeAI{response: `{"error": "", "pos": "", "display_word": ""}`}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_DisplayWordEmpty_Verb(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)
	mock.ExpectExec("UPDATE training_cards SET pos = ?, display_word = ? WHERE id = ?").WithArgs("verb", "to spy", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	displayEN := "to spy"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", DisplayEN: &displayEN}}
	aiSvc := &fakeAI{response: `{"error": "", "pos": "verb", "display_word": ""}`}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_DisplayWordEmpty_NonVerb_LemmaFromWordEN(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "apple", "яблоко", "fruit", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)
	mock.ExpectExec("UPDATE training_cards SET pos = ?, display_word = ? WHERE id = ?").WithArgs("noun", "apple", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: ""}}
	aiSvc := &fakeAI{response: `{"error": "", "pos": "noun", "display_word": ""}`}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_UpdateError(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)
	mock.ExpectExec("UPDATE training_cards SET pos = ?, display_word = ? WHERE id = ?").WithArgs("noun", "spy", int64(1)).WillReturnError(errors.New("update failed"))

	pos := "noun"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", POS: &pos}}
	aiSvc := &fakeAI{response: `{"error": "", "pos": "noun", "display_word": "spy"}`}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRun_Success_WithExampleAndMarkdownResponse(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "to watch secretly", "He spied on them.", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)
	mock.ExpectExec("UPDATE training_cards SET pos = ?, display_word = ? WHERE id = ?").WithArgs("verb", "to spy", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	displayEN := "to spy"
	pos := "verb"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", DisplayEN: &displayEN, POS: &pos}}
	aiSvc := &fakeAI{response: "```json\n{\"error\": \"\", \"pos\": \"verb\", \"display_word\": \"to spy\"}\n```"}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// fakeWordRepo implements wordCardGetter for tests.
type fakeWordRepo struct {
	card *models.WordCard
	err  error
}

func (f *fakeWordRepo) GetWordCardByID(id int64) (*models.WordCard, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.card, nil
}

// fakeAI implements aiResponder for tests.
type fakeAI struct {
	response string
	err      error
}

func (f *fakeAI) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.response, nil
}

func TestRun_ResponsePreviewLong(t *testing.T) {
	migrationDelay = 0

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "word_ru", "meaning_en", "example_en", "pos", "display_word"}).
		AddRow(int64(1), int64(10), "spy", "шпион", "meaning", "", nil, nil)
	mock.ExpectQuery(selectTrainingCardsQuery).WillReturnRows(rows)
	longResp := `{"error": "", "pos": "noun", "display_word": "spy", "padding": "` + strings.Repeat("x", 350) + `"}`
	aiSvc := &fakeAI{response: longResp}
	mock.ExpectExec("UPDATE training_cards SET pos = ?, display_word = ? WHERE id = ?").WithArgs("noun", "spy", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	pos := "noun"
	wordRepo := &fakeWordRepo{card: &models.WordCard{ID: 10, Word: "spy", POS: &pos}}
	log, _ := zap.NewDevelopment()
	err = run(context.Background(), db, wordRepo, aiSvc, log)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
