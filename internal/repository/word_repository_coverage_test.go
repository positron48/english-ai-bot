package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const wordInvalidDSN = "postgres://x:x@invalid.invalid:1/db?connect_timeout=1"

// invalidWordDB returns a *sql.DB with an invalid DSN so any operation returns an error.
func invalidWordDB(t *testing.T) *sql.DB {
	t.Helper()
	// Ensure postgres_compat driver is registered by initializing the shared test DB first.
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", wordInvalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestWordRepository_GetWordCardByID_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	_, err := repo.GetWordCardByID(1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_GetWordCardByLemma_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	_, err := repo.GetWordCardByLemma("word")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_SaveWordCard_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	err := repo.SaveWordCard("word", "definition")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_UpsertWordCardLemma_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	card := &models.WordCard{Word: "word", Definition: "def"}
	_, err := repo.UpsertWordCardLemma(card)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_GetWordFormMapping_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	_, err := repo.GetWordFormMapping("form")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_UpsertWordFormMapping_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	err := repo.UpsertWordFormMapping("form", 1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_AddWordRequestHistoryWithCard_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	err := repo.AddWordRequestHistoryWithCard(1, "word", nil, nil)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_GetUserIDsByWord_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	_, err := repo.GetUserIDsByWord("word")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_ListPronunciationCandidates_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	_, err := repo.ListPronunciationCandidates(10)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_MarkWordCardProcessedError_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	err := repo.MarkWordCardProcessedError(1, "error")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_ResetWordCardProcessed_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	err := repo.ResetWordCardProcessed(1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_UpdateWordCardDefinition_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	err := repo.UpdateWordCardDefinition(1, "new def")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_UpdateWordCard_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	card := &models.WordCard{ID: 1, Word: "word", Definition: "def"}
	err := repo.UpdateWordCard(card)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_ListWordCardsAdmin_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	_, err := repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "", "desc")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_CountWordCardsAdmin_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	_, err := repo.CountWordCardsAdmin(nil, false, nil, "", "")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_GetWordCardRequestingUsers_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	_, err := repo.GetWordCardRequestingUsers(1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestWordRepository_DeleteWordCard_DBError(t *testing.T) {
	db := invalidWordDB(t)
	repo := NewWordRepository(db, zap.NewNop())

	err := repo.DeleteWordCard(1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

// TestWordRepository_GetWordCardByID_WithAllOptionalFields covers the branches where
// processedAtStr != "" and processingErrorStr != "" are both true.
func TestWordRepository_GetWordCardByID_WithAllOptionalFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	pos := "noun"
	transcription := "/test/"
	definitionRU := "тест"
	examplesJSON := `[{"en":"test","ru":"тест"}]`
	verbFormsJSON := `{"v1":"test"}`
	displayEN := "test"

	card := &models.WordCard{
		Word:          "fulloptword",
		Definition:    "full optional fields",
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU:  &definitionRU,
		ExamplesJSON:  &examplesJSON,
		VerbFormsJSON: &verbFormsJSON,
		DisplayEN:     &displayEN,
	}
	id, err := repo.UpsertWordCardLemma(card)
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	// Mark with error so processedAtStr and processingErrorStr are set
	if err := repo.MarkWordCardProcessedError(id, "test error text"); err != nil {
		t.Fatalf("MarkWordCardProcessedError: %v", err)
	}

	retrieved, err := repo.GetWordCardByID(id)
	if err != nil {
		t.Fatalf("GetWordCardByID() error = %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected non-nil card")
	}
	if retrieved.ProcessedAt == nil {
		t.Error("expected ProcessedAt to be set")
	}
	if retrieved.ProcessingError == nil || *retrieved.ProcessingError != "test error text" {
		t.Errorf("expected ProcessingError='test error text', got %v", retrieved.ProcessingError)
	}
	if retrieved.POS == nil || *retrieved.POS != "noun" {
		t.Errorf("expected POS='noun', got %v", retrieved.POS)
	}
	if retrieved.Transcription == nil || *retrieved.Transcription != "/test/" {
		t.Errorf("expected Transcription='/test/', got %v", retrieved.Transcription)
	}
	if retrieved.DefinitionRU == nil || *retrieved.DefinitionRU != "тест" {
		t.Errorf("expected DefinitionRU='тест', got %v", retrieved.DefinitionRU)
	}
	if retrieved.ExamplesJSON == nil {
		t.Error("expected ExamplesJSON to be set")
	}
	if retrieved.VerbFormsJSON == nil {
		t.Error("expected VerbFormsJSON to be set")
	}
	if retrieved.DisplayEN == nil || *retrieved.DisplayEN != "test" {
		t.Errorf("expected DisplayEN='test', got %v", retrieved.DisplayEN)
	}
}

// TestWordRepository_GetWordCardByLemma_WithProcessedAt covers the branch where processedAtStr != "".
func TestWordRepository_GetWordCardByLemma_WithProcessedAt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	if err := repo.SaveWordCard("processedlemma", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, err := repo.GetWordCardByLemma("processedlemma")
	if err != nil {
		t.Fatalf("GetWordCardByLemma: %v", err)
	}
	if err := repo.MarkWordCardProcessedError(card.ID, "lemma error"); err != nil {
		t.Fatalf("MarkWordCardProcessedError: %v", err)
	}

	retrieved, err := repo.GetWordCardByLemma("processedlemma")
	if err != nil {
		t.Fatalf("GetWordCardByLemma() error = %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected non-nil card")
	}
	if retrieved.ProcessedAt == nil {
		t.Error("expected ProcessedAt to be set after MarkWordCardProcessedError")
	}
	if retrieved.ProcessingError == nil || *retrieved.ProcessingError != "lemma error" {
		t.Errorf("expected ProcessingError='lemma error', got %v", retrieved.ProcessingError)
	}
}

// TestWordRepository_UpsertWordCardLemma_ConflictReturnsExistingID covers the branch where
// LastInsertId returns 0 (ON CONFLICT UPDATE case) and we fall back to GetWordCardByLemma.
func TestWordRepository_UpsertWordCardLemma_ConflictReturnsExistingID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	// First insert
	card := &models.WordCard{Word: "conflictlemma", Definition: "first def"}
	id1, err := repo.UpsertWordCardLemma(card)
	if err != nil {
		t.Fatalf("UpsertWordCardLemma first: %v", err)
	}

	// Second insert (conflict - triggers UPDATE, LastInsertId may return 0)
	card2 := &models.WordCard{Word: "conflictlemma", Definition: "second def"}
	id2, err := repo.UpsertWordCardLemma(card2)
	if err != nil {
		t.Fatalf("UpsertWordCardLemma second: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same ID on conflict, got %d vs %d", id1, id2)
	}
}

// TestWordRepository_ListPronunciationCandidates_LimitReached covers the break branch
// when len(candidates) >= limit.
func TestWordRepository_ListPronunciationCandidates_LimitReached(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	// Insert more words than the limit
	for i := 0; i < 5; i++ {
		word := "limitword" + string(rune('a'+i))
		_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", word, "def")
	}

	// Request only 2 candidates; the query fetches limit*3=6 rows but we stop at 2
	candidates, err := repo.ListPronunciationCandidates(2)
	if err != nil {
		t.Fatalf("ListPronunciationCandidates() error = %v", err)
	}
	if len(candidates) > 2 {
		t.Errorf("expected at most 2 candidates, got %d", len(candidates))
	}
}

// TestWordRepository_GetUserIDsByWord_WithWordCard covers the branch where wordCard != nil.
func TestWordRepository_GetUserIDsByWord_WithWordCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)
	userRepo := NewUserRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(99010)
	if err := repo.SaveWordCard("withcardword", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, _ := repo.GetWordCardByLemma("withcardword")
	word := "withcardword"
	_ = repo.AddWordRequestHistoryWithCard(user.ID, "withcardword", &card.ID, &word)

	userIDs, err := repo.GetUserIDsByWord("withcardword")
	if err != nil {
		t.Fatalf("GetUserIDsByWord() error = %v", err)
	}
	if len(userIDs) == 0 {
		t.Error("expected at least 1 user ID")
	}
}

// TestWordRepository_ListWordCardsAdmin_WithFilterUserID covers the filterUserID branch.
func TestWordRepository_ListWordCardsAdmin_WithFilterUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)
	userRepo := NewUserRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(99020)
	if err := repo.SaveWordCard("filteruserword", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, _ := repo.GetWordCardByLemma("filteruserword")
	word := "filteruserword"
	_ = repo.AddWordRequestHistoryWithCard(user.ID, "filteruserword", &card.ID, &word)

	userID := user.ID
	cards, err := repo.ListWordCardsAdmin(&userID, false, nil, "", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(filterUserID) error = %v", err)
	}
	_ = cards
}

// TestWordRepository_DeleteWordCard_WithUserCards_Success covers the branch in DeleteWordCard
// where deleting user_cards succeeds.
func TestWordRepository_DeleteWordCard_WithUserCards_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	tcRepo := NewTrainingCardRepository(db, logger)
	ucRepo := NewUserCardRepository(db, logger)

	if err := repo.SaveWordCard("deleteucword", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, _ := repo.GetWordCardByLemma("deleteucword")
	user, _ := userRepo.GetOrCreateUser(99030)
	tc := &models.TrainingCard{WordCardID: card.ID, WordEN: "deleteucword", SenseIndex: 0, WordRU: "удалить", MeaningEN: "delete"}
	tcID, _ := tcRepo.CreateTrainingCard(tc)

	uc := &models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateNew, EF: 2.5}
	_, _ = ucRepo.CreateUserCard(uc)

	err := repo.DeleteWordCard(card.ID)
	if err != nil {
		t.Fatalf("DeleteWordCard() error = %v", err)
	}
	deleted, _ := repo.GetWordCardByLemma("deleteucword")
	if deleted != nil {
		t.Error("expected word card to be deleted")
	}
}

