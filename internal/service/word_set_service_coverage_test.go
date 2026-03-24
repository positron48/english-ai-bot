package service

// Additional tests to achieve 100% coverage for internal/service/word_set_service.go

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

// mockWordRepo is a mock implementation of wordRepoForWordSet for testing error paths.
type mockWordRepo struct {
	getWordCardByLemmaFn  func(lemma string) (*models.WordCard, error)
	getWordCardByIDFn     func(id int64) (*models.WordCard, error)
	saveWordCardFn        func(word, content string) error
	upsertWordCardLemmaFn func(card *models.WordCard) (int64, error)
	upsertWordFormMapping func(form string, wordCardID int64) error
}

func (m *mockWordRepo) GetWordCardByLemma(lemma string) (*models.WordCard, error) {
	if m.getWordCardByLemmaFn != nil {
		return m.getWordCardByLemmaFn(lemma)
	}
	return nil, nil
}

func (m *mockWordRepo) GetWordCardByID(id int64) (*models.WordCard, error) {
	if m.getWordCardByIDFn != nil {
		return m.getWordCardByIDFn(id)
	}
	return nil, nil
}

func (m *mockWordRepo) SaveWordCard(word, content string) error {
	if m.saveWordCardFn != nil {
		return m.saveWordCardFn(word, content)
	}
	return nil
}

func (m *mockWordRepo) UpsertWordCardLemma(card *models.WordCard) (int64, error) {
	if m.upsertWordCardLemmaFn != nil {
		return m.upsertWordCardLemmaFn(card)
	}
	return 1, nil
}

func (m *mockWordRepo) UpsertWordFormMapping(form string, wordCardID int64) error {
	if m.upsertWordFormMapping != nil {
		return m.upsertWordFormMapping(form, wordCardID)
	}
	return nil
}

// setupWordSetServiceSecondDB creates a second isolated Postgres container for DB error tests.
func setupWordSetServiceSecondDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	dsn := testutil.SecondPostgresDSN(t)
	var dbWrap *database.DB
	var err error
	dbWrap, err = database.NewWithConfig("postgres", "", dsn, logger)
	if dbWrap == nil {
		t.Skipf("second DB not available: %v", err)
	}
	cleanup := func() { _ = dbWrap.GetConnection().Close() }
	return dbWrap, cleanup
}

// ── EnsureWordCardExistsMinimal ───────────────────────────────────────────────

