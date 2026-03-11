package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

const twInvalidDSN = "postgres://x:x@invalid.invalid:1/db?connect_timeout=1"

// openInvalidSQLDB opens a *sql.DB with an invalid DSN to trigger errors on all operations.
// The caller must ensure postgres_compat driver is already registered (e.g. by calling
// testutil.SetupTestDB or testutil.SetupTestDatabase first).
func openInvalidSQLDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres_compat", twInvalidDSN)
	if err != nil {
		t.Skipf("postgres_compat driver not registered: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func newInvalidWordRepo(t *testing.T, logger *zap.Logger) *repository.WordRepository {
	t.Helper()
	return repository.NewWordRepository(openInvalidSQLDB(t), logger)
}

func newInvalidTrainingCardRepo(t *testing.T, logger *zap.Logger) *repository.TrainingCardRepository {
	t.Helper()
	return repository.NewTrainingCardRepository(openInvalidSQLDB(t), logger)
}

func newInvalidUserCardRepo(t *testing.T, logger *zap.Logger) *repository.UserCardRepository {
	t.Helper()
	return repository.NewUserCardRepository(openInvalidSQLDB(t), logger)
}

func newInvalidUserRepo(t *testing.T, logger *zap.Logger) *repository.UserRepository {
	t.Helper()
	return repository.NewUserRepository(openInvalidSQLDB(t), logger)
}

// TestTrainingWorker_processCards_GetCardsError covers the error path when
// GetWordCardsWithoutTrainingCards returns an error.
func TestTrainingWorker_processCards_GetCardsError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)

	// Insert a word card so processCards has something to fetch (otherwise returns early)
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "getcardserr", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_ = cardID

	mockCB := &mockCircuitBreakerForWorker{}
	worker := NewTrainingWorker(
		ai.NewService("", "", "", "", logger),
		wordRepo,
		newInvalidTrainingCardRepo(t, logger),
		userCardRepo,
		userRepo,
		nil,
		mockCB,
		nil,
		0,
		2,
		1,
		0,
		"",
		logger,
	)

	// processCards should log error and return without panic
	worker.processCards(context.Background())
}

// TestTrainingWorker_fillWordCardData_VerbPOSFromWordInfo covers the branch where
// wordCardModel.POS is "noun" (preserved) but wordInfo.POS is "verb" (else if branch).
func TestTrainingWorker_fillWordCardData_VerbPOSFromWordInfo(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		// wordInfo.POS = "verb", VerbForms.V1 = "run"
		content := `{"input_word":"run","lemma":"run","pos":"verb","transcription":"rʌn","definition_ru":"бежать","verb_forms":{"v1":"run","v2":"ran","v3":"run"}}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	// Create a word card with POS="noun" already set (preserved), but missing transcription/definition
	// wordCardModel.POS will be "noun" (preserved from existing), wordInfo.POS = "verb"
	// This triggers: else if wordInfo.POS == "verb" && wordInfo.VerbForms != nil && wordInfo.VerbForms.V1 != ""
	pos := "noun"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "runverb", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}
	// wordCard.POS = "noun", missing transcription and definitionRU -> hasMissingData = true
	// wordCardModel.POS = "noun" (preserved), wordInfo.POS = "verb"
	// -> else if branch fires -> displayEN = "to run"

	if err := worker.fillWordCardData(context.Background(), wordCard); err != nil {
		t.Fatalf("fillWordCardData: %v", err)
	}

	updated, _ := wordRepo.GetWordCardByID(cardID)
	if updated.DisplayEN == nil || *updated.DisplayEN != "to run" {
		t.Errorf("expected display_en 'to run' from else-if branch, got %q", ptrStr(updated.DisplayEN))
	}
}

// TestTrainingWorker_fillWordCardData_UpdateWordCardError covers the UpdateWordCard error path.
func TestTrainingWorker_fillWordCardData_UpdateWordCardError(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		content := `{"input_word":"errword","lemma":"errword","pos":"noun","transcription":"ɛr","definition_ru":"ошибка"}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "errword", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	// Replace wordRepo with an invalid one so UpdateWordCard fails
	logger, _ := zap.NewDevelopment()
	worker.wordRepo = newInvalidWordRepo(t, logger)

	err = worker.fillWordCardData(context.Background(), wordCard)
	if err == nil {
		t.Fatal("expected error when UpdateWordCard fails")
	}
}

