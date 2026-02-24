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
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

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
	service := NewWordService(wordRepo, nil, nil, nil, logger)

	_, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple"})
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

func TestGetWordDefinition_AIResponse_JSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	aiService := newAIServiceWithResponse(t, logger, `{"lemma":"banana","pos":"noun","transcription":"bənænə","definition_ru":"банан","examples":[{"example_en":"I ate a banana.","gloss_ru":"Я съел банан."}]}`)
	service := NewWordService(wordRepo, nil, nil, aiService, logger)

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
	service := NewWordService(wordRepo, nil, nil, aiService, logger)

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
	service := NewWordService(wordRepo, nil, nil, aiService, logger)

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
	service := NewWordService(wordRepo, nil, nil, aiService, logger)

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
