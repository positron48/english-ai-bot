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
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newAIServiceWithResponse(t *testing.T, logger *zap.Logger, content string) *ai.Service {
	t.Helper()

	aiService := ai.NewService("http://example.com", "model", "key", "prompt", logger)
	mockClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			payload := ai.ChatResponse{
				Choices: []ai.Choice{
					{Message: ai.Message{Role: "assistant", Content: content}},
				},
			}
			body, _ := json.Marshal(payload)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}

	aiValue := reflect.ValueOf(aiService).Elem()
	clientField := aiValue.FieldByName("client")
	if !clientField.IsValid() {
		t.Fatalf("ai service client field not found")
	}
	reflect.NewAt(clientField.Type(), unsafe.Pointer(clientField.UnsafeAddr())).Elem().Set(reflect.ValueOf(mockClient))
	return aiService
}

func TestGetWordDefinition_FoundInDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	service := NewWordService(wordRepo, nil, nil, nil, config.DefaultLearningConfig(), logger)

	_, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple", CourseCode: "en_ru"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}

	resp, err := service.GetWordDefinition(context.Background(), 1, "apple")
	if err != nil {
		t.Fatalf("GetWordDefinition error: %v", err)
	}
	if !strings.Contains(resp, "apple") {
		t.Fatalf("expected response to include word, got %q", resp)
	}
}

// TestGetWordDefinition_WordFormMapping covers resolution via word_forms (form -> lemma).
func TestGetWordDefinition_WordFormMapping(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	userRepo := repository.NewUserRepository(conn, logger)
	_, err := userRepo.GetOrCreateUser(1)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	wordRepo := repository.NewWordRepository(conn, logger)
	service := NewWordService(wordRepo, nil, nil, nil, config.DefaultLearningConfig(), logger)

	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run", Definition: "to move fast", CourseCode: "en_ru"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	err = wordRepo.UpsertWordFormMapping("ran", cardID)
	if err != nil {
		t.Fatalf("UpsertWordFormMapping: %v", err)
	}

	resp, err := service.GetWordDefinition(context.Background(), 1, "ran")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	// Form "ran" must resolve to lemma "run" and appear in response
	if !strings.Contains(resp, "run") {
		t.Errorf("expected response to resolve form 'ran' to lemma 'run', got %q", resp)
	}
}

func TestGetWordDefinition_AIResponse_JSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	aiService := newAIServiceWithResponse(t, logger, `{"lemma":"banana","pos":"noun","transcription":"bənænə","definition_ru":"банан","examples":[{"example_en":"I ate a banana.","gloss_ru":"Я съел банан."}]}`)
	service := NewWordService(wordRepo, nil, nil, aiService, config.DefaultLearningConfig(), logger)

	resp, err := service.GetWordDefinition(context.Background(), 1, "banana")
	if err != nil {
		t.Fatalf("GetWordDefinition error: %v", err)
	}
	if !strings.Contains(resp, "банан") {
		t.Fatalf("expected response to include definition, got %q", resp)
	}

	card, err := wordRepo.GetWordCardByLemma("banana")
	if err != nil {
		t.Fatalf("GetWordCardByLemma error: %v", err)
	}
	if card == nil || card.Word != "banana" {
		t.Fatalf("expected saved word card")
	}
}

func TestGetWordDefinition_AIResponse_ErrorHint(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	aiService := newAIServiceWithResponse(t, logger, `{"error": true, "hint": "проверьте написание"}`)
	service := NewWordService(wordRepo, nil, nil, aiService, config.DefaultLearningConfig(), logger)

	resp, err := service.GetWordDefinition(context.Background(), 1, "asdf")
	if err != nil {
		t.Fatalf("GetWordDefinition error: %v", err)
	}
	if !strings.Contains(resp, "проверьте написание") {
		t.Fatalf("expected hint response, got %q", resp)
	}

	card, err := wordRepo.GetWordCardByLemma("asdf")
	if err != nil {
		t.Fatalf("GetWordCardByLemma error: %v", err)
	}
	if card != nil {
		t.Fatalf("expected no word card saved")
	}
}

func TestGetWordDefinition_AIResponse_NoDefinitionRU(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	aiService := newAIServiceWithResponse(t, logger, `{"lemma":"ghost","pos":"noun"}`)
	service := NewWordService(wordRepo, nil, nil, aiService, config.DefaultLearningConfig(), logger)

	resp, err := service.GetWordDefinition(context.Background(), 1, "ghost")
	if err != nil {
		t.Fatalf("GetWordDefinition error: %v", err)
	}
	if !strings.Contains(resp, "опечатка") {
		t.Fatalf("expected default error message, got %q", resp)
	}
}

