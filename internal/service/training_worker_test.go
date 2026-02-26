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
