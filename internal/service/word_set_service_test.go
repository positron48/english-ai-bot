package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newJSONHTTPResponse(status int, payload any) *http.Response {
	body, _ := json.Marshal(payload)
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func newWordSetService(t *testing.T, transport http.RoundTripper) (*WordSetService, *repository.WordRepository, *repository.TrainingCardRepository, *repository.UserCardRepository, *repository.UserWordKnowledgeRepository, *repository.WordSetRepository, *database.DB, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	userWordKnowledgeRepo := repository.NewUserWordKnowledgeRepository(db.GetConnection(), logger)
	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), logger)

	var aiService *ai.Service
	if transport != nil {
		aiService = ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
		aiService.SetTrainingPrompt("Generate card for: ")
		setAITransport(aiService, transport)
	}

	svc := NewWordSetService(
		wordSetRepo,
		wordSetCategoryRepo,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userWordKnowledgeRepo,
		aiService,
		"",
		logger,
	)

	cleanup := func() {} // shared db, do not close

	return svc, wordRepo, trainingCardRepo, userCardRepo, userWordKnowledgeRepo, wordSetRepo, db, cleanup
}

func setAITransport(service *ai.Service, transport http.RoundTripper) {
	client := &http.Client{Transport: transport}
	v := reflect.ValueOf(service).Elem().FieldByName("client")
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(client))
}

func TestEnsureWordCardExistsMinimal(t *testing.T) {
	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, nil)
	defer cleanup()

	id, err := svc.EnsureWordCardExistsMinimal("  Apple ")
	if err != nil {
		t.Fatalf("EnsureWordCardExistsMinimal error: %v", err)
	}

	card, err := wordRepo.GetWordCardByID(id)
	if err != nil {
		t.Fatalf("GetWordCardByID error: %v", err)
	}
	if card == nil || card.Word != "apple" {
		t.Fatalf("expected normalized word card")
	}
}

func TestEnsureWordCardExists_AI_JSON(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"run","lemma":"run","pos":"verb","transcription":"rʌn","definition_ru":"бежать","examples":[],"verb_forms":{"v1":"run"}}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	id, err := svc.EnsureWordCardExists(context.Background(), "Run")
	if err != nil {
		t.Fatalf("EnsureWordCardExists error: %v", err)
	}

	card, err := wordRepo.GetWordCardByID(id)
	if err != nil {
		t.Fatalf("GetWordCardByID error: %v", err)
	}
	if card == nil || card.DisplayEN == nil || *card.DisplayEN != "to run" {
		t.Fatalf("expected display EN to be 'to run'")
	}
}

func TestEnsureWordCardExists_AI_NonJSON(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: "legacy definition"}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	id, err := svc.EnsureWordCardExists(context.Background(), "Apple")
	if err != nil {
		t.Fatalf("EnsureWordCardExists error: %v", err)
	}

	card, err := wordRepo.GetWordCardByID(id)
	if err != nil {
		t.Fatalf("GetWordCardByID error: %v", err)
	}
	if card == nil {
		t.Fatalf("expected word card")
	}
}

func TestEnsureWordCardExists_ExistingCard(t *testing.T) {
	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, nil)
	defer cleanup()

	// Pre-create word card
	if err := wordRepo.SaveWordCard("existing", "definition"); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}
	existing, err := wordRepo.GetWordCardByLemma("existing")
	if err != nil || existing == nil {
		t.Fatalf("GetWordCardByLemma: %v", err)
	}

	id, err := svc.EnsureWordCardExists(context.Background(), "  EXISTING  ")
	if err != nil {
		t.Fatalf("EnsureWordCardExists error: %v", err)
	}
	if id != existing.ID {
		t.Errorf("expected existing card id %d, got %d", existing.ID, id)
	}
}

func TestEnsureWordCardExists_AINil(t *testing.T) {
	svc, _, _, _, _, _, _, cleanup := newWordSetService(t, nil)
	defer cleanup()

	_, err := svc.EnsureWordCardExists(context.Background(), "newword")
	if err == nil {
		t.Fatal("expected error when AI service is nil")
	}
	if !strings.Contains(err.Error(), "AI service not available") {
		t.Errorf("expected AI not available error, got: %v", err)
	}
}

func TestEnsureWordCardExists_LLMRejectsWord(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"error": "not a valid English word", "lemma": ""}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, _, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	_, err := svc.EnsureWordCardExists(context.Background(), "xyz")
	if err == nil {
		t.Fatal("expected error when LLM rejects word")
	}
	if !strings.Contains(err.Error(), "rejected by LLM") {
		t.Errorf("expected LLM reject error, got: %v", err)
	}
}

func TestEnsureWordCardExists_AIError(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})

	svc, _, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	// Word not in DB so AI is called; transport returns error
	_, err := svc.EnsureWordCardExists(context.Background(), "newword")
	if err == nil {
		t.Fatal("expected error when AI returns error")
	}
	if !strings.Contains(err.Error(), "failed to get AI response") && !strings.Contains(err.Error(), "AI") {
		t.Errorf("expected AI-related error, got: %v", err)
	}
}

func TestEnsureTrainingCardsExist(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"apple","lemma":"apple","transcription":"","senses":[{"pos":"noun","word_ru":"яблоко","meaning_en":"apple","example_en":"","example_ru":"","distractors_ru":["груша","слива"],"distractors_en":["orange","banana"],"hint":""}]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, wordRepo, trainingCardRepo, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	pos := "noun"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}

	if err := svc.EnsureTrainingCardsExist(context.Background(), cardID); err != nil {
		t.Fatalf("EnsureTrainingCardsExist error: %v", err)
	}

	cards, err := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID error: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 training card, got %d", len(cards))
	}

	// Second call should no-op
	if err := svc.EnsureTrainingCardsExist(context.Background(), cardID); err != nil {
		t.Fatalf("EnsureTrainingCardsExist second call error: %v", err)
	}
}