// TestEnsureWordCardExistsMinimal_GetWordCardByLemmaError covers the GetWordCardByLemma error path.
func TestEnsureWordCardExistsMinimal_GetWordCardByLemmaError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Drop word_cards to make GetWordCardByLemma fail
	_, err := conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE word_cards CASCADE`)
	if err != nil {
		t.Skipf("cannot drop word_cards: %v", err)
	}

	_, err = svc.EnsureWordCardExistsMinimal("apple")
	if err == nil {
		t.Fatal("expected error when word_cards table is dropped")
	}
}

// TestEnsureWordCardExistsMinimal_UpsertError covers the UpsertWordCardLemma error path.
func TestEnsureWordCardExistsMinimal_UpsertError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Add trigger that blocks INSERT on word_cards
	_, err := conn.Exec(`
		CREATE OR REPLACE FUNCTION block_wc_insert() RETURNS TRIGGER AS $$
		BEGIN
			RAISE EXCEPTION 'Insert blocked for testing';
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER block_word_cards_insert
		BEFORE INSERT ON word_cards
		FOR EACH ROW EXECUTE FUNCTION block_wc_insert();
	`)
	if err != nil {
		t.Skipf("cannot create trigger: %v", err)
	}

	_, err = svc.EnsureWordCardExistsMinimal("newword")
	if err == nil {
		t.Fatal("expected error when INSERT on word_cards is blocked")
	}
}

// ── EnsureWordCardExists ──────────────────────────────────────────────────────

// TestEnsureWordCardExists_NonJSON_SaveError covers the SaveWordCard error path (non-JSON response).
func TestEnsureWordCardExists_NonJSON_SaveError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns non-JSON
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: "legacy definition"}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	// Add trigger that blocks INSERT on word_cards to make SaveWordCard fail
	_, err := conn.Exec(`
		CREATE OR REPLACE FUNCTION block_wc_insert2() RETURNS TRIGGER AS $$
		BEGIN
			RAISE EXCEPTION 'Insert blocked for testing';
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER block_word_cards_insert2
		BEFORE INSERT ON word_cards
		FOR EACH ROW EXECUTE FUNCTION block_wc_insert2();
	`)
	if err != nil {
		t.Skipf("cannot create trigger: %v", err)
	}

	_, err = svc.EnsureWordCardExists(context.Background(), "newword")
	if err == nil {
		t.Fatal("expected error when SaveWordCard fails")
	}
}

// TestEnsureWordCardExists_NonJSON_GetAfterSaveError covers the GetWordCardByLemma after save error path.
// We need SaveWordCard to succeed but GetWordCardByLemma to fail after.
// This is hard to do without mocking; instead we test the "word card not found after save" path
// by having SaveWordCard succeed but then the word not being found.
// The simplest approach: use a second DB where we can manipulate the table.
func TestEnsureWordCardExists_NonJSON_WordCardNilAfterSave(t *testing.T) {
	// This path requires SaveWordCard to succeed but GetWordCardByLemma to return nil.
	// In practice this shouldn't happen, but we can simulate it with a trigger that
	// deletes the row after insert.
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns non-JSON (legacy path)
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: "legacy definition"}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	// Add trigger that deletes the row after insert (so GetWordCardByLemma returns nil)
	_, err := conn.Exec(`
		CREATE OR REPLACE FUNCTION delete_after_wc_insert() RETURNS TRIGGER AS $$
		BEGIN
			DELETE FROM word_cards WHERE id = NEW.id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER delete_word_card_after_insert
		AFTER INSERT ON word_cards
		FOR EACH ROW EXECUTE FUNCTION delete_after_wc_insert();
	`)
	if err != nil {
		t.Skipf("cannot create trigger: %v", err)
	}

	_, err = svc.EnsureWordCardExists(context.Background(), "vanishword")
	if err == nil {
		t.Fatal("expected error when word card not found after save")
	}
}

// TestEnsureWordCardExists_JSON_UpsertError covers the UpsertWordCardLemma error path (JSON response).
func TestEnsureWordCardExists_JSON_UpsertError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns valid JSON
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"apple","lemma":"apple","pos":"noun","transcription":"ˈæpəl","definition_ru":"яблоко","examples":[]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	// Add trigger that blocks INSERT on word_cards
	_, err := conn.Exec(`
		CREATE OR REPLACE FUNCTION block_wc_insert3() RETURNS TRIGGER AS $$
		BEGIN
			RAISE EXCEPTION 'Insert blocked for testing';
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER block_word_cards_insert3
		BEFORE INSERT ON word_cards
		FOR EACH ROW EXECUTE FUNCTION block_wc_insert3();
	`)
	if err != nil {
		t.Skipf("cannot create trigger: %v", err)
	}

	_, err = svc.EnsureWordCardExists(context.Background(), "apple")
	if err == nil {
		t.Fatal("expected error when UpsertWordCardLemma fails")
	}
}

// TestEnsureWordCardExists_JSON_GetByIDError covers the GetWordCardByID error path after upsert.
func TestEnsureWordCardExists_JSON_GetByIDError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns valid JSON
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"apple","lemma":"apple","pos":"noun","transcription":"ˈæpəl","definition_ru":"яблоко","examples":[]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	// Add trigger that deletes the row after insert (so GetWordCardByID returns nil)
	_, err := conn.Exec(`
		CREATE OR REPLACE FUNCTION delete_after_wc_insert2() RETURNS TRIGGER AS $$
		BEGIN
			DELETE FROM word_cards WHERE id = NEW.id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER delete_word_card_after_insert2
		AFTER INSERT ON word_cards
		FOR EACH ROW EXECUTE FUNCTION delete_after_wc_insert2();
	`)
	if err != nil {
		t.Skipf("cannot create trigger: %v", err)
	}

	_, err = svc.EnsureWordCardExists(context.Background(), "apple")
	if err == nil {
		t.Fatal("expected error when word card not found after upsert")
	}
}

// TestEnsureWordCardExists_JSON_WithWordFormMapping covers the UpsertWordFormMapping path
// where normalizedWord != lemma (e.g. "running" -> lemma "run").
func TestEnsureWordCardExists_JSON_WithWordFormMapping(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns lemma "run" for input "running" (normalizedWord != lemma)
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"running","lemma":"run","pos":"verb","transcription":"rʌn","definition_ru":"бежать","examples":[],"verb_forms":{"v1":"run"}}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	id, err := svc.EnsureWordCardExists(context.Background(), "running")
	if err != nil {
		t.Fatalf("EnsureWordCardExists error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero word card ID")
	}
}

// TestEnsureWordCardExists_JSON_NoExamples covers the path where examples is empty (examplesJSON nil).
func TestEnsureWordCardExists_JSON_NoExamples(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns no examples and no verb_forms
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"cat","lemma":"cat","pos":"noun","transcription":"kæt","definition_ru":"кошка","examples":[]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	id, err := svc.EnsureWordCardExists(context.Background(), "cat")
	if err != nil {
		t.Fatalf("EnsureWordCardExists error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero word card ID")
	}
}

// ── EnsureTrainingCardsExist ──────────────────────────────────────────────────

// TestEnsureTrainingCardsExist_ValidationFail_HighModel_AIError covers the path where
// validation fails, high model is available, but high model AI call returns error.
func TestEnsureTrainingCardsExist_ValidationFail_HighModel_AIError(t *testing.T) {
	callCount := 0
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// First call: default model returns invalid (Cyrillic in distractors_en)
			return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
				Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","lemma":"run","transcription":"","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"move fast on foot","example_en":"","example_ru":"","distractors_ru":["идти","плыть"],"distractors_en":["бежать","to walk"],"hint":""}]}`}}},
			}), nil
		}
		// Second call: high model returns error
		return nil, context.DeadlineExceeded
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()
	svc.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run2", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when high model AI call fails")
	}
}