// TestTrainingWorker_processCard_ReloadError covers the GetWordCardByID error path when reloading.
func TestTrainingWorker_processCard_ReloadError(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// fillWordCardData fails (no missing data needed)
			return nil, fmt.Errorf("AI unavailable")
		}
		// GenerateTrainingCard
		content := `{"word_en":"reloadw","lemma":"reloadw","transcription":"","senses":[{"pos":"noun","word_ru":"перезагрузка","meaning_en":"reload","example_en":"","example_ru":"","distractors_ru":["загрузка","выгрузка"],"distractors_en":["load","unload"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, trainingCardRepo, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "reloadw", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	// Replace wordRepo with an invalid one so GetWordCardByID fails during reload
	logger, _ := zap.NewDevelopment()
	worker.wordRepo = newInvalidWordRepo(t, logger)

	// processCard should continue with original wordCard even if reload fails
	// (training card creation will also fail, but reload error path is covered)
	_ = worker.processCard(context.Background(), wordCard)
	_ = trainingCardRepo
}

// TestTrainingWorker_processCard_LLMGenerationFailed covers the LLM generation failed error.
func TestTrainingWorker_processCard_LLMGenerationFailed(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// fillWordCardData: AI unavailable
			return nil, fmt.Errorf("AI unavailable")
		}
		// GenerateTrainingCard: fail
		return nil, fmt.Errorf("LLM generation failed")
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	pos, trans, def := "noun", "tɛst", "тест"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word: "testfail2", Definition: "", POS: &pos, Transcription: &trans, DefinitionRU: &def,
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err == nil {
		t.Fatal("expected error when LLM generation fails")
	}
}

// TestTrainingWorker_processCard_MarkProcessedErrorFails covers MarkWordCardProcessedError error
// when LLM rejects the word.
func TestTrainingWorker_processCard_MarkProcessedErrorFails(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		content := `{"error": "not a valid word", "word_en": "badword2"}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	pos, trans, def := "noun", "bæd", "плохой"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word: "badword2", Definition: "", POS: &pos, Transcription: &trans, DefinitionRU: &def,
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	// Replace wordRepo with invalid one so MarkWordCardProcessedError fails
	logger, _ := zap.NewDevelopment()
	worker.wordRepo = newInvalidWordRepo(t, logger)

	// Should return nil (not an error) even when MarkWordCardProcessedError fails
	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should return nil even when MarkWordCardProcessedError fails: %v", err)
	}
}

// TestTrainingWorker_processCard_HighModelParseError covers the high model parse error path.
func TestTrainingWorker_processCard_HighModelParseError(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// fillWordCardData: skip
			return nil, fmt.Errorf("skip fill")
		}
		if callCount == 2 {
			// Default model: invalid (Cyrillic in distractors_en)
			content := `{"word_en":"jump3","lemma":"jump3","transcription":"","senses":[{"pos":"verb","word_ru":"прыгать","meaning_en":"jump","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["прыгать","to run"],"hint":""}]}`
			resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
			return newJSONHTTPResponseTW(http.StatusOK, resp), nil
		}
		// High model: returns invalid JSON
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: "not json at all"}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	logger, _ := zap.NewDevelopment()
	worker.cbService = NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger)
	worker.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "jump3", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard with high model parse error should return nil: %v", err)
	}
}

// TestTrainingWorker_processCard_HighModelLLMError covers the high model LLM error path.
func TestTrainingWorker_processCard_HighModelLLMError(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return nil, fmt.Errorf("skip fill")
		}
		if callCount == 2 {
			// Default model: invalid
			content := `{"word_en":"swim2","lemma":"swim2","transcription":"","senses":[{"pos":"verb","word_ru":"плыть","meaning_en":"swim","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["плыть","to run"],"hint":""}]}`
			resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
			return newJSONHTTPResponseTW(http.StatusOK, resp), nil
		}
		// High model: LLM error
		return nil, fmt.Errorf("high model LLM error")
	})

	worker, wordRepo, _, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	logger, _ := zap.NewDevelopment()
	worker.cbService = NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger)
	worker.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "swim2", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard with high model LLM error should return nil: %v", err)
	}
}

