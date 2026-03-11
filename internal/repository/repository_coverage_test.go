package repository

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// newBrokenDB returns a *sql.DB that is already closed (for testing error paths).
func newBrokenDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := testutil.GetTestDSN(t)
	conn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("postgres_compat open: %v", err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		t.Skipf("ping: %v", err)
	}
	conn.Close()
	return conn
}

// ============================================================
// word_set_repository.go error paths
// ============================================================

func TestWordSetRepository_CreateWordSet_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetRepository(db, logger)

	_, err := repo.CreateWordSet(&models.WordSet{Title: "x"})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetRepository_GetWordSet_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetRepository(db, logger)

	_, err := repo.GetWordSet(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetRepository_ListWordSets_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetRepository(db, logger)

	_, err := repo.ListWordSets(nil, 10, 0)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetRepository_UpdateWordSet_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetRepository(db, logger)

	err := repo.UpdateWordSet(&models.WordSet{ID: 1, Title: "x"})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetRepository_DeleteWordSet_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetRepository(db, logger)

	err := repo.DeleteWordSet(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetRepository_GetWordSetProgress_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetRepository(db, logger)

	_, err := repo.GetWordSetProgress(1, 1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// TestWordSetRepository_GetWordSetProgress_CountTotalError is skipped because dropping
// word_set_items would break the shared test DB for other tests.
// The error path is covered by TestWordSetRepository_GetWordSetProgress_DBError instead.

func TestWordSetRepository_GetWordSetWords_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetRepository(db, logger)

	_, err := repo.GetWordSetWords(1, 1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetRepository_SetWordSetItems_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetRepository(db, logger)

	err := repo.SetWordSetItems(1, []int64{1, 2})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// Cover ListWordSets with description and preferredPOS set (Valid branches in loop)
func TestWordSetRepository_ListWordSets_WithDescAndPOS(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetRepository(db, logger)

	desc := "list desc"
	pos := "verb"
	_, err := repo.CreateWordSet(&models.WordSet{
		Title:        "List With Desc",
		Description:  &desc,
		PreferredPOS: &pos,
		IsPublished:  true,
		SortOrder:    1,
	})
	if err != nil {
		t.Fatalf("CreateWordSet: %v", err)
	}

	sets, err := repo.ListWordSets(nil, 10, 0, true)
	if err != nil {
		t.Fatalf("ListWordSets: %v", err)
	}
	found := false
	for _, s := range sets {
		if s.Title == "List With Desc" && s.Description != nil && s.PreferredPOS != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected word set with description and preferredPOS in list")
	}
}

// Cover GetWordSetProgress with words in vocab (wordsInVocab > 0)
func TestWordSetRepository_GetWordSetProgress_WithWordsInVocab(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	userRepo := NewUserRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(9010)
	setID, _ := repo.CreateWordSet(&models.WordSet{Title: "Vocab Test", IsPublished: true})

	wordCardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "vocabword", Definition: "def"})
	_ = repo.SetWordSetItems(setID, []int64{wordCardID})

	// Create training card and user card so word is "in_vocab"
	var tcID int64
	_ = db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES ($1, 'vocabword', 0, 'слово', 'word') RETURNING id`, wordCardID).Scan(&tcID)
	_, _ = db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'en_ru', 'review')`, user.ID, tcID)

	progress, err := repo.GetWordSetProgress(setID, user.ID)
	if err != nil {
		t.Fatalf("GetWordSetProgress: %v", err)
	}
	if progress.WordsInVocab != 1 {
		t.Errorf("expected 1 word in vocab, got %d", progress.WordsInVocab)
	}
}

// Cover GetWordSetWords with all optional fields set (transcription, wordRU, etc.)
func TestWordSetRepository_GetWordSetWords_WithAllFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	userRepo := NewUserRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(9011)
	setID, _ := repo.CreateWordSet(&models.WordSet{Title: "Full Fields Test", IsPublished: true})
	wordCardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "fullword", Definition: "def"})
	_ = repo.SetWordSetItems(setID, []int64{wordCardID})

	// Create training card with all optional fields
	transcription := "/fʊlwɜːd/"
	wordRU := "полное слово"
	meaningEN := "a full word"
	exampleEN := "This is a full word."
	exampleRU := "Это полное слово."
	_, _ = db.Exec(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, transcription, example_en, example_ru) VALUES ($1, 'fullword', 0, $2, $3, $4, $5, $6)`,
		wordCardID, wordRU, meaningEN, transcription, exampleEN, exampleRU)

	words, err := repo.GetWordSetWords(setID, user.ID)
	if err != nil {
		t.Fatalf("GetWordSetWords: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	// The word set has no preferred_pos so transcription etc. come as NULL
	// Just verify the basic word is returned
	if words[0].Word != "fullword" {
		t.Errorf("expected word 'fullword', got %q", words[0].Word)
	}
}

// Cover GetWordSetWords with preferred_pos and training cards having all optional fields
func TestWordSetRepository_GetWordSetWords_WithPreferredPOSAndAllFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	tcRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(9012)
	pos := "noun"
	setID, _ := repo.CreateWordSet(&models.WordSet{Title: "POS Full Fields", IsPublished: true, PreferredPOS: &pos})
	wordCardID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "posword", Definition: "def"})
	_ = repo.SetWordSetItems(setID, []int64{wordCardID})

	transcription := "/pɒswɜːd/"
	wordRU := "слово с частью речи"
	meaningEN := "a word with POS"
	exampleEN := "This is a POS word."
	exampleRU := "Это слово с частью речи."
	displayWord := "posword"
	_, err := tcRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "posword",
		POS:         &pos,
		DisplayWord: &displayWord,
		SenseIndex:  0,
		WordRU:      wordRU,
		MeaningEN:   meaningEN,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}
	// Update with transcription and examples
	_, _ = db.Exec(`UPDATE training_cards SET transcription = $1, example_en = $2, example_ru = $3 WHERE word_card_id = $4`,
		transcription, exampleEN, exampleRU, wordCardID)

	words, err := repo.GetWordSetWords(setID, user.ID)
	if err != nil {
		t.Fatalf("GetWordSetWords: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	if words[0].Transcription == nil || *words[0].Transcription != transcription {
		t.Errorf("expected transcription %q, got %v", transcription, words[0].Transcription)
	}
	if words[0].WordRU == nil || *words[0].WordRU != wordRU {
		t.Errorf("expected wordRU %q, got %v", wordRU, words[0].WordRU)
	}
	if words[0].MeaningEN == nil || *words[0].MeaningEN != meaningEN {
		t.Errorf("expected meaningEN %q, got %v", meaningEN, words[0].MeaningEN)
	}
	if words[0].ExampleEN == nil || *words[0].ExampleEN != exampleEN {
		t.Errorf("expected exampleEN %q, got %v", exampleEN, words[0].ExampleEN)
	}
	if words[0].ExampleRU == nil || *words[0].ExampleRU != exampleRU {
		t.Errorf("expected exampleRU %q, got %v", exampleRU, words[0].ExampleRU)
	}
}

// Cover GetWordSetWords with word set not found
func TestWordSetRepository_GetWordSetWords_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetRepository(db, logger)

	_, err := repo.GetWordSetWords(99999, 1)
	if err == nil {
		t.Error("expected error for non-existent word set")
	}
}

// ============================================================
// word_set_category_repository.go error paths
// ============================================================

func TestWordSetCategoryRepository_CreateCategory_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	_, err := repo.CreateCategory(&models.WordSetCategory{Name: "x"})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetCategoryRepository_GetCategory_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	_, err := repo.GetCategory(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetCategoryRepository_GetAllCategories_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	_, err := repo.GetAllCategories()
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetCategoryRepository_GetPublishedCategories_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	_, err := repo.GetPublishedCategories()
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetCategoryRepository_UpdateCategory_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	err := repo.UpdateCategory(&models.WordSetCategory{ID: 1, Name: "x"})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWordSetCategoryRepository_DeleteCategory_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	err := repo.DeleteCategory(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// Cover UpdateCategory with parent_id and description (nil branches)
func TestWordSetCategoryRepository_UpdateCategory_WithParentAndDesc(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	parentID, err := repo.CreateCategory(&models.WordSetCategory{Name: "Parent", SortOrder: 0})
	if err != nil {
		t.Fatalf("CreateCategory parent: %v", err)
	}
	childID, err := repo.CreateCategory(&models.WordSetCategory{Name: "Child", SortOrder: 1, ParentID: &parentID})
	if err != nil {
		t.Fatalf("CreateCategory child: %v", err)
	}

	desc := "updated desc"
	err = repo.UpdateCategory(&models.WordSetCategory{
		ID:          childID,
		Name:        "Updated Child",
		ParentID:    &parentID,
		Description: &desc,
		IsPublished: true,
		SortOrder:   2,
	})
	if err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}

	got, err := repo.GetCategory(childID)
	if err != nil {
		t.Fatalf("GetCategory: %v", err)
	}
	if got.Description == nil || *got.Description != desc {
		t.Errorf("expected description %q, got %v", desc, got.Description)
	}
	if got.ParentID == nil || *got.ParentID != parentID {
		t.Errorf("expected parent_id %d, got %v", parentID, got.ParentID)
	}
}

// Cover GetAllCategories with parent_id and description set (to hit the Valid branches)
func TestWordSetCategoryRepository_GetAllCategories_WithParentAndDesc(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	parentID, _ := repo.CreateCategory(&models.WordSetCategory{Name: "ParentCat", SortOrder: 0})
	desc := "some desc"
	_, _ = repo.CreateCategory(&models.WordSetCategory{
		Name:        "ChildCat",
		SortOrder:   1,
		ParentID:    &parentID,
		Description: &desc,
		IsPublished: true,
	})

	cats, err := repo.GetAllCategories()
	if err != nil {
		t.Fatalf("GetAllCategories: %v", err)
	}
	found := false
	for _, c := range cats {
		if c.Name == "ChildCat" && c.ParentID != nil && c.Description != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected ChildCat with parent_id and description in GetAllCategories result")
	}
}

// Cover GetPublishedCategories with parent_id and description set
func TestWordSetCategoryRepository_GetPublishedCategories_WithParentAndDesc(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetCategoryRepository(db, logger)

	parentID, _ := repo.CreateCategory(&models.WordSetCategory{Name: "PubParent", SortOrder: 0, IsPublished: true})
	desc := "pub desc"
	_, _ = repo.CreateCategory(&models.WordSetCategory{
		Name:        "PubChild",
		SortOrder:   1,
		ParentID:    &parentID,
		Description: &desc,
		IsPublished: true,
	})

	cats, err := repo.GetPublishedCategories()
	if err != nil {
		t.Fatalf("GetPublishedCategories: %v", err)
	}
	found := false
	for _, c := range cats {
		if c.Name == "PubChild" && c.ParentID != nil && c.Description != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected PubChild with parent_id and description in GetPublishedCategories result")
	}
}

// ============================================================
// grammar_attempt_repository.go error paths
// ============================================================

func TestGrammarAttemptRepository_CreateAttempt_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.CreateAttempt(&TestAttempt{
		UserID: 1, ScopeType: "chapter", ScopeID: "ch-1",
		StartedAt: time.Now(), AnswersJSON: "[]", ResultsJSON: "[]",
	})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_UpdateProgress_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	err := repo.UpdateProgress(1, "ch-1", 80, true)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_GetCategoryTestProgress_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.GetCategoryTestProgress(1, "s-1")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_HasCategoryTestAttempt_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.HasCategoryTestAttempt(1, "s-1")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_GetCategoryTestBestScore_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.GetCategoryTestBestScore(1, "s-1")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_GetChapterProgress_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.GetChapterProgress(1, "ch-1")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_GetUserAttempts_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.GetUserAttempts(1, "chapter", "ch-1", 10)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_GetBestScore_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.GetBestScore(1, "chapter", "ch-1")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_GetAverageTestScore_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.GetAverageTestScore(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_SavePlacementTestResult_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	err := repo.SavePlacementTestResult(1, 50, 10, []string{"s1"})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestGrammarAttemptRepository_GetPlacementTestResult_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewGrammarAttemptRepository(db, zap.NewNop())

	_, err := repo.GetPlacementTestResult(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// Cover GetUserAttempts with courseVersion set (Valid branch)
func TestGrammarAttemptRepository_GetUserAttempts_WithCourseVersion(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(200)

	finished := time.Now()
	cv := "v1.0"
	attempt := &TestAttempt{
		UserID:         user.ID,
		ScopeType:      "chapter",
		ScopeID:        "ch-cv",
		StartedAt:      time.Now(),
		FinishedAt:     &finished,
		Score:          70,
		Passed:         true,
		TotalQuestions: 5,
		AnswersJSON:    "[]",
		ResultsJSON:    "[]",
		CourseVersion:  &cv,
	}
	if _, err := repo.CreateAttempt(attempt); err != nil {
		t.Fatalf("CreateAttempt: %v", err)
	}

	attempts, err := repo.GetUserAttempts(user.ID, "chapter", "ch-cv", 10)
	if err != nil {
		t.Fatalf("GetUserAttempts: %v", err)
	}
	if len(attempts) != 1 {
		t.Fatalf("expected 1 attempt, got %d", len(attempts))
	}
	if attempts[0].CourseVersion == nil || *attempts[0].CourseVersion != cv {
		t.Errorf("expected CourseVersion %q, got %v", cv, attempts[0].CourseVersion)
	}
}

// ============================================================
// user_word_mastering_repository.go error paths
// ============================================================

func TestUserWordMasteringRepository_GetWordMasteringStatsBatch_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewUserWordMasteringRepository(db, zap.NewNop())

	_, err := repo.GetWordMasteringStatsBatch([]UserWordPair{{UserID: 1, WordCardID: 1}})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestUserWordMasteringRepository_GetWordCardIDsBySessionID_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewUserWordMasteringRepository(db, zap.NewNop())

	_, err := repo.GetWordCardIDsBySessionID(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestUserWordMasteringRepository_GetKnownWordCardIDsForUser_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewUserWordMasteringRepository(db, zap.NewNop())

	_, err := repo.GetKnownWordCardIDsForUser(1, []int64{1, 2})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestUserWordMasteringRepository_GetKnownForPairs_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewUserWordMasteringRepository(db, zap.NewNop())

	_, err := repo.GetKnownForPairs([]UserWordPair{{UserID: 1, WordCardID: 1}})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestUserWordMasteringRepository_UpsertBatch_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewUserWordMasteringRepository(db, zap.NewNop())

	err := repo.UpsertBatch([]struct {
		UserID     int64
		WordCardID int64
		Score      int
	}{{UserID: 1, WordCardID: 1, Score: 50}})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestUserWordMasteringRepository_GetScore_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewUserWordMasteringRepository(db, zap.NewNop())

	_, err := repo.GetScore(1, 1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// ============================================================
// web_session_repository.go error paths
// ============================================================

func TestWebSessionRepository_CreateSession_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWebSessionRepository(db, logger)

	err := repo.CreateSession(&WebSession{
		UserID:    1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWebSessionRepository_GetSessionByToken_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWebSessionRepository(db, logger)

	_, err := repo.GetSessionByToken("some-token")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWebSessionRepository_DeleteSession_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWebSessionRepository(db, logger)

	err := repo.DeleteSession("some-token")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWebSessionRepository_UpdateLastSeen_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWebSessionRepository(db, logger)

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "some-token"})

	err := repo.UpdateLastSeen(req)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestWebSessionRepository_CleanupExpiredSessions_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewWebSessionRepository(db, logger)

	err := repo.CleanupExpiredSessions()
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// Cover CreateSession with short token (len <= 16)
func TestWebSessionRepository_CreateSession_ShortToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(9001)

	repo := NewWebSessionRepository(db, logger)
	session := &WebSession{
		UserID:    user.ID,
		Token:     "short",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := repo.CreateSession(session); err != nil {
		t.Fatalf("CreateSession short token: %v", err)
	}
	if session.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

// ============================================================
// circuit_breaker_repository.go error paths
// ============================================================

func TestCircuitBreakerRepository_RecordFailure_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	err := repo.RecordFailure("test error")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestCircuitBreakerRepository_Open_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	err := repo.Open()
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestCircuitBreakerRepository_Reset_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	err := repo.Reset()
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestCircuitBreakerRepository_GetState_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	_, err := repo.GetState()
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// ============================================================
// session_repository.go error paths
// ============================================================

func TestSessionRepository_CreateSession_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	_, err := repo.CreateSession(&models.TrainingSession{UserID: 1, Source: models.SourceManual})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestSessionRepository_GetSession_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	_, err := repo.GetSession(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestSessionRepository_UpdateSession_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	err := repo.UpdateSession(&models.TrainingSession{ID: 1})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestSessionRepository_FinishSession_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	err := repo.FinishSession(1, 5)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestSessionRepository_GetActiveSession_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	_, err := repo.GetActiveSession(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestSessionRepository_CreateReviewEvent_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	_, err := repo.CreateReviewEvent(&models.ReviewEvent{UserID: 1, UserCardID: 1})
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestSessionRepository_GetSessionStats_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	_, _, err := repo.GetSessionStats(1)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestSessionRepository_GetTodaySessionCount_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	_, err := repo.GetTodaySessionCount(1, "2025-01-01")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestSessionRepository_GetTrainingStreak_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewSessionRepository(db, logger)

	_, _, err := repo.GetTrainingStreak(1, "UTC", "2025-01-01")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// Cover GetSession with endedAt set (Valid branch)
func TestSessionRepository_GetSession_WithEndedAt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(9002)

	repo := NewSessionRepository(db, logger)
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := repo.FinishSession(id, 3); err != nil {
		t.Fatalf("FinishSession: %v", err)
	}
	got, err := repo.GetSession(id)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
}

// Cover GetActiveSession with endedAt set (Valid branch - finished session has endedAt)
func TestSessionRepository_GetActiveSession_WithEndedAt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(9003)

	repo := NewSessionRepository(db, logger)
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Active session should have no EndedAt
	active, err := repo.GetActiveSession(user.ID)
	if err != nil {
		t.Fatalf("GetActiveSession: %v", err)
	}
	if active == nil {
		t.Fatal("expected active session")
	}
	if active.ID != id {
		t.Errorf("expected session ID %d, got %d", id, active.ID)
	}
}

// ============================================================
// tts_status_repository.go error paths
// ============================================================

func TestTTSStatusRepository_GetByWord_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	_, err := repo.GetByWord("hello")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTTSStatusRepository_UpsertPending_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	err := repo.UpsertPending("hello")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTTSStatusRepository_MarkAttempt_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	err := repo.MarkAttempt("hello", "prov", "code", "msg", true)
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTTSStatusRepository_MarkReady_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	err := repo.MarkReady("hello", "prov", "path")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTTSStatusRepository_MarkTerminal_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	err := repo.MarkTerminal("hello", "prov", "code", "msg")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

func TestTTSStatusRepository_ResetForForceRegenerate_DBError(t *testing.T) {
	db := newBrokenDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	err := repo.ResetForForceRegenerate("hello")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// Cover normalizeTTSWord with digit-only word (no latin letters)
func TestNormalizeTTSWord_NoLatinLetters(t *testing.T) {
	_, ok := normalizeTTSWord("123")
	if ok {
		t.Error("normalizeTTSWord('123') should return ok=false (no latin letters)")
	}
}

// ============================================================
// nudge_repository.go error paths
// ============================================================

func TestNudgeRepository_GetUnconsumedNudge_WithMessageID_Coverage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(9004)

	repo := NewNudgeRepository(db, logger)

	msgID := 42
	nudge := &models.TrainingNudge{
		UserID:         user.ID,
		LocalDate:      "2025-01-15",
		DueCountAtSend: 5,
		MessageID:      &msgID,
	}
	_, err := repo.CreateNudge(nudge)
	if err != nil {
		t.Fatalf("CreateNudge: %v", err)
	}

	got, err := repo.GetUnconsumedNudge(user.ID, "2025-01-15")
	if err != nil {
		t.Fatalf("GetUnconsumedNudge: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil nudge")
	}
	if got.MessageID == nil || *got.MessageID != msgID {
		t.Errorf("expected MessageID %d, got %v", msgID, got.MessageID)
	}
}

func TestNudgeRepository_GetUnconsumedNudge_DBError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := newBrokenDB(t)
	repo := NewNudgeRepository(db, logger)

	_, err := repo.GetUnconsumedNudge(1, "2025-01-15")
	if err == nil {
		t.Error("expected error on broken DB")
	}
}

// Cover nudge_repository.go: consumedAt.Valid branch using sqlmock
func TestNudgeRepository_GetUnconsumedNudge_WithConsumedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row with consumedAt set (sqlmock doesn't enforce WHERE consumed_at IS NULL)
	mock.ExpectQuery("SELECT id, user_id, local_date, sent_at, consumed_at").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "local_date", "sent_at", "consumed_at", "due_count_at_send", "message_id"}).
			AddRow(1, 1, "2025-02-01", "2025-02-01 10:00:00",
				sql.NullString{String: "2025-02-01 11:00:00", Valid: true},
				3, sql.NullInt64{Valid: false}))

	repo := NewNudgeRepository(db, zap.NewNop())
	got, err := repo.GetUnconsumedNudge(1, "2025-02-01")
	if err != nil {
		t.Fatalf("GetUnconsumedNudge: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil nudge")
	}
	if got.ConsumedAt == nil {
		t.Error("expected ConsumedAt to be set")
	}
}

// ============================================================
// word_set_repository.go: additional coverage
// ============================================================

// Cover UpdateWordSet with IsPublished = true (line 214-216)
func TestWordSetRepository_UpdateWordSet_IsPublished(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewWordSetRepository(db, logger)

	id, err := repo.CreateWordSet(&models.WordSet{Title: "Pub Test", IsPublished: false})
	if err != nil {
		t.Fatalf("CreateWordSet: %v", err)
	}

	err = repo.UpdateWordSet(&models.WordSet{ID: id, Title: "Pub Test Updated", IsPublished: true, SortOrder: 1})
	if err != nil {
		t.Fatalf("UpdateWordSet: %v", err)
	}

	got, err := repo.GetWordSet(id)
	if err != nil {
		t.Fatalf("GetWordSet: %v", err)
	}
	if !got.IsPublished {
		t.Error("expected IsPublished to be true")
	}
}

// Cover GetWordSetProgress intermediate errors using sqlmock
func TestWordSetRepository_GetWordSetProgress_CountTotalWordsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordSet succeeds
	mock.ExpectQuery("SELECT .+ FROM word_sets WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "created_at", "updated_at"}).
			AddRow(1, nil, "Test", nil, 1, 0, nil, "2024-01-01 00:00:00", "2024-01-01 00:00:00"))
	// COUNT total words fails
	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(fmt.Errorf("count error"))

	repo := NewWordSetRepository(db, zap.NewNop())
	_, err = repo.GetWordSetProgress(1, 1)
	if err == nil {
		t.Error("expected error when count total words fails")
	}
}

func TestWordSetRepository_GetWordSetProgress_CountKnownWordsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordSet succeeds
	mock.ExpectQuery("SELECT .+ FROM word_sets WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "created_at", "updated_at"}).
			AddRow(1, nil, "Test", nil, 1, 0, nil, "2024-01-01 00:00:00", "2024-01-01 00:00:00"))
	// COUNT total words succeeds
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	// COUNT known words fails
	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(fmt.Errorf("known count error"))

	repo := NewWordSetRepository(db, zap.NewNop())
	_, err = repo.GetWordSetProgress(1, 1)
	if err == nil {
		t.Error("expected error when count known words fails")
	}
}

func TestWordSetRepository_GetWordSetProgress_CountWordsInVocabError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordSet succeeds
	mock.ExpectQuery("SELECT .+ FROM word_sets WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "created_at", "updated_at"}).
			AddRow(1, nil, "Test", nil, 1, 0, nil, "2024-01-01 00:00:00", "2024-01-01 00:00:00"))
	// COUNT total words succeeds
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	// COUNT known words succeeds
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	// COUNT words in vocab fails
	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(fmt.Errorf("vocab count error"))

	repo := NewWordSetRepository(db, zap.NewNop())
	_, err = repo.GetWordSetProgress(1, 1)
	if err == nil {
		t.Error("expected error when count words in vocab fails")
	}
}

// Cover GetWordSetWords query error (line 384) using sqlmock
func TestWordSetRepository_GetWordSetWords_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordSet succeeds
	mock.ExpectQuery("SELECT .+ FROM word_sets WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "created_at", "updated_at"}).
			AddRow(1, nil, "Test", nil, 1, 0, nil, "2024-01-01 00:00:00", "2024-01-01 00:00:00"))
	// Main query fails
	mock.ExpectQuery("SELECT").
		WillReturnError(fmt.Errorf("query error"))

	repo := NewWordSetRepository(db, zap.NewNop())
	_, err = repo.GetWordSetWords(1, 1)
	if err == nil {
		t.Error("expected error when main query fails")
	}
}

// Cover SetWordSetItems transaction errors using sqlmock
func TestWordSetRepository_SetWordSetItems_DeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM word_set_items").
		WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	repo := NewWordSetRepository(db, zap.NewNop())
	err = repo.SetWordSetItems(1, []int64{1, 2})
	if err == nil {
		t.Error("expected error when delete fails")
	}
}

func TestWordSetRepository_SetWordSetItems_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM word_set_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO word_set_items").
		WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	repo := NewWordSetRepository(db, zap.NewNop())
	err = repo.SetWordSetItems(1, []int64{1, 2})
	if err == nil {
		t.Error("expected error when insert fails")
	}
}

func TestWordSetRepository_SetWordSetItems_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM word_set_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO word_set_items").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	repo := NewWordSetRepository(db, zap.NewNop())
	err = repo.SetWordSetItems(1, []int64{1})
	if err == nil {
		t.Error("expected error when commit fails")
	}
}

// ============================================================
// word_set_category_repository.go: parseTime ISO with Z format
// ============================================================

// Cover parseTime ISO format with Z (line 79-81)
func TestParseTime_ISOWithZ(t *testing.T) {
	ts := "2024-03-15T10:30:00Z"
	result, err := parseTime(ts)
	if err != nil {
		t.Fatalf("parseTime(%q) unexpected error: %v", ts, err)
	}
	if result.IsZero() {
		t.Error("expected non-zero time")
	}
}

// Cover DeleteCategory error paths using sqlmock
func TestWordSetCategoryRepository_DeleteCategory_CheckSetsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Check children succeeds (0 children)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// Check sets fails
	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(fmt.Errorf("sets count error"))

	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	err = repo.DeleteCategory(1)
	if err == nil {
		t.Error("expected error when check sets fails")
	}
}

func TestWordSetCategoryRepository_DeleteCategory_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Check children succeeds (0 children)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// Check sets succeeds (0 sets)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// DELETE fails
	mock.ExpectExec("DELETE FROM word_set_categories").
		WillReturnError(fmt.Errorf("delete error"))

	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	err = repo.DeleteCategory(1)
	if err == nil {
		t.Error("expected error when delete fails")
	}
}

// ============================================================
// grammar_attempt_repository.go: additional coverage
// ============================================================

// Cover GetAverageTestScore rows.Err() path using sqlmock
func TestGrammarAttemptRepository_GetAverageTestScore_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"score"}).
		AddRow(80).
		RowError(0, fmt.Errorf("rows error"))
	mock.ExpectQuery("SELECT score").WillReturnRows(rows)

	repo := NewGrammarAttemptRepository(db, zap.NewNop())
	_, err = repo.GetAverageTestScore(1)
	if err == nil {
		t.Error("expected error from rows.Err()")
	}
}

// Cover SavePlacementTestResult insert error (line 381-383) using sqlmock
func TestGrammarAttemptRepository_SavePlacementTestResult_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Check existing: no rows
	mock.ExpectQuery("SELECT score").WillReturnError(sql.ErrNoRows)
	// INSERT fails
	mock.ExpectExec("INSERT INTO grammar_placement_test").
		WillReturnError(fmt.Errorf("insert error"))

	repo := NewGrammarAttemptRepository(db, zap.NewNop())
	err = repo.SavePlacementTestResult(1, 80, 10, []string{"s1"})
	if err == nil {
		t.Error("expected error when insert fails")
	}
}

// Cover SavePlacementTestResult update error (line 397-399) using sqlmock
func TestGrammarAttemptRepository_SavePlacementTestResult_UpdateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Check existing: score=50 (lower than new score 80)
	mock.ExpectQuery("SELECT score").
		WillReturnRows(sqlmock.NewRows([]string{"score", "opened_sections_json"}).AddRow(50, nil))
	// UPDATE fails
	mock.ExpectExec("UPDATE grammar_placement_test").
		WillReturnError(fmt.Errorf("update error"))

	repo := NewGrammarAttemptRepository(db, zap.NewNop())
	err = repo.SavePlacementTestResult(1, 80, 10, []string{"s1"})
	if err == nil {
		t.Error("expected error when update fails")
	}
}

// ============================================================
// session_repository.go: additional coverage
// ============================================================

// Cover GetActiveSession endedAt.Valid branch using sqlmock
func TestSessionRepository_GetActiveSession_EndedAtValid(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a session with endedAt set (even though query has ended_at IS NULL, sqlmock doesn't enforce this)
	mock.ExpectQuery("SELECT id, user_id, started_at, ended_at").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "started_at", "ended_at", "source", "planned_count", "done_count", "session_json"}).
			AddRow(1, 1, "2024-01-01 10:00:00", sql.NullString{String: "2024-01-01 11:00:00", Valid: true}, "manual", 5, 3, "{}"))

	repo := NewSessionRepository(db, zap.NewNop())
	session, err := repo.GetActiveSession(1)
	if err != nil {
		t.Fatalf("GetActiveSession: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if session.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
}

// Cover CreateReviewEvent EarlyReveal=true branch (line 157-159)
func TestSessionRepository_CreateReviewEvent_EarlyReveal(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(9030)

	repo := NewSessionRepository(db, logger)
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	sessionID, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Create a word card and user card for the review event
	wordRepo := NewWordRepository(db, logger)
	wordCardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "earlyword", Definition: "def"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	var tcID int64
	err = db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES ($1, 'earlyword', 0, 'слово', 'word') RETURNING id`, wordCardID).Scan(&tcID)
	if err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	var ucID int64
	err = db.QueryRow(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'en_ru', 'new') RETURNING id`, user.ID, tcID).Scan(&ucID)
	if err != nil {
		t.Fatalf("insert user_card: %v", err)
	}

	now := time.Now()
	event := &models.ReviewEvent{
		SessionID:   &sessionID,
		UserID:      user.ID,
		UserCardID:  ucID,
		Direction:   "en_ru",
		ShownAt:     now,
		EarlyReveal: true, // This covers the earlyReveal=1 branch
		IsCorrect:   true,
		Quality:     5,
	}
	id, err := repo.CreateReviewEvent(event)
	if err != nil {
		t.Fatalf("CreateReviewEvent: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero review event ID")
	}
}

// Cover GetTrainingStreak scan error using sqlmock
func TestSessionRepository_GetTrainingStreak_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row that will fail to scan (wrong type for time.Time)
	rows := sqlmock.NewRows([]string{"started_at"}).AddRow("not-a-time")
	mock.ExpectQuery("SELECT started_at FROM training_sessions").WillReturnRows(rows)

	repo := NewSessionRepository(db, zap.NewNop())
	_, _, err = repo.GetTrainingStreak(1, "UTC", "2025-01-15")
	if err == nil {
		t.Error("expected scan error")
	}
}

// ============================================================
// tts_status_repository.go: additional coverage
// ============================================================

// Cover MarkAttempt second exec error (line 118-120) using sqlmock
func TestTTSStatusRepository_MarkAttempt_SecondExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// First exec (upsert) succeeds
	mock.ExpectExec("INSERT INTO tts_generation_status").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Second exec (cap to terminal) fails
	mock.ExpectExec("UPDATE tts_generation_status").
		WillReturnError(fmt.Errorf("cap error"))

	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	err = repo.MarkAttempt("hello", "", "", "", true)
	if err == nil {
		t.Error("expected error when second exec fails")
	}
}

// ============================================================
// user_word_mastering_repository.go: additional coverage
// ============================================================

// Cover GetWordMasteringStatsBatch scan error using sqlmock
func TestUserWordMasteringRepository_GetWordMasteringStatsBatch_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "word_card_id", "total", "correct", "recent_total", "recent_correct"}).
		AddRow("not-an-int", 1, 10, 5, 3, 2)
	mock.ExpectQuery("WITH ev AS").WillReturnRows(rows)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetWordMasteringStatsBatch([]UserWordPair{{UserID: 1, WordCardID: 1}})
	if err == nil {
		t.Error("expected scan error")
	}
}

// Cover GetWordCardIDsBySessionID scan error using sqlmock
func TestUserWordMasteringRepository_GetWordCardIDsBySessionID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "word_card_id"}).
		AddRow("not-an-int", 1)
	mock.ExpectQuery("SELECT DISTINCT").WillReturnRows(rows)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetWordCardIDsBySessionID(1)
	if err == nil {
		t.Error("expected scan error")
	}
}

// Cover GetWordCardIDsBySessionID rows.Err() using sqlmock
func TestUserWordMasteringRepository_GetWordCardIDsBySessionID_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "word_card_id"}).
		AddRow(1, 1).
		RowError(0, fmt.Errorf("rows error"))
	mock.ExpectQuery("SELECT DISTINCT").WillReturnRows(rows)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetWordCardIDsBySessionID(1)
	if err == nil {
		t.Error("expected rows error")
	}
}

// Cover GetKnownWordCardIDsForUser scan error using sqlmock
func TestUserWordMasteringRepository_GetKnownWordCardIDsForUser_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"word_card_id"}).AddRow("not-an-int")
	mock.ExpectQuery("SELECT word_card_id FROM user_word_knowledge").WillReturnRows(rows)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetKnownWordCardIDsForUser(1, []int64{1, 2})
	if err == nil {
		t.Error("expected scan error")
	}
}

// Cover GetKnownWordCardIDsForUser rows.Err() using sqlmock
func TestUserWordMasteringRepository_GetKnownWordCardIDsForUser_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"word_card_id"}).
		AddRow(1).
		RowError(0, fmt.Errorf("rows error"))
	mock.ExpectQuery("SELECT word_card_id FROM user_word_knowledge").WillReturnRows(rows)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetKnownWordCardIDsForUser(1, []int64{1, 2})
	if err == nil {
		t.Error("expected rows error")
	}
}

// Cover GetKnownForPairs scan error using sqlmock
func TestUserWordMasteringRepository_GetKnownForPairs_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "word_card_id"}).
		AddRow("not-an-int", 1)
	mock.ExpectQuery("SELECT user_id, word_card_id FROM user_word_knowledge").WillReturnRows(rows)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetKnownForPairs([]UserWordPair{{UserID: 1, WordCardID: 1}})
	if err == nil {
		t.Error("expected scan error")
	}
}

// Cover GetKnownForPairs rows.Err() using sqlmock
func TestUserWordMasteringRepository_GetKnownForPairs_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "word_card_id"}).
		AddRow(1, 1).
		RowError(0, fmt.Errorf("rows error"))
	mock.ExpectQuery("SELECT user_id, word_card_id FROM user_word_knowledge").WillReturnRows(rows)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetKnownForPairs([]UserWordPair{{UserID: 1, WordCardID: 1}})
	if err == nil {
		t.Error("expected rows error")
	}
}

// Cover GetWordMasteringStatsBatch rows.Err() using sqlmock (line 98-100)
func TestUserWordMasteringRepository_GetWordMasteringStatsBatch_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "word_card_id", "total", "correct", "recent_total", "recent_correct"}).
		AddRow(1, 1, 10, 5, 3, 2).
		RowError(0, fmt.Errorf("rows error"))
	mock.ExpectQuery("WITH ev AS").WillReturnRows(rows)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetWordMasteringStatsBatch([]UserWordPair{{UserID: 1, WordCardID: 1}})
	if err == nil {
		t.Error("expected rows error")
	}
}

// Cover GetWordCardIDsBySessionID rows.Err() - already tested above but verify
// Cover GetScore DB error path using sqlmock
func TestUserWordMasteringRepository_GetScore_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT mastering_score").WillReturnError(fmt.Errorf("db error"))

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	_, err = repo.GetScore(1, 1)
	if err == nil {
		t.Error("expected error from db")
	}
}

// ============================================================
// circuit_breaker_repository.go: alternative time parse formats
// ============================================================

// Cover circuit_breaker_repository.go alternative time parse formats using sqlmock
func TestCircuitBreakerRepository_GetState_AltTimeFormats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return times in ISO 8601 with Z format to trigger alternative parse branches
	mock.ExpectQuery("SELECT id, is_open, failure_count").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "is_open", "failure_count", "last_failure_at", "last_failure_message", "last_reset_at", "updated_at",
		}).AddRow(
			1, false, 2,
			sql.NullString{String: "2024-01-01T10:00:00Z", Valid: true},
			"",
			sql.NullString{String: "2024-01-01T09:00:00Z", Valid: true},
			sql.NullString{String: "2024-01-01T10:00:00Z", Valid: true},
		))

	repo := NewCircuitBreakerRepository(db, zap.NewNop())
	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil state")
	}
	if state.LastFailureAt == nil {
		t.Error("expected LastFailureAt to be set")
	}
}

func TestCircuitBreakerRepository_GetState_AltTimeFormats2(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return times in ISO 8601 without Z to trigger the third alternative parse branch
	mock.ExpectQuery("SELECT id, is_open, failure_count").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "is_open", "failure_count", "last_failure_at", "last_failure_message", "last_reset_at", "updated_at",
		}).AddRow(
			1, true, 1,
			sql.NullString{String: "2024-01-01T10:00:00", Valid: true},
			"",
			sql.NullString{String: "2024-01-01T09:00:00", Valid: true},
			sql.NullString{String: "2024-01-01T10:00:00", Valid: true},
		))

	repo := NewCircuitBreakerRepository(db, zap.NewNop())
	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil state")
	}
}

// Cover initializeState error path using sqlmock
func TestCircuitBreakerRepository_initializeState_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetState returns no rows (triggers initializeState)
	mock.ExpectQuery("SELECT id, is_open, failure_count").WillReturnError(sql.ErrNoRows)
	// initializeState INSERT fails
	mock.ExpectExec("INSERT INTO circuit_breaker_state").
		WillReturnError(fmt.Errorf("insert error"))

	repo := NewCircuitBreakerRepository(db, zap.NewNop())
	_, err = repo.GetState()
	if err == nil {
		t.Error("expected error when initializeState fails")
	}
}

// ============================================================
// web_session_repository.go: alternative time parse formats
// ============================================================

// Cover web_session_repository.go alternative time parse formats using sqlmock
// Format 1: ISO 8601 with Z (line 119)
func TestWebSessionRepository_GetSessionByToken_ISO8601Z(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, user_id, session_token, expires_at, created_at, last_seen_at").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "session_token", "expires_at", "created_at", "last_seen_at",
		}).AddRow(1, 1, "tok1",
			"2024-12-31T23:59:59Z", "2024-01-01T00:00:00Z", "2024-06-15T12:00:00Z"))

	repo := NewWebSessionRepository(db, zap.NewNop())
	session, err := repo.GetSessionByToken("tok1")
	if err != nil {
		t.Fatalf("GetSessionByToken: %v", err)
	}
	if session == nil || session.ExpiresAt.IsZero() {
		t.Error("expected non-nil session with non-zero ExpiresAt")
	}
}

// Format 2: ISO 8601 with timezone offset (line 123)
func TestWebSessionRepository_GetSessionByToken_ISO8601Offset(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, user_id, session_token, expires_at, created_at, last_seen_at").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "session_token", "expires_at", "created_at", "last_seen_at",
		}).AddRow(1, 1, "tok2",
			"2024-12-31T23:59:59-05:00", "2024-01-01T00:00:00-05:00", "2024-06-15T12:00:00-05:00"))

	repo := NewWebSessionRepository(db, zap.NewNop())
	session, err := repo.GetSessionByToken("tok2")
	if err != nil {
		t.Fatalf("GetSessionByToken: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
}

// Format 3: custom format "2006-01-02 15:04:05" (line 127)
func TestWebSessionRepository_GetSessionByToken_CustomFormat(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// This format fails formats 1 and 2 but passes format 3
	mock.ExpectQuery("SELECT id, user_id, session_token, expires_at, created_at, last_seen_at").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "session_token", "expires_at", "created_at", "last_seen_at",
		}).AddRow(1, 1, "tok3",
			"2024-12-31 23:59:59", "2024-01-01 00:00:00", "2024-06-15 12:00:00"))

	repo := NewWebSessionRepository(db, zap.NewNop())
	session, err := repo.GetSessionByToken("tok3")
	if err != nil {
		t.Fatalf("GetSessionByToken: %v", err)
	}
	if session == nil || session.ExpiresAt.IsZero() {
		t.Error("expected non-nil session with non-zero ExpiresAt")
	}
}

// Format 4: RFC3339 (line 131) - use fractional seconds which fail formats 1-3
func TestWebSessionRepository_GetSessionByToken_RFC3339Format(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// RFC3339 with fractional seconds - fails formats 1, 2, 3 but passes RFC3339
	mock.ExpectQuery("SELECT id, user_id, session_token, expires_at, created_at, last_seen_at").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "session_token", "expires_at", "created_at", "last_seen_at",
		}).AddRow(1, 1, "tok4",
			"2024-12-31T23:59:59.000+03:00",
			"2024-01-01T00:00:00.000+03:00",
			"2024-06-15T12:00:00.000+03:00"))

	repo := NewWebSessionRepository(db, zap.NewNop())
	session, err := repo.GetSessionByToken("tok4")
	if err != nil {
		t.Fatalf("GetSessionByToken: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
}

// Warn path: all formats fail (line 134)
func TestWebSessionRepository_GetSessionByToken_UnparsableTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, user_id, session_token, expires_at, created_at, last_seen_at").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "session_token", "expires_at", "created_at", "last_seen_at",
		}).AddRow(1, 1, "tok5", "not-a-time", "also-not-a-time", "still-not-a-time"))

	repo := NewWebSessionRepository(db, zap.NewNop())
	session, err := repo.GetSessionByToken("tok5")
	if err != nil {
		t.Fatalf("GetSessionByToken: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session (even with bad times)")
	}
	if !session.ExpiresAt.IsZero() {
		t.Error("expected zero ExpiresAt for unparsable time")
	}
}

// ============================================================
// Additional coverage for remaining uncovered lines
// ============================================================

// Cover word_set_repository.go:166-168 - scan error in ListWordSets loop using sqlmock
func TestWordSetRepository_ListWordSets_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row that will fail to scan (wrong type for id)
	rows := sqlmock.NewRows([]string{"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "created_at", "updated_at"}).
		AddRow("not-an-int", nil, "Test", nil, 1, 0, nil, "2024-01-01 00:00:00", "2024-01-01 00:00:00")
	mock.ExpectQuery("SELECT id, category_id, title").WillReturnRows(rows)

	repo := NewWordSetRepository(db, zap.NewNop())
	sets, err := repo.ListWordSets(nil, 10, 0)
	if err != nil {
		t.Fatalf("ListWordSets: %v", err)
	}
	// Scan error causes continue, so result should be empty
	if len(sets) != 0 {
		t.Errorf("expected empty list when scan fails, got %d", len(sets))
	}
}

// Cover word_set_repository.go:287-289 - unknownWords < 0 using sqlmock
func TestWordSetRepository_GetWordSetProgress_NegativeUnknown(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordSet succeeds
	mock.ExpectQuery("SELECT .+ FROM word_sets WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "created_at", "updated_at"}).
			AddRow(1, nil, "Test", nil, 1, 0, nil, "2024-01-01 00:00:00", "2024-01-01 00:00:00"))
	// COUNT total = 3
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	// COUNT known = 2
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	// COUNT in_vocab = 2 (total=3, known=2, in_vocab=2 => unknown=-1 < 0)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	repo := NewWordSetRepository(db, zap.NewNop())
	progress, err := repo.GetWordSetProgress(1, 1)
	if err != nil {
		t.Fatalf("GetWordSetProgress: %v", err)
	}
	if progress == nil {
		t.Fatal("expected non-nil progress")
	}
	// unknownWords should be clamped to 0
	if progress.UnknownWords < 0 {
		t.Error("expected UnknownWords >= 0")
	}
}

// Cover word_set_repository.go:410-412 - scan error in GetWordSetWords loop using sqlmock
func TestWordSetRepository_GetWordSetWords_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordSet succeeds (no preferred_pos)
	mock.ExpectQuery("SELECT .+ FROM word_sets WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "created_at", "updated_at"}).
			AddRow(1, nil, "Test", nil, 1, 0, nil, "2024-01-01 00:00:00", "2024-01-01 00:00:00"))
	// Main query returns row with wrong type for word_card_id
	rows := sqlmock.NewRows([]string{"word_card_id", "word", "status", "display_word_pref", "transcription_pref", "word_ru_pref", "meaning_en_pref", "example_en_pref", "example_ru_pref"}).
		AddRow("not-an-int", "word", "unknown", nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewWordSetRepository(db, zap.NewNop())
	words, err := repo.GetWordSetWords(1, 1)
	if err != nil {
		t.Fatalf("GetWordSetWords: %v", err)
	}
	// Scan error causes continue, so result should be empty
	if len(words) != 0 {
		t.Errorf("expected empty list when scan fails, got %d", len(words))
	}
}

// Cover word_set_repository.go:418-420 - displayWord else branch using sqlmock
func TestWordSetRepository_GetWordSetWords_DisplayWordEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetWordSet succeeds (no preferred_pos)
	mock.ExpectQuery("SELECT .+ FROM word_sets WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "title", "description", "is_published", "sort_order", "preferred_pos", "created_at", "updated_at"}).
			AddRow(1, nil, "Test", nil, 1, 0, nil, "2024-01-01 00:00:00", "2024-01-01 00:00:00"))
	// Main query returns row with NULL display_word (triggers else branch)
	rows := sqlmock.NewRows([]string{"word_card_id", "word", "status", "display_word_pref", "transcription_pref", "word_ru_pref", "meaning_en_pref", "example_en_pref", "example_ru_pref"}).
		AddRow(1, "testword", "unknown", sql.NullString{Valid: false}, nil, nil, nil, nil, nil)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewWordSetRepository(db, zap.NewNop())
	words, err := repo.GetWordSetWords(1, 1)
	if err != nil {
		t.Fatalf("GetWordSetWords: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	// DisplayWord should fall back to Word
	if words[0].DisplayWord != "testword" {
		t.Errorf("expected DisplayWord='testword', got %q", words[0].DisplayWord)
	}
}

// Cover word_set_category_repository.go:79-81 - parseTime ISO with Z
func TestParseTime_ISOWithZ_Coverage(t *testing.T) {
	ts := "2024-03-15T10:30:00Z"
	result, err := parseTime(ts)
	if err != nil {
		t.Fatalf("parseTime(%q) unexpected error: %v", ts, err)
	}
	if result.IsZero() {
		t.Error("expected non-zero time for ISO with Z format")
	}
}

// Cover word_set_category_repository.go:172-174 - scan error in GetAllCategories loop using sqlmock
func TestWordSetCategoryRepository_GetAllCategories_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "parent_id", "name", "description", "is_published", "sort_order", "created_at", "updated_at"}).
		AddRow("not-an-int", nil, "Cat", nil, 1, 0, "2024-01-01 00:00:00", "2024-01-01 00:00:00")
	mock.ExpectQuery("SELECT id, parent_id, name").WillReturnRows(rows)

	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	cats, err := repo.GetAllCategories()
	if err != nil {
		t.Fatalf("GetAllCategories: %v", err)
	}
	// Scan error causes continue, so result should be empty
	if len(cats) != 0 {
		t.Errorf("expected empty list when scan fails, got %d", len(cats))
	}
}

// Cover word_set_category_repository.go:235-237 - scan error in GetPublishedCategories loop using sqlmock
func TestWordSetCategoryRepository_GetPublishedCategories_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "parent_id", "name", "description", "is_published", "sort_order", "created_at", "updated_at"}).
		AddRow("not-an-int", nil, "Cat", nil, 1, 0, "2024-01-01 00:00:00", "2024-01-01 00:00:00")
	mock.ExpectQuery("SELECT id, parent_id, name").WillReturnRows(rows)

	repo := NewWordSetCategoryRepository(db, zap.NewNop())
	cats, err := repo.GetPublishedCategories()
	if err != nil {
		t.Fatalf("GetPublishedCategories: %v", err)
	}
	// Scan error causes continue, so result should be empty
	if len(cats) != 0 {
		t.Errorf("expected empty list when scan fails, got %d", len(cats))
	}
}

// Cover tts_status_repository.go:48-50 - GetByWord no rows
func TestTTSStatusRepository_GetByWord_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT word, state").WillReturnError(sql.ErrNoRows)

	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	result, err := repo.GetByWord("nonexistent")
	if err != nil {
		t.Fatalf("GetByWord: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for no rows")
	}
}

// Cover tts_status_repository.go:221-223 - normalizeTTSWord no latin
func TestNormalizeTTSWord_NoLatinLetters_Coverage(t *testing.T) {
	_, ok := normalizeTTSWord("привет")
	if ok {
		t.Error("normalizeTTSWord('привет') should return ok=false (no latin letters)")
	}
}

// Cover session_repository.go:237-239 - rows.Err in GetTrainingStreak using sqlmock
func TestSessionRepository_GetTrainingStreak_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"started_at"}).
		AddRow(time.Now()).
		RowError(0, fmt.Errorf("rows error"))
	mock.ExpectQuery("SELECT started_at FROM training_sessions").WillReturnRows(rows)

	repo := NewSessionRepository(db, zap.NewNop())
	_, _, err = repo.GetTrainingStreak(1, "UTC", "2025-01-15")
	if err == nil {
		t.Error("expected rows error")
	}
}

// Cover grammar_attempt_repository.go:243-245 - scan error in GetUserAttempts loop using sqlmock
func TestGrammarAttemptRepository_GetUserAttempts_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "user_id", "scope_type", "scope_id", "started_at", "finished_at", "score", "passed", "total_questions", "answers_json", "results_json", "course_version"}).
		AddRow("not-an-int", 1, "chapter", "ch1", nil, nil, 80, 1, 10, "[]", "[]", nil)
	mock.ExpectQuery("SELECT id, user_id, scope_type").WillReturnRows(rows)

	repo := NewGrammarAttemptRepository(db, zap.NewNop())
	_, err = repo.GetUserAttempts(1, "chapter", "ch1", 10)
	if err == nil {
		t.Error("expected scan error")
	}
}

// Cover grammar_attempt_repository.go:304-305 - scan error (continue) in GetAverageTestScore using sqlmock
func TestGrammarAttemptRepository_GetAverageTestScore_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row that will fail to scan (wrong type for score)
	rows := sqlmock.NewRows([]string{"score"}).AddRow("not-an-int")
	mock.ExpectQuery("SELECT score").WillReturnRows(rows)

	repo := NewGrammarAttemptRepository(db, zap.NewNop())
	// Scan error causes continue, so result should be 0 (no valid scores)
	score, err := repo.GetAverageTestScore(1)
	if err != nil {
		t.Fatalf("GetAverageTestScore: %v", err)
	}
	if score != 0 {
		t.Errorf("expected 0 when all scans fail, got %d", score)
	}
}

// Cover circuit_breaker_repository.go:80-82 - third alternative time format (ISO without Z)
// This is already covered by TestCircuitBreakerRepository_GetState_AltTimeFormats2
// but let's add a specific test for the updatedAt field third format
func TestCircuitBreakerRepository_GetState_UpdatedAtThirdFormat(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// updatedAt in ISO format without Z (fails RFC3339 and "2006-01-02 15:04:05" but passes "2006-01-02T15:04:05")
	mock.ExpectQuery("SELECT id, is_open, failure_count").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "is_open", "failure_count", "last_failure_at", "last_failure_message", "last_reset_at", "updated_at",
		}).AddRow(
			1, false, 0,
			sql.NullString{Valid: false},
			"",
			sql.NullString{Valid: false},
			sql.NullString{String: "2024-01-01T10:00:00", Valid: true},
		))

	repo := NewCircuitBreakerRepository(db, zap.NewNop())
	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil state")
	}
}

// Cover normalizeTTSWord !hasLatin branch (line 221-223):
// word contains only spaces/hyphens but no letters
func TestNormalizeTTSWord_OnlyHyphens_NoLatin(t *testing.T) {
	_, ok := normalizeTTSWord("- -")
	if ok {
		t.Error("normalizeTTSWord('- -') should return ok=false (no latin letters)")
	}
}

// Cover GetWordMasteringStatsBatch with 2+ pairs to hit the i>0 branch (line 54-56)
func TestUserWordMasteringRepository_GetWordMasteringStatsBatch_MultiplePairs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("WITH ev AS").WillReturnRows(
		sqlmock.NewRows([]string{"user_id", "word_card_id", "total", "correct", "recent_total", "recent_correct"}),
	)

	repo := NewUserWordMasteringRepository(db, zap.NewNop())
	pairs := []UserWordPair{
		{UserID: 1, WordCardID: 1},
		{UserID: 1, WordCardID: 2},
	}
	result, err := repo.GetWordMasteringStatsBatch(pairs)
	if err != nil {
		t.Fatalf("GetWordMasteringStatsBatch: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}