// TestEnsureTrainingCardsExist_ValidationFail_HighModel_ParseError covers the path where
// validation fails, high model is available, but high model response is invalid JSON.
func TestEnsureTrainingCardsExist_ValidationFail_HighModel_ParseError(t *testing.T) {
	callCount := 0
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// First call: default model returns invalid (Cyrillic in distractors_en)
			return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
				Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","lemma":"run","transcription":"","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"move fast on foot","example_en":"","example_ru":"","distractors_ru":["идти","плыть"],"distractors_en":["бежать","to walk"],"hint":""}]}`}}},
			}), nil
		}
		// Second call: high model returns invalid JSON
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `not valid json`}}},
		}), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()
	svc.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run3", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when high model returns invalid JSON")
	}
}

// TestEnsureTrainingCardsExist_ValidationFail_HighModel_LLMError covers the path where
// validation fails, high model is available, but high model LLM returns error field.
func TestEnsureTrainingCardsExist_ValidationFail_HighModel_LLMError(t *testing.T) {
	callCount := 0
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// First call: default model returns invalid (Cyrillic in distractors_en)
			return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
				Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","lemma":"run","transcription":"","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"move fast on foot","example_en":"","example_ru":"","distractors_ru":["идти","плыть"],"distractors_en":["бежать","to walk"],"hint":""}]}`}}},
			}), nil
		}
		// Second call: high model returns LLM error
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"error":"word rejected","word_en":"run","senses":[]}`}}},
		}), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()
	svc.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run4", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when high model LLM returns error")
	}
}

// TestEnsureTrainingCardsExist_ValidationFail_HighModel_NoSenses covers the path where
// validation fails, high model is available, but high model response has no senses.
func TestEnsureTrainingCardsExist_ValidationFail_HighModel_NoSenses(t *testing.T) {
	callCount := 0
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// First call: default model returns invalid (Cyrillic in distractors_en)
			return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
				Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","lemma":"run","transcription":"","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"move fast on foot","example_en":"","example_ru":"","distractors_ru":["идти","плыть"],"distractors_en":["бежать","to walk"],"hint":""}]}`}}},
			}), nil
		}
		// Second call: high model returns empty senses
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","lemma":"run","transcription":"","senses":[]}`}}},
		}), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()
	svc.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run5", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when high model returns no senses")
	}
}