func TestGetWordDefinition_AIResponse_InvalidJSON_Legacy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	aiService := newAIServiceWithResponse(t, logger, "not json")
	service := NewWordService(wordRepo, nil, nil, aiService, config.DefaultLearningConfig(), logger)

	resp, err := service.GetWordDefinition(context.Background(), 1, "legacy")
	if err != nil {
		t.Fatalf("GetWordDefinition error: %v", err)
	}
	if resp != "not json" {
		t.Fatalf("expected legacy response, got %q", resp)
	}

	card, err := wordRepo.GetWordCardByLemma("legacy")
	if err != nil {
		t.Fatalf("GetWordCardByLemma error: %v", err)
	}
	if card == nil || card.Definition != "not json" {
		t.Fatalf("expected legacy word card saved")
	}
}

// TestGetWordDefinition_AINil covers word not in DB with nil AI service.
func TestGetWordDefinition_AINil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	service := NewWordService(wordRepo, nil, nil, nil, config.DefaultLearningConfig(), logger)

	_, err := service.GetWordDefinition(context.Background(), 1, "newword")
	if err == nil {
		t.Fatal("expected error when AI service is nil")
	}
	if !strings.Contains(err.Error(), "AI service not available") {
		t.Errorf("expected AI not available error, got: %v", err)
	}
}

// TestGetWordDefinition_Cyrillic returns AI response without saving when word contains Cyrillic.
func TestGetWordDefinition_Cyrillic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	aiService := newAIServiceWithResponse(t, logger, `{"lemma":"тест","definition_ru":"проверка"}`)
	service := NewWordService(wordRepo, nil, nil, aiService, config.DefaultLearningConfig(), logger)

	resp, err := service.GetWordDefinition(context.Background(), 1, "привет")
	if err != nil {
		t.Fatalf("GetWordDefinition error: %v", err)
	}
	if !strings.Contains(resp, "проверка") && !strings.Contains(resp, "тест") {
		t.Errorf("expected AI response returned as-is, got %q", resp)
	}
	card, _ := wordRepo.GetWordCardByLemma("привет")
	if card != nil {
		t.Error("expected no word card saved for Cyrillic input")
	}
}

// TestGetWordDefinition_AIResponse_ErrorNoHint covers Error.IsTrue with empty hint -> default message.
func TestGetWordDefinition_AIResponse_ErrorNoHint(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	aiService := newAIServiceWithResponse(t, logger, `{"error": true, "hint": ""}`)
	service := NewWordService(wordRepo, nil, nil, aiService, config.DefaultLearningConfig(), logger)

	resp, err := service.GetWordDefinition(context.Background(), 1, "xyz")
	if err != nil {
		t.Fatalf("GetWordDefinition error: %v", err)
	}
	if !strings.Contains(resp, "опечатка") && !strings.Contains(resp, "несуществующее") {
		t.Errorf("expected default error message, got %q", resp)
	}
}

// TestGetWordDefinition_AIResponse_ErrorKeyword covers legacy string error with keyword (e.g. gibberish).
func TestGetWordDefinition_AIResponse_ErrorKeyword(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantHint string
	}{
		{
			name:     "keyword with hint",
			response: `{"error": "gibberish word", "hint": "проверьте слово", "lemma": ""}`,
			wantHint: "проверьте слово",
		},
		{
			name:     "keyword no hint",
			response: `{"error": "not a valid English word", "hint": "", "lemma": ""}`,
			wantHint: "опечатка",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			db := testutil.SetupTestDatabase(t)
			wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
			aiService := newAIServiceWithResponse(t, logger, tt.response)
			service := NewWordService(wordRepo, nil, nil, aiService, config.DefaultLearningConfig(), logger)

			resp, err := service.GetWordDefinition(context.Background(), 1, "qwerty")
			if err != nil {
				t.Fatalf("GetWordDefinition error: %v", err)
			}
			if tt.wantHint != "опечатка" && !strings.Contains(resp, tt.wantHint) {
				t.Errorf("expected response to contain %q, got %q", tt.wantHint, resp)
			}
			if tt.wantHint == "опечатка" && !strings.Contains(resp, "опечатка") && !strings.Contains(resp, "несуществующее") {
				t.Errorf("expected default error message, got %q", resp)
			}
		})
	}
}

// TestGetWordDefinition_FoundInDB_ByLemma covers finding by lemma and returning markdown (e.g. "Run" -> "run").
func TestGetWordDefinition_FoundInDB_ByLemma(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	wordRepo := repository.NewWordRepository(conn, logger)
	service := NewWordService(wordRepo, nil, nil, nil, config.DefaultLearningConfig(), logger)

	_, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run", Definition: "to move", CourseCode: "en_ru"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	resp, err := service.GetWordDefinition(context.Background(), 1, "Run")
	if err != nil {
		t.Fatalf("GetWordDefinition: %v", err)
	}
	if !strings.Contains(resp, "run") {
		t.Errorf("expected lemma in response, got %q", resp)
	}
}