// TestWordRepository_ListWordCardsAdmin_WithTTSFields covers the scan path where
// ttsStateStr, ttsErrorStr, and ttsAudioRelPath are non-empty.
func TestWordRepository_ListWordCardsAdmin_WithTTSFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	if err := repo.SaveWordCard("ttsadminword", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	// Insert TTS row with all fields populated
	_, err := db.Exec(`INSERT INTO tts_generation_status (word, state, audio_rel_path, last_error_message, attempt_count, max_attempts)
		VALUES ('ttsadminword', 'ready', 'ab/cd/ttsadminword.mp3', 'some error', 1, 3)
		ON CONFLICT (course_code, word) DO UPDATE SET state = 'ready', audio_rel_path = 'ab/cd/ttsadminword.mp3', last_error_message = 'some error'`)
	if err != nil {
		t.Fatalf("insert tts status: %v", err)
	}

	cards, err := repo.ListWordCardsAdmin(nil, false, nil, "ttsadmin", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	var found bool
	for _, c := range cards {
		if c.Word == "ttsadminword" {
			found = true
			if c.TTSState == nil {
				t.Error("expected TTSState to be set")
			}
			if c.TTSAudioURL == nil {
				t.Error("expected TTSAudioURL to be set")
			}
		}
	}
	if !found {
		t.Error("expected ttsadminword in list")
	}
}

// TestWordRepository_CountWordCardsAdmin_HasAudioFilter covers the hasAudio filter branches.
func TestWordRepository_CountWordCardsAdmin_HasAudioFilter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	if err := repo.SaveWordCard("audiocount1", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	if err := repo.SaveWordCard("audiocount2", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	// Insert TTS row with audio for one word
	_, err := db.Exec(`INSERT INTO tts_generation_status (word, state, audio_rel_path, attempt_count, max_attempts)
		VALUES ('audiocount1', 'ready', 'ab/cd/audiocount1.mp3', 1, 3)
		ON CONFLICT (course_code, word) DO UPDATE SET state = 'ready', audio_rel_path = 'ab/cd/audiocount1.mp3'`)
	if err != nil {
		t.Fatalf("insert tts status: %v", err)
	}

	hasAudioTrue := true
	countWith, err := repo.CountWordCardsAdmin(nil, false, &hasAudioTrue, "", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin(hasAudio=true) error = %v", err)
	}
	if countWith < 1 {
		t.Errorf("expected at least 1 word with audio, got %d", countWith)
	}

	hasAudioFalse := false
	countWithout, err := repo.CountWordCardsAdmin(nil, false, &hasAudioFalse, "", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin(hasAudio=false) error = %v", err)
	}
	if countWithout < 1 {
		t.Errorf("expected at least 1 word without audio, got %d", countWithout)
	}
}

// TestWordRepository_ListWordCardsAdmin_WithAllOptionalFields covers the optional field branches
// in ListWordCardsAdmin (pos, transcription, definitionRU, examplesJSON, verbFormsJSON, displayEN).
func TestWordRepository_ListWordCardsAdmin_WithAllOptionalFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	pos := "noun"
	transcription := "/ˈɔːl.ɒp.ʃ.nəl/"
	definitionRU := "все поля"
	examplesJSON := `[{"en":"all optional","ru":"все необязательные"}]`
	verbFormsJSON := `{"v1":"opt"}`
	displayEN := "all optional"

	card := &models.WordCard{
		Word:          "alloptfields",
		Definition:    "all optional fields",
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU:  &definitionRU,
		ExamplesJSON:  &examplesJSON,
		VerbFormsJSON: &verbFormsJSON,
		DisplayEN:     &displayEN,
	}
	_, err := repo.UpsertWordCardLemma(card)
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	cards, err := repo.ListWordCardsAdmin(nil, false, nil, "alloptfields", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	var found bool
	for _, c := range cards {
		if c.Word == "alloptfields" {
			found = true
			if c.POS == nil || *c.POS != "noun" {
				t.Errorf("expected POS='noun', got %v", c.POS)
			}
			if c.Transcription == nil {
				t.Error("expected Transcription to be set")
			}
			if c.DefinitionRU == nil {
				t.Error("expected DefinitionRU to be set")
			}
			if c.ExamplesJSON == nil {
				t.Error("expected ExamplesJSON to be set")
			}
			if c.VerbFormsJSON == nil {
				t.Error("expected VerbFormsJSON to be set")
			}
			if c.DisplayEN == nil {
				t.Error("expected DisplayEN to be set")
			}
		}
	}
	if !found {
		t.Error("expected alloptfields in list")
	}
}

// TestWordRepository_ListPronunciationCandidates_EmptyCandidate covers the empty candidate skip branch.
func TestWordRepository_ListPronunciationCandidates_EmptyCandidate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	// Insert a word with spaces only (will be trimmed to empty)
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "   ", "def with spaces")
	// Insert a normal word
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "normalcandidate", "def")

	candidates, err := repo.ListPronunciationCandidates(10)
	if err != nil {
		t.Fatalf("ListPronunciationCandidates() error = %v", err)
	}
	// normalcandidate should be in candidates; the spaces-only word should be skipped
	for _, c := range candidates {
		if c == "   " {
			t.Error("expected empty/whitespace candidate to be skipped")
		}
	}
}