// TestEnsureTrainingCardsExist_AIError covers the AI call error path.
func TestEnsureTrainingCardsExist_AIError(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	pos := "noun"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "aierr", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when AI call fails")
	}
}

// TestEnsureTrainingCardsExist_ParseError covers the JSON parse error path.
func TestEnsureTrainingCardsExist_ParseError(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `not valid json`}}},
		}), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	pos := "noun"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "parseerr", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when AI response is invalid JSON")
	}
}

// TestEnsureTrainingCardsExist_WithDisplayWordAndPOS covers the training card creation
// with display_word and POS set from sense (not from wordCard.POS).
func TestEnsureTrainingCardsExist_WithDisplayWordAndPOS(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","transcription":"rʌn","senses":[{"pos":"verb","display_word":"to run","word_ru":"бежать","meaning_en":"move fast on foot","example_en":"I run","example_ru":"Я бегу","distractors_ru":["идти","плыть"],"distractors_en":["to walk","to swim"],"hint":"motion"}]}`}}},
		}), nil
	})

	svc, wordRepo, trainingCardRepo, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	// Word card with no POS (so pos comes from sense)
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run6", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err != nil {
		t.Fatalf("EnsureTrainingCardsExist: %v", err)
	}

	cards, err := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 training card, got %d", len(cards))
	}
	if cards[0].DisplayWord == nil || *cards[0].DisplayWord != "to run" {
		t.Errorf("expected display_word 'to run', got %v", cards[0].DisplayWord)
	}
}

// ── EnsureUserCardsForWord ────────────────────────────────────────────────────

// TestEnsureUserCardsForWord_CreateUserCardFails covers the warn path when CreateUserCard fails.
func TestEnsureUserCardsForWord_CreateUserCardFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)
	userRepo := repository.NewUserRepository(conn, logger)

	user, _ := userRepo.GetOrCreateUser(55555)
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "failcard", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_, err = tcRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: cardID,
		WordEN:     "failcard",
		WordRU:     "неудача",
		MeaningEN:  "fail",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	// Mock user card repo that always fails on CreateUserCard
	mockUC := &mockUserCardRepoForWordSet{
		createFunc: func(card *models.UserCard) (int64, error) {
			return 0, fmt.Errorf("mock create error")
		},
	}

	svc := NewWordSetServiceWithMastering(
		wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, mockUC, uwkRepo, nil, nil, config.DefaultLearningConfig(), "", logger,
	)

	// Should not return error (warn path)
	err = svc.EnsureUserCardsForWord(user.ID, cardID)
	if err != nil {
		t.Errorf("EnsureUserCardsForWord should return nil when CreateUserCard fails (warn path), got: %v", err)
	}
}

// TestEnsureUserCardsForWord_GetTrainingCardsError covers the GetTrainingCards error path.
func TestEnsureUserCardsForWord_GetTrainingCardsError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Drop training_cards to make GetTrainingCardsByWordCardID fail
	_, err := conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE training_cards CASCADE`)
	if err != nil {
		t.Skipf("cannot drop training_cards: %v", err)
	}

	err = svc.EnsureUserCardsForWord(1, 1)
	if err == nil {
		t.Fatal("expected error when training_cards table is dropped")
	}
}

// ── MarkKnown ─────────────────────────────────────────────────────────────────

// TestMarkKnown_MarkKnownError covers the MarkKnown error path (userWordKnowledgeRepo.MarkKnown fails).
func TestMarkKnown_MarkKnownError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Drop user_word_knowledge to make MarkKnown fail
	_, err := conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE user_word_knowledge CASCADE`)
	if err != nil {
		t.Skipf("cannot drop user_word_knowledge: %v", err)
	}

	err = svc.MarkKnown(1, 1)
	if err == nil {
		t.Fatal("expected error when user_word_knowledge table is dropped")
	}
}