func TestWordService_renderWordCardMarkdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewWordService(nil, nil, nil, nil, config.DefaultLearningConfig(), logger)

	examplesJSON := `[{"example_en":"I run.","gloss_ru":"Я бегу."}]`
	verbFormsJSON := `{"v1":"run","v2":"ran","v3":"run"}`
	card := &models.WordCard{
		Word:          "run",
		Definition:    "to move fast",
		ExamplesJSON:  &examplesJSON,
		VerbFormsJSON: &verbFormsJSON,
	}
	md := service.renderWordCardMarkdown(card)
	if md == "" {
		t.Fatal("expected non-empty markdown")
	}
	if !strings.Contains(md, "run") {
		t.Errorf("expected word in markdown, got %q", md)
	}
	if !strings.Contains(md, "I run") && !strings.Contains(md, "бегу") {
		t.Errorf("expected examples in markdown, got %q", md)
	}
}

func TestWordService_ensureUserCardsForWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userRepo := repository.NewUserRepository(conn, logger)

	user, _ := userRepo.GetOrCreateUser(999)
	wordCardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "ensureword", CourseCode: "en_ru"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	tcID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "ensureword",
		WordRU:     "обеспечить",
		MeaningEN:  "ensure",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}
	_ = tcID

	service := NewWordService(wordRepo, trainingCardRepo, userCardRepo, nil, config.DefaultLearningConfig(), logger)
	err = service.ensureUserCardsForWord(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("ensureUserCardsForWord: %v", err)
	}

	var count int
	err = conn.QueryRow(
		`SELECT COUNT(*) FROM user_cards uc INNER JOIN training_cards tc ON uc.training_card_id = tc.id WHERE uc.user_id = $1 AND tc.word_card_id = $2`,
		user.ID, wordCardID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count user_cards: %v", err)
	}
	if count < 2 {
		t.Errorf("expected at least 2 user_cards (both directions), got %d", count)
	}
}

// TestWordService_ensureUserCardsForWord_NoTrainingCards covers len(trainingCards)==0 -> nil.
func TestWordService_ensureUserCardsForWord_NoTrainingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userRepo := repository.NewUserRepository(conn, logger)

	user, _ := userRepo.GetOrCreateUser(888)
	wordCardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "nocard", CourseCode: "en_ru"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	// No training cards for this word

	service := NewWordService(wordRepo, trainingCardRepo, userCardRepo, nil, config.DefaultLearningConfig(), logger)
	err = service.ensureUserCardsForWord(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("ensureUserCardsForWord: %v", err)
	}
}

// TestWordService_ensureUserCardsForWord_SecondCallIdempotent covers CreateUserCard duplicate (warn path); second call still succeeds.
func TestWordService_ensureUserCardsForWord_SecondCallIdempotent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userRepo := repository.NewUserRepository(conn, logger)

	user, _ := userRepo.GetOrCreateUser(777)
	wordCardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "idem", CourseCode: "en_ru"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_, err = trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "idem",
		WordRU:     "идем",
		MeaningEN:  "go",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	service := NewWordService(wordRepo, trainingCardRepo, userCardRepo, nil, config.DefaultLearningConfig(), logger)
	err = service.ensureUserCardsForWord(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("ensureUserCardsForWord first call: %v", err)
	}
	err = service.ensureUserCardsForWord(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("ensureUserCardsForWord second call (idempotent): %v", err)
	}
}

// TestWordService_ensureUserCardsForWord_WithMasteringRepo covers createdCount > 0 and userWordMasteringRepo.Upsert.
func TestWordService_ensureUserCardsForWord_WithMasteringRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	userRepo := repository.NewUserRepository(conn, logger)

	user, _ := userRepo.GetOrCreateUser(666)
	wordCardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "master", CourseCode: "en_ru"})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_, err = trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "master",
		WordRU:     "мастер",
		MeaningEN:  "master",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	service := NewWordServiceWithMastering(wordRepo, trainingCardRepo, userCardRepo, userWordMasteringRepo, nil, config.DefaultLearningConfig(), logger)
	err = service.ensureUserCardsForWord(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("ensureUserCardsForWord: %v", err)
	}
	score, err := userWordMasteringRepo.GetScore(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("GetScore: %v", err)
	}
	if score != 0 {
		t.Errorf("expected initial mastering score 0, got %d", score)
	}
}
