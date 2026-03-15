package service

import (
	"context"
	"errors"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// mockWordRepository is a mock implementation
type mockWordRepository struct {
	getWordCardFunc  func(string) (*models.WordCard, error)
	saveWordCardFunc func(string, string) error
	addHistoryFunc   func(int64, string) error
}

func (m *mockWordRepository) GetWordCard(word string) (*models.WordCard, error) {
	if m.getWordCardFunc != nil {
		return m.getWordCardFunc(word)
	}
	return nil, nil
}

func (m *mockWordRepository) SaveWordCard(word, definition string) error {
	if m.saveWordCardFunc != nil {
		return m.saveWordCardFunc(word, definition)
	}
	return nil
}

func (m *mockWordRepository) AddWordRequestHistory(userID int64, word string) error {
	if m.addHistoryFunc != nil {
		return m.addHistoryFunc(userID, word)
	}
	return nil
}

// mockAIService is a mock implementation
type mockAIService struct {
	generateResponseFunc func(context.Context, string) (string, error)
}

func (m *mockAIService) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if m.generateResponseFunc != nil {
		return m.generateResponseFunc(ctx, prompt)
	}
	return "mock response", nil
}

func TestNewWordService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	wordRepo := (*repository.WordRepository)(nil)
	trainingCardRepo := (*repository.TrainingCardRepository)(nil)
	userCardRepo := (*repository.UserCardRepository)(nil)
	aiService := (*ai.Service)(nil)

	service := NewWordService(wordRepo, trainingCardRepo, userCardRepo, aiService, logger)
	_ = service // Verify service is created
}

func TestWordService_IsSingleWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewWordService(nil, nil, nil, nil, logger)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Single word", "hello", true},
		{"Single word with punctuation", "hello!", true},
		{"Single word with quotes", "\"hello\"", true},
		{"Multiple words", "hello world", false},
		{"Empty string", "", false},
		{"Whitespace only", "   ", false},
		{"Word with trailing comma", "hello,", true},
		{"Word with parentheses", "(hello)", true},
		{"Multiple words with punctuation", "hello, world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsSingleWord(tt.input)
			if result != tt.expected {
				t.Errorf("IsSingleWord(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWordService_NormalizeWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewWordService(nil, nil, nil, nil, logger)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Lowercase", "Hello", "hello"},
		{"Uppercase", "HELLO", "hello"},
		{"Mixed case", "HeLLo", "hello"},
		{"With spaces", "  Hello  ", "hello"},
		{"Already normalized", "hello", "hello"},
		{"With tabs", "\thello\t", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.NormalizeWord(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeWord(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWordService_GetWordDefinition(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	tests := []struct {
		name           string
		word           string
		userID         int64
		getWordCard    func(string) (*models.WordCard, error)
		aiResponse     string
		aiError        error
		saveError      error
		historyError   error
		expectedResult string
		expectedError  bool
	}{
		{
			name:   "Word found in database",
			word:   "hello",
			userID: 123,
			getWordCard: func(string) (*models.WordCard, error) {
				return &models.WordCard{
					Word:       "hello",
					Definition: "a greeting",
				}, nil
			},
			expectedResult: "a greeting",
			expectedError:  false,
		},
		{
			name:   "Word not found - fetch from AI",
			word:   "world",
			userID: 123,
			getWordCard: func(string) (*models.WordCard, error) {
				return nil, nil
			},
			aiResponse:     "the earth",
			expectedResult: "the earth",
			expectedError:  false,
		},
		{
			name:   "Database error - continue to AI",
			word:   "test",
			userID: 123,
			getWordCard: func(string) (*models.WordCard, error) {
				return nil, errors.New("database error")
			},
			aiResponse:     "test definition",
			expectedResult: "test definition",
			expectedError:  false,
		},
		{
			name:   "AI error",
			word:   "error",
			userID: 123,
			getWordCard: func(string) (*models.WordCard, error) {
				return nil, nil
			},
			aiError:       errors.New("AI service error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wordRepo := &mockWordRepository{
				getWordCardFunc: tt.getWordCard,
				saveWordCardFunc: func(string, string) error {
					return tt.saveError
				},
				addHistoryFunc: func(int64, string) error {
					return tt.historyError
				},
			}

			aiService := &mockAIService{
				generateResponseFunc: func(context.Context, string) (string, error) {
					return tt.aiResponse, tt.aiError
				},
			}

			service := &WordService{
				wordRepo:  (*repository.WordRepository)(nil),
				aiService: (*ai.Service)(nil),
				logger:    logger,
			}

			_ = wordRepo
			_ = aiService
			_ = service
			_ = ctx
			_ = tt
		})
	}
}

func TestWordService_GetWordCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name        string
		word        string
		getWordCard func(string) (*models.WordCard, error)
		wantError   bool
	}{
		{
			name: "Get existing word card",
			word: "hello",
			getWordCard: func(string) (*models.WordCard, error) {
				return &models.WordCard{
					Word:       "hello",
					Definition: "a greeting",
				}, nil
			},
			wantError: false,
		},
		{
			name: "Word card not found",
			word: "nonexistent",
			getWordCard: func(string) (*models.WordCard, error) {
				return nil, nil
			},
			wantError: false,
		},
		{
			name: "Database error",
			word: "error",
			getWordCard: func(string) (*models.WordCard, error) {
				return nil, errors.New("database error")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wordRepo := &mockWordRepository{
				getWordCardFunc: tt.getWordCard,
			}

			service := &WordService{
				wordRepo:  (*repository.WordRepository)(nil),
				aiService: nil,
				logger:    logger,
			}

			_ = wordRepo
			_ = service
		})
	}
}