// ── ProcessWordSetItems ───────────────────────────────────────────────────────

// TestProcessWordSetItems_SetWordSetItemsError covers the SetWordSetItems error path.
func TestProcessWordSetItems_SetWordSetItemsError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Create a word set first (before dropping word_set_items)
	wordSetID, err := wordSetRepo.CreateWordSet(&models.WordSet{Title: "error set"})
	if err != nil {
		t.Fatalf("CreateWordSet: %v", err)
	}

	// Drop word_set_items to make SetWordSetItems fail
	_, err = conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE word_set_items CASCADE`)
	if err != nil {
		t.Skipf("cannot drop word_set_items: %v", err)
	}

	err = svc.ProcessWordSetItems(context.Background(), wordSetID, "apple,banana")
	if err == nil {
		t.Fatal("expected error when word_set_items table is dropped")
	}
}

// TestProcessWordSetItems_EnsureMinimalFails covers the warn path when EnsureWordCardExistsMinimal fails.
func TestProcessWordSetItems_EnsureMinimalFails(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Create a word set first
	wordSetID, err := wordSetRepo.CreateWordSet(&models.WordSet{Title: "minimal fail set"})
	if err != nil {
		t.Fatalf("CreateWordSet: %v", err)
	}

	// Add trigger that blocks INSERT on word_cards to make EnsureWordCardExistsMinimal fail
	_, err = conn.Exec(`
		CREATE OR REPLACE FUNCTION block_wc_insert4() RETURNS TRIGGER AS $$
		BEGIN
			RAISE EXCEPTION 'Insert blocked for testing';
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER block_word_cards_insert4
		BEFORE INSERT ON word_cards
		FOR EACH ROW EXECUTE FUNCTION block_wc_insert4();
	`)
	if err != nil {
		t.Skipf("cannot create trigger: %v", err)
	}

	// ProcessWordSetItems should warn (not fail) when EnsureWordCardExistsMinimal fails
	// but SetWordSetItems should succeed (with empty list)
	err = svc.ProcessWordSetItems(context.Background(), wordSetID, "apple,banana")
	if err != nil {
		t.Errorf("ProcessWordSetItems should succeed (warn path for failed words), got: %v", err)
	}
}

// ── EnsureWordCardExists additional coverage ──────────────────────────────────

// TestEnsureWordCardExists_GetWordCardByLemmaError covers the GetWordCardByLemma error path
// at the start of EnsureWordCardExists (line 123).
func TestEnsureWordCardExists_GetWordCardByLemmaError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Drop word_cards to make GetWordCardByLemma fail
	_, err := conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE word_cards CASCADE`)
	if err != nil {
		t.Skipf("cannot drop word_cards: %v", err)
	}

	_, err = svc.EnsureWordCardExists(context.Background(), "apple")
	if err == nil {
		t.Fatal("expected error when word_cards table is dropped")
	}
}

// TestEnsureWordCardExists_NonJSON_GetWordCardByLemmaError covers the GetWordCardByLemma error
// after SaveWordCard succeeds (line 150).
func TestEnsureWordCardExists_NonJSON_GetWordCardByLemmaError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns non-JSON (legacy format)
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: "legacy definition text"}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	// Add trigger that drops word_cards AFTER INSERT (so SaveWordCard succeeds but GetWordCardByLemma fails)
	_, err := conn.Exec(`
		CREATE OR REPLACE FUNCTION drop_wc_after_insert_fn() RETURNS TRIGGER AS $$
		BEGIN
			EXECUTE 'DROP TABLE IF EXISTS word_cards CASCADE';
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER drop_wc_after_insert
		AFTER INSERT ON word_cards
		FOR EACH ROW EXECUTE FUNCTION drop_wc_after_insert_fn();
	`)
	if err != nil {
		t.Skipf("cannot create trigger: %v", err)
	}

	_, err = svc.EnsureWordCardExists(context.Background(), "testword")
	if err == nil {
		t.Fatal("expected error when GetWordCardByLemma fails after SaveWordCard")
	}
}