// TestWordRepository_ListPronunciationCandidates_Deduplication covers the deduplication branch.
func TestWordRepository_ListPronunciationCandidates_Deduplication(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	// Insert words with same lowercase form (case-insensitive dedup)
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "DedupWord", "def1")
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "dedupword", "def2")

	candidates, err := repo.ListPronunciationCandidates(10)
	if err != nil {
		t.Fatalf("ListPronunciationCandidates() error = %v", err)
	}
	// Count how many times "dedupword" appears (case-insensitive)
	count := 0
	for _, c := range candidates {
		if strings.EqualFold(c, "dedupword") {
			count++
		}
	}
	if count > 1 {
		t.Errorf("expected dedupword to appear at most once (deduplication), got %d", count)
	}
}

// TestWordRepository_UpsertWordCardLemma_GetByLemmaError covers the branch where
// GetWordCardByLemma returns an error (fallback path when id == 0).
func TestWordRepository_UpsertWordCardLemma_GetByLemmaError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Exec for INSERT: succeeds but returns id=0 (ON CONFLICT UPDATE case)
	mock.ExpectExec("INSERT INTO word_cards").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// GetWordCardByLemma: fails
	mock.ExpectQuery("SELECT .+ FROM word_cards").
		WillReturnError(fmt.Errorf("get lemma error"))

	repo := NewWordRepository(db, zap.NewNop())
	card := &models.WordCard{Word: "errorlemma", Definition: "def"}
	_, err = repo.UpsertWordCardLemma(card)
	if err == nil {
		t.Fatal("expected error when GetWordCardByLemma fails")
	}
}

