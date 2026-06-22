package repository

import (
	"fmt"
	"testing"

	"tgbot-skeleton/internal/models"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// ============================================================
// training_card_repository.go error paths
// ============================================================

func TestTrainingCardRepository_GetTrainingCardByWordCardIDAndSenseIndex_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.GetTrainingCardByWordCardIDAndSenseIndex(1, 0)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_GetTrainingCard_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.GetTrainingCard(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_CreateTrainingCard_CheckExistingError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetTrainingCardByWordCardIDAndSenseIndex returns a real DB error (not ErrNoRows)
	mock.ExpectQuery("SELECT id, word_card_id, word_en").
		WillReturnError(fmt.Errorf("db error"))

	repo := NewTrainingCardRepository(db, zap.NewNop())
	_, err = repo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: 1,
		WordEN:     "test",
		SenseIndex: 0,
		WordRU:     "тест",
		MeaningEN:  "test",
	})
	if err == nil {
		t.Error("expected error when check existing fails")
	}
}

func TestTrainingCardRepository_CreateTrainingCard_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetTrainingCardByWordCardIDAndSenseIndex returns no rows (ErrNoRows → nil, nil)
	mock.ExpectQuery("SELECT id, word_card_id, word_en").
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "transcription", "sense_index",
			"word_ru", "meaning_en", "example_en", "example_ru",
			"distractors_ru", "distractors_en", "hint", "pos", "display_word", "created_at"}))
	// InsertAndReturnID uses QueryRow with RETURNING id, so it's a query
	mock.ExpectQuery("INSERT INTO training_cards").
		WillReturnError(fmt.Errorf("insert error"))

	repo := NewTrainingCardRepository(db, zap.NewNop())
	_, err = repo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: 1,
		WordEN:     "test",
		SenseIndex: 0,
		WordRU:     "тест",
		MeaningEN:  "test",
	})
	if err == nil {
		t.Error("expected error when insert fails")
	}
}

func TestTrainingCardRepository_GetTrainingCardsByWordCardID_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.GetTrainingCardsByWordCardID(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_GetTrainingCardsByWordCardID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "transcription", "sense_index",
		"word_ru", "meaning_en", "example_en", "example_ru",
		"distractors_ru", "distractors_en", "hint", "pos", "display_word", "created_at"}).
		AddRow("not-an-int", 1, "test", "", 0, "тест", "test", "", "", "", "", "", nil, nil, "2024-01-01 00:00:00")
	mock.ExpectQuery("SELECT id, word_card_id, word_en").WillReturnRows(rows)

	repo := NewTrainingCardRepository(db, zap.NewNop())
	_, err = repo.GetTrainingCardsByWordCardID(1)
	if err == nil {
		t.Error("expected scan error")
	}
}

func TestTrainingCardRepository_GetWordCardsWithoutTrainingCards_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.GetWordCardsWithoutTrainingCards(10)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_GetWordCardsWithoutTrainingCards_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word", "definition",
		"pos", "transcription", "definition_ru",
		"examples_json", "verb_forms_json", "display_en",
		"processed_at", "processing_error", "created_at", "updated_at"}).
		AddRow("not-an-int", "word", "def", nil, nil, nil, nil, nil, nil, "", "", "2024-01-01 00:00:00", "2024-01-01 00:00:00")
	mock.ExpectQuery("SELECT wc.id, wc.word").WillReturnRows(rows)

	repo := NewTrainingCardRepository(db, zap.NewNop())
	_, err = repo.GetWordCardsWithoutTrainingCards(10)
	if err == nil {
		t.Error("expected scan error")
	}
}

func TestTrainingCardRepository_HasTrainingCards_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.HasTrainingCards(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_DeleteTrainingCardsByWordEN_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.DeleteTrainingCardsByWordEN("test")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_DeleteTrainingCardsByWordEN_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM training_cards WHERE word_en").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	repo := NewTrainingCardRepository(db, zap.NewNop())
	_, err = repo.DeleteTrainingCardsByWordEN("test")
	if err == nil {
		t.Error("expected error when RowsAffected fails")
	}
}

func TestTrainingCardRepository_DeleteAllTrainingCards_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.DeleteAllTrainingCards()
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_DeleteAllTrainingCards_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM training_cards").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	repo := NewTrainingCardRepository(db, zap.NewNop())
	_, err = repo.DeleteAllTrainingCards()
	if err == nil {
		t.Error("expected error when RowsAffected fails")
	}
}