// TestEnsureWordCardExists_EmptyLemma covers the path where AI returns empty lemma (line 166).
func TestEnsureWordCardExists_EmptyLemma(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns valid JSON with empty lemma
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"apple","lemma":"","pos":"noun","transcription":"ˈæpəl","definition_ru":"яблоко","examples":[]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	id, err := svc.EnsureWordCardExists(context.Background(), "apple")
	if err != nil {
		t.Fatalf("EnsureWordCardExists error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero word card ID")
	}
}

// TestEnsureWordCardExists_WithExamples covers the path where AI returns non-empty examples (line 177).
func TestEnsureWordCardExists_WithExamples(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns valid JSON with non-empty examples (correct WordInfoExample format)
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"dog","lemma":"dog","pos":"noun","transcription":"dɒɡ","definition_ru":"собака","examples":[{"example_en":"The dog barked.","gloss_ru":"Собака лаяла."}]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	id, err := svc.EnsureWordCardExists(context.Background(), "dog")
	if err != nil {
		t.Fatalf("EnsureWordCardExists error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero word card ID")
	}
}

// ── EnsureTrainingCardsExist additional coverage ──────────────────────────────

// TestEnsureTrainingCardsExist_GetTrainingCardsError covers the GetTrainingCardsByWordCardID error path.
func TestEnsureTrainingCardsExist_GetTrainingCardsError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Insert a word card first
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "errorword", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	// Drop training_cards to make GetTrainingCardsByWordCardID fail
	_, err = conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE training_cards CASCADE`)
	if err != nil {
		t.Skipf("cannot drop training_cards: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when training_cards table is dropped")
	}
}

// TestEnsureTrainingCardsExist_GetWordCardByIDError covers the GetWordCardByID error path.
func TestEnsureTrainingCardsExist_GetWordCardByIDError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Insert a word card first
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "errorword2", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	// Drop word_cards to make GetWordCardByID fail (after training_cards check passes)
	_, err = conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE word_cards CASCADE`)
	if err != nil {
		t.Skipf("cannot drop word_cards: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when word_cards table is dropped")
	}
}