// TestWordRepository_UpsertWordCardLemma_GetByLemmaNil covers the branch where
// GetWordCardByLemma returns nil (word not found after upsert).
func TestWordRepository_UpsertWordCardLemma_GetByLemmaNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Exec for INSERT: succeeds but returns id=0
	mock.ExpectExec("INSERT INTO word_cards").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// GetWordCardByLemma: returns no rows (nil)
	mock.ExpectQuery("SELECT .+ FROM word_cards").
		WillReturnError(sql.ErrNoRows)

	repo := NewWordRepository(db, zap.NewNop())
	card := &models.WordCard{Word: "nillemma", Definition: "def"}
	_, err = repo.UpsertWordCardLemma(card)
	if err == nil {
		t.Fatal("expected error when word card not found after upsert")
	}
}

// TestWordRepository_GetUserIDsByWord_QueryError covers the query error path (wordCard != nil case).
func TestWordRepository_GetUserIDsByWord_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordCardByLemma: returns a word card
	wordRows := sqlmock.NewRows([]string{"id", "word", "definition", "pos", "transcription",
		"definition_ru", "examples_json", "verb_forms_json", "display_en",
		"processed_at", "processing_error", "created_at", "updated_at"}).
		AddRow(1, "testword", "def", nil, nil, nil, nil, nil, nil, "", "", "2024-01-01 00:00:00", "2024-01-01 00:00:00")
	mock.ExpectQuery("SELECT .+ FROM word_cards").WillReturnRows(wordRows)
	// Query for user IDs: fails
	mock.ExpectQuery("SELECT DISTINCT user_id").WillReturnError(fmt.Errorf("query error"))

	repo := NewWordRepository(db, zap.NewNop())
	_, err = repo.GetUserIDsByWord("testword")
	if err == nil {
		t.Fatal("expected error when query fails")
	}
}