func TestTrainingCardRepository_GetTrainingCardsByWordEN_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.GetTrainingCardsByWordEN("test")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_GetTrainingCardsByWordEN_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "word_card_id", "word_en", "transcription", "sense_index",
		"word_ru", "meaning_en", "example_en", "example_ru",
		"distractors_ru", "distractors_en", "hint", "pos", "display_word", "created_at"}).
		AddRow("not-an-int", 1, "test", "", 0, "тест", "test", "", "", "", "", "", nil, nil, "2024-01-01 00:00:00")
	mock.ExpectQuery("SELECT id, word_card_id, word_en").WillReturnRows(rows)

	repo := NewTrainingCardRepository(db, zap.NewNop())
	_, err = repo.GetTrainingCardsByWordEN("test")
	if err == nil {
		t.Error("expected scan error")
	}
}

func TestTrainingCardRepository_UpdateTrainingCard_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	err := repo.UpdateTrainingCard(&models.TrainingCard{ID: 1, WordEN: "test", WordRU: "тест", MeaningEN: "test"})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_DeleteTrainingCard_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	err := repo.DeleteTrainingCard(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_DeleteTrainingCard_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM training_cards WHERE id").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	repo := NewTrainingCardRepository(db, zap.NewNop())
	err = repo.DeleteTrainingCard(1)
	if err == nil {
		t.Error("expected error when RowsAffected fails")
	}
}

func TestTrainingCardRepository_ListOrphanedTrainingCards_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.ListOrphanedTrainingCards(10, 0)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTrainingCardRepository_ListOrphanedTrainingCards_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"training_card_id", "word_card_id", "word_en", "transcription", "sense_index",
		"word_ru", "meaning_en", "example_en", "example_ru",
		"pos", "display_word", "created_at", "user_cards_count",
	}).AddRow("not-an-int", 1, "test", "", 0, "тест", "test", "", "", nil, nil, "2024-01-01 00:00:00", 0)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewTrainingCardRepository(db, zap.NewNop())
	_, err = repo.ListOrphanedTrainingCards(10, 0)
	if err == nil {
		t.Error("expected scan error")
	}
}

func TestTrainingCardRepository_CountOrphanedTrainingCards_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTrainingCardRepository(db, zap.NewNop())

	_, err := repo.CountOrphanedTrainingCards()
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// Cover UpdateTrainingCard with DisplayWord set (line 367-369 true branch)
func TestTrainingCardRepository_UpdateTrainingCard_WithDisplayWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	repo := NewTrainingCardRepository(db, logger)

	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "updatedisp", "to update display")
	if err != nil {
		t.Fatalf("insert word card: %v", err)
	}

	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "updatedisp",
		SenseIndex: 0,
		WordRU:     "обновить",
		MeaningEN:  "to update display",
	}
	id, err := repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	displayWord := "to update display"
	pos := "verb"
	card.ID = id
	card.DisplayWord = &displayWord
	card.POS = &pos
	card.MeaningEN = "updated meaning with display word"

	err = repo.UpdateTrainingCard(card)
	if err != nil {
		t.Fatalf("UpdateTrainingCard with DisplayWord: %v", err)
	}

	updated, err := repo.GetTrainingCard(id)
	if err != nil {
		t.Fatalf("GetTrainingCard: %v", err)
	}
	if updated.DisplayWord == nil || *updated.DisplayWord != displayWord {
		t.Errorf("expected DisplayWord %q, got %v", displayWord, updated.DisplayWord)
	}
}

// Cover GetTrainingCardsByWordCardID with pos and display_word set (Valid branches)
func TestTrainingCardRepository_GetTrainingCardsByWordCardID_WithPOSAndDisplayWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	repo := NewTrainingCardRepository(db, logger)

	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "poscard", "def")
	if err != nil {
		t.Fatalf("insert word card: %v", err)
	}

	pos := "noun"
	displayWord := "pos card"
	card := &models.TrainingCard{
		WordCardID:  1,
		WordEN:      "poscard",
		SenseIndex:  0,
		WordRU:      "карточка",
		MeaningEN:   "def",
		POS:         &pos,
		DisplayWord: &displayWord,
	}
	_, err = repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	cards, err := repo.GetTrainingCardsByWordCardID(1)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].POS == nil || *cards[0].POS != pos {
		t.Errorf("expected POS %q, got %v", pos, cards[0].POS)
	}
	if cards[0].DisplayWord == nil || *cards[0].DisplayWord != displayWord {
		t.Errorf("expected DisplayWord %q, got %v", displayWord, cards[0].DisplayWord)
	}
}