// TestEnsureTrainingCardsExist_POSFromWordCard covers the pos == "" && wordCard.POS != nil branch (line 364).
func TestEnsureTrainingCardsExist_POSFromWordCard(t *testing.T) {
	// AI returns senses with empty POS, but wordCard has POS set
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"jump","transcription":"dʒʌmp","senses":[{"pos":"","display_word":"","word_ru":"прыгать","meaning_en":"to jump","example_en":"He jumps high","example_ru":"Он прыгает высоко","distractors_ru":["бежать","идти"],"distractors_en":["to run","to walk"],"hint":""}]}`}}},
		}), nil
	})

	svc, wordRepo, trainingCardRepo, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	// Create word card with POS set (so wordCard.POS != nil)
	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "jump_pos", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err != nil {
		t.Fatalf("EnsureTrainingCardsExist: %v", err)
	}

	cards, err := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 training card, got %d", len(cards))
	}
	// POS should be set from wordCard.POS (since sense.POS was "")
	if cards[0].POS == nil || *cards[0].POS != "verb" {
		t.Errorf("expected POS 'verb' from wordCard, got %v", cards[0].POS)
	}
}

// TestEnsureTrainingCardsExist_CreateTrainingCardError covers the CreateTrainingCard error path.
func TestEnsureTrainingCardsExist_CreateTrainingCardError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns valid training card response
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","transcription":"rʌn","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"to run","example_en":"I run","example_ru":"Я бегу","distractors_ru":["идти","плыть"],"distractors_en":["to walk","to swim"],"hint":""}]}`}}},
		}), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	// Insert a word card
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run_tc_err", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	// Drop training_cards to make CreateTrainingCard fail
	_, err = conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE training_cards CASCADE`)
	if err != nil {
		t.Skipf("cannot drop training_cards: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when training_cards table is dropped")
	}
}

// TestEnsureWordCardExists_JSON_UpsertWordFormMappingError covers the UpsertWordFormMapping warn path (line 221).
func TestEnsureWordCardExists_JSON_UpsertWordFormMappingError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns lemma "run" for input "running" (normalizedWord != lemma, so UpsertWordFormMapping is called)
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"running","lemma":"run","pos":"verb","transcription":"rʌn","definition_ru":"бежать","examples":[],"verb_forms":{"v1":"run"}}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	// Drop word_forms table to make UpsertWordFormMapping fail (warn path)
	_, err := conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE word_forms CASCADE`)
	if err != nil {
		t.Skipf("cannot drop word_forms: %v", err)
	}

	// Should succeed (warn path for UpsertWordFormMapping)
	id, err := svc.EnsureWordCardExists(context.Background(), "running")
	if err != nil {
		t.Fatalf("EnsureWordCardExists should succeed (warn path), got: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero word card ID")
	}
}