// TestWordRepository_GetUserIDsByWord_ScanError covers the scan error path.
func TestWordRepository_GetUserIDsByWord_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordCardByLemma: returns no rows (nil)
	mock.ExpectQuery("SELECT .+ FROM word_cards").WillReturnError(sql.ErrNoRows)
	// Query for user IDs: returns row with wrong type
	rows := sqlmock.NewRows([]string{"user_id"}).AddRow("not-an-int")
	mock.ExpectQuery("SELECT DISTINCT user_id").WillReturnRows(rows)

	repo := NewWordRepository(db, zap.NewNop())
	_, err = repo.GetUserIDsByWord("unknownword")
	if err == nil {
		t.Fatal("expected scan error")
	}
}

// TestWordRepository_ListPronunciationCandidates_ScanError covers the scan error path.
func TestWordRepository_ListPronunciationCandidates_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"candidate"}).AddRow(nil) // nil will fail to scan into string
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewWordRepository(db, zap.NewNop())
	_, err = repo.ListPronunciationCandidates(10)
	if err == nil {
		t.Fatal("expected scan error")
	}
}

// TestWordRepository_ListWordCardsAdmin_ScanError covers the scan error path.
func TestWordRepository_ListWordCardsAdmin_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row that will fail to scan (wrong type for ID)
	rows := sqlmock.NewRows([]string{"id", "word", "definition",
		"pos", "transcription", "definition_ru", "examples_json", "verb_forms_json", "display_en",
		"processed_at", "processing_error", "tts_state", "tts_error", "tts_audio_rel_path",
		"created_at", "updated_at", "has_training_cards"}).
		AddRow("not-an-int", "word", "def", "", "", "", "", "", "", "", "", "", "", "", "2024-01-01", "2024-01-01", 0)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewWordRepository(db, zap.NewNop())
	_, err = repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "", "desc")
	if err == nil {
		t.Fatal("expected scan error")
	}
}

// TestWordRepository_ListWordCardsAdmin_DefaultSortOrder covers the default sort order branch.
func TestWordRepository_ListWordCardsAdmin_DefaultSortOrder(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	if err := repo.SaveWordCard("defaultsortword", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}

	// Use empty sortOrder to trigger the default branch
	cards, err := repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "", "")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin(sortOrder='') error = %v", err)
	}
	_ = cards
}

// TestWordRepository_DeleteWordCard_UserCardsDeleteError covers the branch where
// deleting user_cards fails (logs warning and continues).
func TestWordRepository_DeleteWordCard_UserCardsDeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Get word: returns "testword"
	mock.ExpectQuery("SELECT word FROM word_cards").
		WillReturnRows(sqlmock.NewRows([]string{"word"}).AddRow("testword"))
	// Delete user_cards: fails (but continues)
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnError(fmt.Errorf("user cards delete error"))
	// Delete training_cards: succeeds
	mock.ExpectExec("DELETE FROM training_cards").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Delete word_card: succeeds
	mock.ExpectExec("DELETE FROM word_cards").
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewWordRepository(db, zap.NewNop())
	err = repo.DeleteWordCard(1)
	if err != nil {
		t.Fatalf("DeleteWordCard() unexpected error = %v", err)
	}
}

// TestWordRepository_DeleteWordCard_TrainingCardsDeleteError covers the branch where
// deleting training_cards fails (logs warning and continues).
func TestWordRepository_DeleteWordCard_TrainingCardsDeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Get word: returns "testword"
	mock.ExpectQuery("SELECT word FROM word_cards").
		WillReturnRows(sqlmock.NewRows([]string{"word"}).AddRow("testword"))
	// Delete user_cards: succeeds
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	// Delete training_cards: fails (but continues)
	mock.ExpectExec("DELETE FROM training_cards").
		WillReturnError(fmt.Errorf("training cards delete error"))
	// Delete word_card: succeeds
	mock.ExpectExec("DELETE FROM word_cards").
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewWordRepository(db, zap.NewNop())
	err = repo.DeleteWordCard(1)
	if err != nil {
		t.Fatalf("DeleteWordCard() unexpected error = %v", err)
	}
}

