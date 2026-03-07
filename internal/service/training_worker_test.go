package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

type rtFuncTW func(*http.Request) (*http.Response, error)

func (f rtFuncTW) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newJSONHTTPResponseTW(status int, payload any) *http.Response {
	body, _ := json.Marshal(payload)
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func setAITransportTW(service *ai.Service, transport http.RoundTripper) {
	client := &http.Client{Transport: transport}
	v := reflect.ValueOf(service).Elem().FieldByName("client")
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(client))
}

func newTrainingWorker(t *testing.T, transport http.RoundTripper) (*TrainingWorker, *repository.WordRepository, *repository.TrainingCardRepository, *repository.UserCardRepository, *repository.UserRepository, *database.DB, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)

	var aiService *ai.Service
	if transport != nil {
		aiService = ai.NewService("http://example.com", "test-model", "test-key", "prompt", logger)
		aiService.SetTrainingPrompt("Generate card for: ")
		setAITransportTW(aiService, transport)
	}

	worker := NewTrainingWorker(
		aiService,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userRepo,
		nil,
		nil,
		nil,
		0,
		10,
		1,
		0,
		"",
		logger,
	)

	cleanup := func() {} // shared db, do not close

	return worker, wordRepo, trainingCardRepo, userCardRepo, userRepo, db, cleanup
}

func TestTrainingWorkerHasMissingData(t *testing.T) {
	worker, _, _, _, _, _, cleanup := newTrainingWorker(t, nil)
	defer cleanup()

	word := "apple"
	pos := "noun"
	trans := "ˈæpəl"
	def := "яблоко"

	if !worker.hasMissingData(&models.WordCard{Word: word}) {
		t.Fatalf("expected missing data")
	}
	if worker.hasMissingData(&models.WordCard{Word: word, POS: &pos, Transcription: &trans, DefinitionRU: &def}) {
		t.Fatalf("expected no missing data")
	}
}