// TestEnsureTrainingCardsExist_CreateTrainingCardErrorWithTrigger covers the CreateTrainingCard error path
// using a trigger that blocks INSERT on training_cards.
func TestEnsureTrainingCardsExist_CreateTrainingCardErrorWithTrigger(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	// AI returns valid training card response
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","transcription":"rʌn","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"to run","example_en":"I run","example_ru":"Я бегу","distractors_ru":["идти","плыть"],"distractors_en":["to walk","to swim"],"hint":""}]}`}}},
		}), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, aiService, config.DefaultLearningConfig(), "", logger)

	// Insert a word card
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run_tc_trigger", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	// Add trigger that blocks INSERT on training_cards
	_, err = conn.Exec(`
		CREATE OR REPLACE FUNCTION block_tc_insert_fn() RETURNS TRIGGER AS $$
		BEGIN
			RAISE EXCEPTION 'Insert on training_cards blocked for testing';
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER block_tc_insert
		BEFORE INSERT ON training_cards
		FOR EACH ROW EXECUTE FUNCTION block_tc_insert_fn();
	`)
	if err != nil {
		t.Skipf("cannot create trigger: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when CreateTrainingCard fails")
	}
}

// TestEnsureUserCardsForWord_MasteringRepoUpsertError covers the userWordMasteringRepo.Upsert error path.
func TestEnsureUserCardsForWord_MasteringRepoUpsertError(t *testing.T) {
	dbWrap, cleanup := setupWordSetServiceSecondDB(t)
	defer cleanup()

	conn := dbWrap.GetConnection()
	logger, _ := zap.NewDevelopment()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)
	masteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	userRepo := repository.NewUserRepository(conn, logger)

	svc := NewWordSetServiceWithMastering(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, masteringRepo, nil, config.DefaultLearningConfig(), "", logger)

	// Create user and word card
	user, err := userRepo.GetOrCreateUser(77777)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "masterword", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	// Create a training card
	tc := &models.TrainingCard{
		WordCardID:    cardID,
		WordEN:        "masterword",
		WordRU:        "мастерслово",
		MeaningEN:     "a master word",
		SenseIndex:    0,
		DistractorsRU: "[]",
		DistractorsEN: "[]",
	}
	_, err = tcRepo.CreateTrainingCard(tc)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	// Drop user_word_mastering to make Upsert fail
	_, err = conn.Exec(`SET session_replication_role = replica`)
	if err != nil {
		t.Skipf("cannot disable FK: %v", err)
	}
	_, err = conn.Exec(`DROP TABLE user_word_mastering CASCADE`)
	if err != nil {
		t.Skipf("cannot drop user_word_mastering: %v", err)
	}

	// EnsureUserCardsForWord should warn (not fail) when Upsert fails
	err = svc.EnsureUserCardsForWord(user.ID, cardID)
	if err != nil {
		t.Errorf("EnsureUserCardsForWord should succeed (warn path), got: %v", err)
	}
}

// ── EnsureWordCardExists mock-based tests ─────────────────────────────────────

// TestEnsureWordCardExists_NonJSON_GetWordCardByLemmaError_Mock covers line 159-161:
// SaveWordCard succeeds but GetWordCardByLemma returns an error (using mock wordRepo).
func TestEnsureWordCardExists_NonJSON_GetWordCardByLemmaError_Mock(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	callCount := 0
	mock := &mockWordRepo{
		// First call (initial lookup) returns nil, nil (word not found)
		// Second call (after SaveWordCard) returns error
		getWordCardByLemmaFn: func(lemma string) (*models.WordCard, error) {
			callCount++
			if callCount == 1 {
				return nil, nil
			}
			return nil, errors.New("db error after save")
		},
		saveWordCardFn: func(word, content string) error {
			return nil
		},
	}

	// AI returns non-JSON response
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: "legacy definition text"}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(nil, nil, mock, nil, nil, nil, aiService, config.DefaultLearningConfig(), "", logger)

	_, err := svc.EnsureWordCardExists(context.Background(), "testword")
	if err == nil {
		t.Fatal("expected error when GetWordCardByLemma fails after SaveWordCard")
	}
}

// TestEnsureWordCardExists_JSON_GetWordCardByIDError covers line 221-223:
// UpsertWordCardLemma succeeds but GetWordCardByID returns an error.
func TestEnsureWordCardExists_JSON_GetWordCardByIDError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mock := &mockWordRepo{
		getWordCardByLemmaFn: func(lemma string) (*models.WordCard, error) {
			return nil, nil // word not found
		},
		upsertWordCardLemmaFn: func(card *models.WordCard) (int64, error) {
			return 42, nil // success
		},
		getWordCardByIDFn: func(id int64) (*models.WordCard, error) {
			return nil, errors.New("db error getting word card by id")
		},
	}

	// AI returns valid JSON
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"lemma":"testword","pos":"noun","transcription":"ˈtɛstwɜːd","definition_ru":"тестовое слово","senses":[]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(nil, nil, mock, nil, nil, nil, aiService, config.DefaultLearningConfig(), "", logger)

	_, err := svc.EnsureWordCardExists(context.Background(), "testword")
	if err == nil {
		t.Fatal("expected error when GetWordCardByID fails after UpsertWordCardLemma")
	}
}

// TestEnsureWordCardExists_JSON_GetWordCardByIDNil covers line 224-226:
// UpsertWordCardLemma succeeds but GetWordCardByID returns nil (no row found).
func TestEnsureWordCardExists_JSON_GetWordCardByIDNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mock := &mockWordRepo{
		getWordCardByLemmaFn: func(lemma string) (*models.WordCard, error) {
			return nil, nil // word not found
		},
		upsertWordCardLemmaFn: func(card *models.WordCard) (int64, error) {
			return 42, nil // success
		},
		getWordCardByIDFn: func(id int64) (*models.WordCard, error) {
			return nil, nil // no row found
		},
	}

	// AI returns valid JSON
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"lemma":"testword","pos":"noun","transcription":"ˈtɛstwɜːd","definition_ru":"тестовое слово","senses":[]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})
	aiService := ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
	aiService.SetTrainingPrompt("Generate card for: ")
	setAITransport(aiService, transport)

	svc := NewWordSetService(nil, nil, mock, nil, nil, nil, aiService, config.DefaultLearningConfig(), "", logger)

	_, err := svc.EnsureWordCardExists(context.Background(), "testword")
	if err == nil {
		t.Fatal("expected error when GetWordCardByID returns nil after UpsertWordCardLemma")
	}
}