// Cover GetTrainingCardsByWordEN with pos and display_word set (Valid branches)
func TestTrainingCardRepository_GetTrainingCardsByWordEN_WithPOSAndDisplayWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	repo := NewTrainingCardRepository(db, logger)

	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "poswordEN", "def")
	if err != nil {
		t.Fatalf("insert word card: %v", err)
	}

	pos := "verb"
	displayWord := "to pos word"
	card := &models.TrainingCard{
		WordCardID:  1,
		WordEN:      "poswordEN",
		SenseIndex:  0,
		WordRU:      "слово",
		MeaningEN:   "def",
		POS:         &pos,
		DisplayWord: &displayWord,
	}
	_, err = repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	cards, err := repo.GetTrainingCardsByWordEN("poswordEN")
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordEN: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].POS == nil || *cards[0].POS != pos {
		t.Errorf("expected POS %q, got %v", pos, cards[0].POS)
	}
	if cards[0].DisplayWord == nil || *cards[0].DisplayWord != displayWord {
		t.Errorf("expected DisplayWord %q, got %v", displayWord, cards[0].DisplayWord)
	}
}

// Cover GetWordCardsWithoutTrainingCards with all optional fields set
// (examplesJSON, verbFormsJSON, displayEN, processedAt, processingError branches)
func TestTrainingCardRepository_GetWordCardsWithoutTrainingCards_AllOptionalFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	repo := NewTrainingCardRepository(db, logger)

	examplesJSON := `["example1"]`
	verbFormsJSON := `{"past": "ran"}`
	displayEN := "all fields word"
	processingError := "some error"

	// Insert word card with all optional fields and a processing_error (but no processed_at so it appears in results)
	_, err := db.Exec(`INSERT INTO word_cards (word, definition, pos, transcription, definition_ru, examples_json, verb_forms_json, display_en, processing_error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		"allfieldword", "def", "noun", "ˈɔːl", "все поля", examplesJSON, verbFormsJSON, displayEN, processingError)
	if err != nil {
		t.Fatalf("insert word card: %v", err)
	}

	cards, err := repo.GetWordCardsWithoutTrainingCards(10)
	if err != nil {
		t.Fatalf("GetWordCardsWithoutTrainingCards: %v", err)
	}

	var found *models.WordCard
	for _, c := range cards {
		if c.Word == "allfieldword" {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("expected to find allfieldword in results")
	}
	if found.ExamplesJSON == nil || *found.ExamplesJSON != examplesJSON {
		t.Errorf("expected ExamplesJSON %q, got %v", examplesJSON, found.ExamplesJSON)
	}
	if found.VerbFormsJSON == nil || *found.VerbFormsJSON != verbFormsJSON {
		t.Errorf("expected VerbFormsJSON %q, got %v", verbFormsJSON, found.VerbFormsJSON)
	}
	if found.DisplayEN == nil || *found.DisplayEN != displayEN {
		t.Errorf("expected DisplayEN %q, got %v", displayEN, found.DisplayEN)
	}
	if found.ProcessingError == nil || *found.ProcessingError != processingError {
		t.Errorf("expected ProcessingError %q, got %v", processingError, found.ProcessingError)
	}
}

// Cover GetWordCardsWithoutTrainingCards with processedAt set (processedAtStr != "" branch)
func TestTrainingCardRepository_GetWordCardsWithoutTrainingCards_WithProcessedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a word card row with processedAt set (non-empty string)
	rows := sqlmock.NewRows([]string{"id", "word", "definition",
		"pos", "transcription", "definition_ru",
		"examples_json", "verb_forms_json", "display_en",
		"course_code",
		"processed_at", "processing_error", "created_at", "updated_at"}).
		AddRow(1, "processedword", "def", nil, nil, nil, nil, nil, nil,
			"", "2024-01-01 10:00:00", "", "2024-01-01 00:00:00", "2024-01-01 00:00:00")
	mock.ExpectQuery("SELECT wc.id, wc.word").WillReturnRows(rows)

	repo := NewTrainingCardRepository(db, zap.NewNop())
	cards, err := repo.GetWordCardsWithoutTrainingCards(10)
	if err != nil {
		t.Fatalf("GetWordCardsWithoutTrainingCards: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].ProcessedAt == nil {
		t.Error("expected ProcessedAt to be set")
	}
}