// TestWordRepository_DeleteWordCard_FinalDeleteError covers the final DELETE error path.
func TestWordRepository_DeleteWordCard_FinalDeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT word FROM word_cards").
		WillReturnRows(sqlmock.NewRows([]string{"word"}).AddRow("testword"))
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectExec("DELETE FROM training_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	// Final DELETE fails
	mock.ExpectExec("DELETE FROM word_cards").
		WillReturnError(fmt.Errorf("final delete error"))

	repo := NewWordRepository(db, zap.NewNop())
	err = repo.DeleteWordCard(1)
	if err == nil {
		t.Fatal("expected error from final DELETE")
	}
}

// TestWordRepository_DeleteWordCard_RowsAffectedError covers the RowsAffected error path.
func TestWordRepository_DeleteWordCard_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT word FROM word_cards").
		WillReturnRows(sqlmock.NewRows([]string{"word"}).AddRow("testword"))
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectExec("DELETE FROM training_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	// Final DELETE: succeeds but RowsAffected fails
	mock.ExpectExec("DELETE FROM word_cards").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	repo := NewWordRepository(db, zap.NewNop())
	err = repo.DeleteWordCard(1)
	if err == nil {
		t.Fatal("expected error from RowsAffected")
	}
}

// TestWordRepository_DeleteWordCard_RowsAffectedZero covers the rowsAffected == 0 path.
func TestWordRepository_DeleteWordCard_RowsAffectedZero(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT word FROM word_cards").
		WillReturnRows(sqlmock.NewRows([]string{"word"}).AddRow("testword"))
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectExec("DELETE FROM training_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	// Final DELETE: 0 rows affected
	mock.ExpectExec("DELETE FROM word_cards").
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo := NewWordRepository(db, zap.NewNop())
	err = repo.DeleteWordCard(1)
	if err == nil {
		t.Fatal("expected 'word card not found' error when 0 rows affected")
	}
}

// TestWordRepository_ListWordCardsAdmin_GetUserIDsByWordError covers the branch where
// GetUserIDsByWord fails (logs warning and continues).
func TestWordRepository_ListWordCardsAdmin_GetUserIDsByWordError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Main query returns one row
	rows := sqlmock.NewRows([]string{"id", "word", "definition",
		"pos", "noun_gender", "opposite_gender_word", "transcription", "definition_ru", "examples_json", "verb_forms_json", "display_en",
		"processed_at", "processing_error", "tts_state", "tts_error", "tts_audio_rel_path",
		"created_at", "updated_at", "has_training_cards"}).
		AddRow(1, "testword", "def", "", "", "", "", "", "", "", "", "", "", "", "", "", "2024-01-01 00:00:00", "2024-01-01 00:00:00", 0)
	mock.ExpectQuery("SELECT wc.id").WillReturnRows(rows)
	// GetUserIDsByWord (GetWordCardByLemma) fails
	mock.ExpectQuery("SELECT .+ FROM word_cards").WillReturnError(fmt.Errorf("get word error"))

	repo := NewWordRepository(db, zap.NewNop())
	items, err := repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() unexpected error = %v", err)
	}
	// Should still return the item (error is logged as warning, not returned)
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

// TestWordRepository_ListWordCardsAdmin_WithProcessedAt covers the branch where processedAtStr != "".
func TestWordRepository_ListWordCardsAdmin_WithProcessedAt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db, logger)

	if err := repo.SaveWordCard("processedadmin", "def"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	card, _ := repo.GetWordCardByLemma("processedadmin")
	if err := repo.MarkWordCardProcessedError(card.ID, "admin error"); err != nil {
		t.Fatalf("MarkWordCardProcessedError: %v", err)
	}

	cards, err := repo.ListWordCardsAdmin(nil, true, nil, "processedadmin", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	var found bool
	for _, c := range cards {
		if c.Word == "processedadmin" {
			found = true
			if c.ProcessedAt == nil {
				t.Error("expected ProcessedAt to be set")
			}
			if c.ProcessingError == nil {
				t.Error("expected ProcessingError to be set")
			}
		}
	}
	if !found {
		t.Error("expected processedadmin in list")
	}
}
