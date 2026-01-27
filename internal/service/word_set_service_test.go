package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"
	"unsafe"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/database"
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
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

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

	cleanup := func() {
		_ = db.Close()
	}

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

func TestEnsureUserCardsForWord(t *testing.T) {
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"apple","lemma":"apple","transcription":"","senses":[{"pos":"noun","word_ru":"яблоко","meaning_en":"apple","example_en":"","example_ru":"","distractors_ru":["груша","слива"],"distractors_en":["orange","banana"],"hint":""}]}`}}},
		}
		return newJSONHTTPResponse(http.StatusOK, resp), nil
	})

	svc, wordRepo, trainingCardRepo, _, _, _, db, cleanup := newWordSetService(t, transport)
	defer cleanup()

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