func TestTrainingWorkerFillWordCardData(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"input_word":"run","lemma":"run","pos":"verb","transcription":"rʌn","definition_ru":"бежать","examples":[],"verb_forms":{"v1":"run"}}`}}},
		}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	wordCardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}

	wordCard, err := wordRepo.GetWordCardByID(wordCardID)
	if err != nil {
		t.Fatalf("GetWordCardByID error: %v", err)
	}

	if err := worker.fillWordCardData(context.Background(), wordCard); err != nil {
		t.Fatalf("fillWordCardData error: %v", err)
	}

	updated, err := wordRepo.GetWordCardByID(wordCardID)
	if err != nil {
		t.Fatalf("GetWordCardByID error: %v", err)
	}
	if updated.POS == nil || updated.DefinitionRU == nil || updated.Transcription == nil {
		t.Fatalf("expected fields to be filled")
	}
	if updated.DisplayEN == nil || *updated.DisplayEN != "to run" {
		t.Fatalf("expected display EN to be 'to run'")
	}
}

// TestTrainingWorker_processCard_NoUsersForWordStillCreatesCards verifies that when no users
// requested the word, training cards are still created and processCard returns nil (no user_cards).
func TestTrainingWorker_processCard_NoUsersForWordStillCreatesCards(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		content := `{"word_en":"orphan","lemma":"orphan","transcription":"","senses":[{"pos":"noun","word_ru":"сирота","meaning_en":"orphan","example_en":"","example_ru":"","distractors_ru":["ребенок","вдова"],"distractors_en":["child","widow"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, trainingCardRepo, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "orphan", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, _ := wordRepo.GetWordCardByID(cardID)
	// Do not add any word request history — getUsersForWord will return []

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard: %v", err)
	}
	cards, _ := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if len(cards) != 1 {
		t.Errorf("expected 1 training card when no users for word, got %d", len(cards))
	}
}

func TestTrainingWorkerProcessCardCreatesCards(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"apple","lemma":"apple","transcription":"","senses":[{"pos":"noun","word_ru":"яблоко","meaning_en":"apple","example_en":"","example_ru":"","distractors_ru":["груша","слива"],"distractors_en":["orange","banana"],"hint":""}]}`}}},
		}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, trainingCardRepo, _, userRepo, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	pos := "noun"
	trans := ""
	def := "яблоко"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple", Definition: "", POS: &pos, Transcription: &trans, DefinitionRU: &def})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}

	user, err := userRepo.GetOrCreateUser(100)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}
	word := "apple"
	if err := wordRepo.AddWordRequestHistoryWithCard(user.ID, word, &cardID, &word); err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard error: %v", err)
	}

	wordCard, _ := wordRepo.GetWordCardByID(cardID)
	if err := worker.processCard(context.Background(), wordCard); err != nil {
		t.Fatalf("processCard error: %v", err)
	}

	trainingCards, err := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID error: %v", err)
	}
	if len(trainingCards) != 1 {
		t.Fatalf("expected 1 training card, got %d", len(trainingCards))
	}

	var count int
	if err := db.GetConnection().QueryRow(
		`SELECT COUNT(*) FROM user_cards uc 
		 INNER JOIN training_cards tc ON uc.training_card_id = tc.id 
		 WHERE uc.user_id = ? AND tc.word_card_id = ?`,
		user.ID, cardID,
	).Scan(&count); err != nil {
		t.Fatalf("count user cards error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 user cards, got %d", count)
	}
}

func TestTrainingWorkerGetUsersForWord(t *testing.T) {
	worker, wordRepo, _, _, userRepo, _, cleanup := newTrainingWorker(t, nil)
	defer cleanup()

	user, err := userRepo.GetOrCreateUser(200)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}

	word := "hello"
	if err := wordRepo.AddWordRequestHistoryWithCard(user.ID, word, nil, &word); err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard error: %v", err)
	}

	users, err := worker.getUsersForWord(word)
	if err != nil {
		t.Fatalf("getUsersForWord error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
}

func TestTrainingWorkerProcessCard_SchedulesPronunciationByCanonicalWord(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: `{"word_en":"spy","lemma":"spy","transcription":"spaɪ","senses":[{"pos":"verb","display_word":"to spy","word_ru":"шпионить","meaning_en":"to secretly collect information","example_en":"They spy on competitors.","example_ru":"Они шпионят за конкурентами.","distractors_ru":["прыгать","читать"],"distractors_en":["to jump","to read"],"hint":"secretly watch"}]}`}}},
		}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	pronService := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://127.0.0.1:1/api/v2/entries/en",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		PrefetchEnabled:   true,
		PrefetchWorkers:   1,
	}, nil, zap.NewNop())
	worker.pronunciationService = pronService

	pos := "verb"
	trans := "spaɪ"
	definitionRU := "шпионить"
	displayEN := "to spy"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word:          "spy",
		Definition:    "to secretly collect information",
		POS:           &pos,
		Transcription: &trans,
		DefinitionRU:  &definitionRU,
		DisplayEN:     &displayEN,
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}

	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil {
		t.Fatalf("GetWordCardByID error: %v", err)
	}
	if wordCard == nil {
		t.Fatalf("wordCard should not be nil")
	}

	if err := worker.processCard(context.Background(), wordCard); err != nil {
		t.Fatalf("processCard error: %v", err)
	}

	select {
	case scheduled := <-pronService.queue:
		if scheduled != "spy" {
			t.Fatalf("expected canonical pronunciation scheduling for 'spy', got %q", scheduled)
		}
	default:
		t.Fatalf("expected pronunciation scheduling for canonical word")
	}

	select {
	case extra := <-pronService.queue:
		t.Fatalf("expected no extra pronunciation scheduling for display form, got %q", extra)
	default:
	}
}

func TestTrainingWorkerProcessCard_LLMReturnsError(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		// LLM explicitly rejects the word
		content := `{"error": "not a valid English word", "word_en": "xyz"}`
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: content}}},
		}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "xyz", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should return nil when LLM rejects word, got: %v", err)
	}

	var processingError string
	err = db.GetConnection().QueryRow(
		`SELECT COALESCE(processing_error, '') FROM word_cards WHERE id = $1`, cardID,
	).Scan(&processingError)
	if err != nil {
		t.Fatalf("query processing_error: %v", err)
	}
	if processingError != "not a valid English word" {
		t.Errorf("expected processing_error to be set, got %q", processingError)
	}
}

func TestTrainingWorkerProcessCard_ValidationFailsNoHighModel(t *testing.T) {
	// distractors_en with Cyrillic triggers R1 validation error
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		content := `{"word_en":"run","lemma":"run","transcription":"","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"run","example_en":"","example_ru":"","distractors_ru":["идти","плыть"],"distractors_en":["бежать","to walk"],"hint":""}]}`
		resp := ai.ChatResponse{
			Choices: []ai.Choice{{Message: ai.Message{Content: content}}},
		}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, trainingCardRepo, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	worker.modelHigh = "" // no high model

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should return nil on validation failure (no circuit breaker), got: %v", err)
	}

	var processingError string
	_ = db.GetConnection().QueryRow(
		`SELECT COALESCE(processing_error, '') FROM word_cards WHERE id = $1`, cardID,
	).Scan(&processingError)
	if processingError == "" {
		t.Error("expected processing_error to be set after validation failure")
	}

	cards, _ := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if len(cards) > 0 {
		t.Errorf("expected no training cards when validation fails, got %d", len(cards))
	}
}

func TestTrainingWorker_processCard_NoSensesReturnsError(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		// First call: fillWordCardData (GenerateResponse) — return valid word info
		if callCount == 1 {
			content := `{"input_word":"empty","lemma":"empty","pos":"noun","transcription":"","definition_ru":"пустой"}`
			resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
			return newJSONHTTPResponseTW(http.StatusOK, resp), nil
		}
		// Second call: GenerateTrainingCard — return empty senses to trigger error
		content := `{"word_en":"empty","lemma":"empty","transcription":"","senses":[]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "empty", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err == nil {
		t.Fatal("processCard expected error when LLM returns no senses")
	}
	if !strings.Contains(err.Error(), "no senses") {
		t.Errorf("expected 'no senses' in error, got: %v", err)
	}
}

func TestTrainingWorker_processCard_FillWordCardDataErrorContinues(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		// First call: fillWordCardData (GenerateResponse) — fail
		if callCount == 1 {
			return nil, fmt.Errorf("AI service unavailable")
		}
		// Second call: GenerateTrainingCard — succeed
		content := `{"word_en":"go","lemma":"go","transcription":"","senses":[{"pos":"verb","word_ru":"идти","meaning_en":"go","example_en":"","example_ru":"","distractors_ru":["бежать","плыть"],"distractors_en":["to run","to swim"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, trainingCardRepo, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()

	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "go", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should succeed despite fillWordCardData error (continue anyway): %v", err)
	}
	cards, _ := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if len(cards) != 1 {
		t.Errorf("expected 1 training card created after fill error, got %d", len(cards))
	}
}

func TestTrainingWorkerProcessCard_ValidationFailsHighModelSucceeds(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		// Call 1: fillWordCardData (GenerateResponse)
		// Call 2: GenerateTrainingCard default model — invalid (Cyrillic in distractors_en)
		// Call 3: GenerateTrainingCard high model — valid
		if callCount == 1 {
			content := `{"input_word":"jump","lemma":"jump","pos":"verb","transcription":"dʒʌmp","definition_ru":"прыгать"}`
			resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
			return newJSONHTTPResponseTW(http.StatusOK, resp), nil
		}
		if callCount == 2 {
			content := `{"word_en":"jump","lemma":"jump","transcription":"","senses":[{"pos":"verb","word_ru":"прыгать","meaning_en":"jump","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["прыгать","to run"],"hint":""}]}`
			resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
			return newJSONHTTPResponseTW(http.StatusOK, resp), nil
		}
		content := `{"word_en":"jump","lemma":"jump","transcription":"dʒʌmp","senses":[{"pos":"verb","display_word":"to jump","word_ru":"прыгать","meaning_en":"jump","example_en":"","example_ru":"","distractors_ru":["бежать","идти"],"distractors_en":["to run","to walk"],"hint":""}]}`
		resp := ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}
		return newJSONHTTPResponseTW(http.StatusOK, resp), nil
	})

	worker, wordRepo, trainingCardRepo, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	worker.modelHigh = "high-model"

	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "jump", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}
	wordCard, err := wordRepo.GetWordCardByID(cardID)
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCardByID: %v", err)
	}

	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 LLM calls (fill + default + high model), got %d", callCount)
	}
	cards, err := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 training card from high model, got %d", len(cards))
	}
}