func TestEnsureTrainingCardsExist_WordCardNotFound(t *testing.T) {
	svc, _, _, _, _, _, _, cleanup := newWordSetService(t, nil)
	defer cleanup()

	err := svc.EnsureTrainingCardsExist(context.Background(), 999999)
	if err == nil {
		t.Fatal("expected error for non-existent word card")
	}
	if !strings.Contains(err.Error(), "word card not found") {
		t.Errorf("expected word card not found error, got: %v", err)
	}
}

func TestEnsureTrainingCardsExist_AINil(t *testing.T) {
	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, nil)
	defer cleanup()

	pos := "noun"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	// Create training cards so we don't hit "already exist" path - actually no, we need no training cards so it tries to generate. So no training cards. Then GetWordCardByID returns the card. Then aiService == nil.
	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when AI service is nil")
	}
	if !strings.Contains(err.Error(), "AI service not available") {
		t.Errorf("expected AI not available error, got: %v", err)
	}
}

func TestEnsureTrainingCardsExist_NoSenses(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"apple","lemma":"apple","transcription":"","senses":[]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	pos := "noun"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when senses empty")
	}
	if !strings.Contains(err.Error(), "no senses") {
		t.Errorf("expected no senses error, got: %v", err)
	}
}

func TestEnsureTrainingCardsExist_LLMError(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"error": "not a valid word", "word_en":"x","senses":[]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()

	pos := "noun"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "x", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected error when LLM returns error")
	}
	if !strings.Contains(err.Error(), "rejected by LLM") {
		t.Errorf("expected LLM reject error, got: %v", err)
	}
}

func TestEnsureTrainingCardsExist_ValidationFailNoHighModel(t *testing.T) {
	// distractors_en with Cyrillic triggers validation error; no modelHigh so return original error
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"run","lemma":"run","transcription":"","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"run","example_en":"","example_ru":"","distractors_ru":["идти","плыть"],"distractors_en":["бежать","to walk"],"hint":""}]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()
	// NewWordSetService sets modelHigh to "" - so no high model
	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected validation failed error, got: %v", err)
	}
}

func TestEnsureTrainingCardsExist_ValidationFailHighModelSucceeds(t *testing.T) {
	callCount := 0
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// First call: default model returns invalid (Cyrillic in distractors_en)
			return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
				Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"jump","lemma":"jump","transcription":"","senses":[{"pos":"verb","word_ru":"прыгать","meaning_en":"jump","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["прыгать","to run"],"hint":""}]}`}}},
			}), nil
		}
		// Second call: high model returns valid
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"jump","lemma":"jump","transcription":"dʒʌmp","senses":[{"pos":"verb","display_word":"to jump","word_ru":"прыгать","meaning_en":"jump","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["to run","to walk"],"hint":""}]}`}}},
		}), nil
	})

	svc, wordRepo, trainingCardRepo, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()
	// Set modelHigh so that retry with high model is attempted
	svc.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "jump", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	if err := svc.EnsureTrainingCardsExist(context.Background(), cardID); err != nil {
		t.Fatalf("EnsureTrainingCardsExist: %v", err)
	}

	cards, err := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 training card from high model, got %d", len(cards))
	}
}

func TestEnsureTrainingCardsExist_ValidationFailHighModelAlsoFails(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		// Both calls return invalid (Cyrillic in distractors_en)
		content := `{"word_en":"run","lemma":"run","transcription":"","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"run","example_en":"","example_ru":"","distractors_ru":["идти","плыть"],"distractors_en":["бежать","to walk"],"hint":""}]}`
		return newJSONHTTPResponse(http.StatusOK, ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: content}}},
		}), nil
	})

	svc, wordRepo, _, _, _, _, _, cleanup := newWordSetService(t, transport)
	defer cleanup()
	svc.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = svc.EnsureTrainingCardsExist(context.Background(), cardID)
	if err == nil {
		t.Fatal("expected validation error when high model also fails")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected validation failed error, got: %v", err)
	}
}

func TestEnsureUserCardsForWord(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"apple","lemma":"apple","transcription":"","senses":[{"pos":"noun","word_ru":"яблоко","meaning_en":"apple","example_en":"","example_ru":"","distractors_ru":["груша","слива"],"distractors_en":["orange","banana"],"hint":""}]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, wordRepo, trainingCardRepo, _, _, _, db, cleanup := newWordSetService(t, transport)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	_, _ = userRepo.GetOrCreateUser(1) // FK user_cards_user_id_fkey

	pos := "noun"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}

	if err := svc.EnsureTrainingCardsExist(context.Background(), cardID); err != nil {
		t.Fatalf("EnsureTrainingCardsExist error: %v", err)
	}

	trainingCards, _ := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if len(trainingCards) == 0 {
		t.Fatalf("expected training cards")
	}

	if err := svc.EnsureUserCardsForWord(1, cardID); err != nil {
		t.Fatalf("EnsureUserCardsForWord error: %v", err)
	}

	var count int
	if err := db.GetConnection().QueryRow(
		`SELECT COUNT(*) FROM user_cards uc 
		 INNER JOIN training_cards tc ON uc.training_card_id = tc.id 
		 WHERE uc.user_id = ? AND tc.word_card_id = ?`,
		1, cardID,
	).Scan(&count); err != nil {
		t.Fatalf("count user cards error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 user cards, got %d", count)
	}
}