// TestTrainingWorker_processCard_HighModelLLMRejectsWord covers the high model LLM error (word rejected).
func TestTrainingWorker_processCard_HighModelLLMRejectsWord(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return nil, fmt.Errorf("skip fill")
		}
		if callCount == 2 {
			// Default model: invalid
			content := `{"word_en":"xyz3","lemma":"xyz3","transcription":"","senses":[{"pos":"verb","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["тест","to run"],"hint":""}]}`
			resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
			return newJSONHTTPResponseTW(http.StatusOK, resp), nil
		}
		// High model: rejects word
		content := `{"error": "not a valid English word", "word_en": "xyz3"}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	logger, _ := zap.NewDevelopment()
	worker.cbService = NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger)
	worker.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "xyz3", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard with high model rejecting word should return nil: %v", err)
	}
}

// TestTrainingWorker_processCard_HighModelNoSenses covers the high model no senses path.
func TestTrainingWorker_processCard_HighModelNoSenses(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return nil, fmt.Errorf("skip fill")
		}
		if callCount == 2 {
			// Default model: invalid
			content := `{"word_en":"nosenses2","lemma":"nosenses2","transcription":"","senses":[{"pos":"verb","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["тест","to run"],"hint":""}]}`
			resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
			return newJSONHTTPResponseTW(http.StatusOK, resp), nil
		}
		// High model: returns empty senses
		content := `{"word_en":"nosenses2","lemma":"nosenses2","transcription":"","senses":[]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	logger, _ := zap.NewDevelopment()
	worker.cbService = NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger)
	worker.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "nosenses2", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard with high model no senses should return nil: %v", err)
	}
}

// TestTrainingWorker_processCard_HighModelValidationAlsoFails covers the path where high model
// validation also fails (validationError = highValidationError).
func TestTrainingWorker_processCard_HighModelValidationAlsoFails(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return nil, fmt.Errorf("skip fill")
		}
		// Both default and high model return invalid responses (Cyrillic in distractors_en)
		content := `{"word_en":"valfail2","lemma":"valfail2","transcription":"","senses":[{"pos":"verb","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["тест","to run"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	logger, _ := zap.NewDevelopment()
	worker.cbService = NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger)
	worker.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "valfail2", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard with both models failing validation should return nil: %v", err)
	}
}

// TestTrainingWorker_processCard_ValidationMarkProcessedErrorFails covers MarkWordCardProcessedError
// error path when validation fails.
func TestTrainingWorker_processCard_ValidationMarkProcessedErrorFails(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		// Return invalid distractors_en (Cyrillic)
		content := `{"word_en":"markfail2","lemma":"markfail2","transcription":"","senses":[{"pos":"verb","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["тест","to run"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	worker.modelHigh = "" // no high model

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "markfail2", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	// Replace wordRepo with invalid one so MarkWordCardProcessedError fails
	logger, _ := zap.NewDevelopment()
	worker.wordRepo = newInvalidWordRepo(t, logger)

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should return nil even when MarkWordCardProcessedError fails: %v", err)
	}
}

// TestTrainingWorker_processCard_CreateTrainingCardError covers CreateTrainingCard error path.
func TestTrainingWorker_processCard_CreateTrainingCardError(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		content := `{"word_en":"createfail2","lemma":"createfail2","transcription":"","senses":[{"pos":"noun","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["яблоко","груша"],"distractors_en":["orange","banana"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	pos, trans, def := "noun", "krɪˈeɪt", "создать"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word: "createfail2", Definition: "", POS: &pos, Transcription: &trans, DefinitionRU: &def,
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	// Replace trainingCardRepo with invalid one
	logger, _ := zap.NewDevelopment()
	worker.trainingCardRepo = newInvalidTrainingCardRepo(t, logger)

	err = worker.processCard(context.Background(), wordCard)
	if err == nil {
		t.Fatal("expected error when CreateTrainingCard fails")
	}
}

// TestTrainingWorker_processCard_GetUsersForWordError covers getUsersForWord error path.
func TestTrainingWorker_processCard_GetUsersForWordError(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		content := `{"word_en":"usererr2","lemma":"usererr2","transcription":"","senses":[{"pos":"noun","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["яблоко","груша"],"distractors_en":["orange","banana"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	pos, trans, def := "noun", "juzər", "пользователь"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word: "usererr2", Definition: "", POS: &pos, Transcription: &trans, DefinitionRU: &def,
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	// Replace wordRepo with invalid one so getUsersForWord fails
	logger, _ := zap.NewDevelopment()
	worker.wordRepo = newInvalidWordRepo(t, logger)

	// Should succeed (getUsersForWord error is non-fatal, users=[])
	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should succeed even when getUsersForWord fails: %v", err)
	}
}

// TestTrainingWorker_processCard_CreateUserCardErrors covers CreateUserCard error paths.
func TestTrainingWorker_processCard_CreateUserCardErrors(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		content := `{"word_en":"ucerr2","lemma":"ucerr2","transcription":"","senses":[{"pos":"noun","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["яблоко","груша"],"distractors_en":["orange","banana"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, userRepo, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	pos, trans, def := "noun", "juːsər", "пользователь"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word: "ucerr2", Definition: "", POS: &pos, Transcription: &trans, DefinitionRU: &def,
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	user, err := userRepo.GetOrCreateUser(9001)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	word := "ucerr2"
	if err := wordRepo.AddWordRequestHistoryWithCard(user.ID, word, &cardID, &word); err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard: %v", err)
	}

	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	// Replace userCardRepo with invalid one so CreateUserCard fails
	logger, _ := zap.NewDevelopment()
	worker.userCardRepo = newInvalidUserCardRepo(t, logger)

	// Should succeed (CreateUserCard errors are logged as warnings, not returned)
	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should succeed even when CreateUserCard fails: %v", err)
	}
}

// TestTrainingWorker_getUsersForWord_GetUserIDsByWordError covers GetUserIDsByWord error path.
func TestTrainingWorker_getUsersForWord_GetUserIDsByWordError(t *testing.T) {
	worker, _, _, _, _, _, cleanup := newTrainingWorker(t, nil)
	defer cleanup()

	// Replace wordRepo with invalid one
	logger, _ := zap.NewDevelopment()
	worker.wordRepo = newInvalidWordRepo(t, logger)

	_, err := worker.getUsersForWord("anyword")
	if err == nil {
		t.Fatal("expected error when GetUserIDsByWord fails")
	}
}

// TestTrainingWorker_getUsersForWord_GetUserByIDError covers GetUserByID error + GetOrCreateUser fallback.
func TestTrainingWorker_getUsersForWord_GetUserByIDError(t *testing.T) {
	worker, wordRepo, _, _, userRepo, _, cleanup := newTrainingWorker(t, nil)
	defer cleanup()

	// Create a user and add word request history
	u, err := userRepo.GetOrCreateUser(9002)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	word := "fallbackword"
	if err := wordRepo.AddWordRequestHistoryWithCard(u.ID, word, nil, &word); err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard: %v", err)
	}

	// Replace userRepo with invalid one so GetUserByID fails
	// The fallback GetOrCreateUser will also fail with invalid repo
	logger, _ := zap.NewDevelopment()
	worker.userRepo = newInvalidUserRepo(t, logger)

	// getUsersForWord should return empty slice (user lookup failed, continue)
	users, err := worker.getUsersForWord(word)
	if err != nil {
		t.Fatalf("getUsersForWord should not return error when user lookup fails: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users when user lookup fails, got %d", len(users))
	}
}

// TestTrainingWorker_getUsersForWord_TelegramIDSeenSkipped covers the seenUserIDs[id] continue branch.
// The branch fires when the loop encounters an id that was already marked as seen because a
// previously processed user had that value as their TelegramID.
//
// Setup: create user A (telegram_id=T). Then create user B whose internal id equals T
// (achieved by inserting with OVERRIDING SYSTEM VALUE). Add word_request_history for both
// A.ID and B.ID. When getUsersForWord processes A.ID first it marks seenUserIDs[T]=true;
// then processing B.ID (==T) hits the seenUserIDs[id] continue.
func TestTrainingWorker_getUsersForWord_TelegramIDSeenSkipped(t *testing.T) {
	worker, wordRepo, _, _, _, db, cleanup := newTrainingWorker(t, nil)
	defer cleanup()

	conn := db.GetConnection()

	// Create user A with a large telegram_id that we will reuse as user B's internal id.
	const telegramIDA int64 = 98765
	var userAID int64
	if err := conn.QueryRow(`INSERT INTO users (telegram_id) VALUES ($1) RETURNING id`, telegramIDA).Scan(&userAID); err != nil {
		t.Fatalf("insert user A: %v", err)
	}

	// Create user B whose internal id == telegramIDA (user A's telegram_id).
	// Using OVERRIDING SYSTEM VALUE lets us supply the id explicitly.
	if _, err := conn.Exec(
		`INSERT INTO users (id, telegram_id) OVERRIDING SYSTEM VALUE VALUES ($1, $2)`,
		telegramIDA, telegramIDA+1,
	); err != nil {
		t.Fatalf("insert user B with id=%d: %v", telegramIDA, err)
	}

	word := "seenword"
	// Entry for user A (internal id)
	if err := wordRepo.AddWordRequestHistoryWithCard(userAID, word, nil, &word); err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard(A): %v", err)
	}
	// Entry for user B (internal id == user A's telegram_id)
	if err := wordRepo.AddWordRequestHistoryWithCard(telegramIDA, word, nil, &word); err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard(B): %v", err)
	}

	users, err := worker.getUsersForWord(word)
	if err != nil {
		t.Fatalf("getUsersForWord: %v", err)
	}
	// User B's id == user A's telegram_id, so after processing user A the loop marks
	// seenUserIDs[telegramIDA]=true; when user B's id is processed it hits the continue.
	// Result: only user A is returned.
	if len(users) != 1 {
		t.Errorf("expected 1 user (B skipped as duplicate of A's telegram_id), got %d", len(users))
	}
}

// TestTrainingWorker_processCard_SensePOSFromWordCard covers the branch where sense.POS is empty
// but wordCard.POS is set (pos = *wordCard.POS).
func TestTrainingWorker_processCard_SensePOSFromWordCard(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		// sense.POS is empty string, but wordCard.POS = "noun"
		content := `{"word_en":"postest","lemma":"postest","transcription":"","senses":[{"pos":"","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["яблоко","груша"],"distractors_en":["orange","banana"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, trainingCardRepo, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	pos, trans, def := "noun", "pɒs", "позиция"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word: "postest", Definition: "", POS: &pos, Transcription: &trans, DefinitionRU: &def,
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard: %v", err)
	}

	cards, err := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 training card, got %d", len(cards))
	}
	if cards[0].POS == nil || *cards[0].POS != "noun" {
		t.Errorf("expected training card POS 'noun' from wordCard, got %v", cards[0].POS)
	}
}

// TestTrainingWorker_processCard_HighModelMarkProcessedErrorFails covers MarkWordCardProcessedError
// error path when high model rejects the word.
func TestTrainingWorker_processCard_HighModelMarkProcessedErrorFails(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return nil, fmt.Errorf("skip fill")
		}
		if callCount == 2 {
			// Default model: invalid (Cyrillic in distractors_en)
			content := `{"word_en":"hmerr","lemma":"hmerr","transcription":"","senses":[{"pos":"verb","word_ru":"тест","meaning_en":"test","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["тест","to run"],"hint":""}]}`
			resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
			return newJSONHTTPResponseTW(http.StatusOK, resp), nil
		}
		// High model: rejects word
		content := `{"error": "not valid", "word_en": "hmerr"}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	worker.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "hmerr", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	// Replace wordRepo with invalid one so MarkWordCardProcessedError fails
	logger, _ := zap.NewDevelopment()
	worker.wordRepo = newInvalidWordRepo(t, logger)

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should return nil even when high model MarkWordCardProcessedError fails: %v", err)
	}
}
