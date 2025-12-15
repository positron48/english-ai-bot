package service

import (
	"context"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// WordService handles word-related business logic
type WordService struct {
	wordRepo  *repository.WordRepository
	aiService *ai.Service
	logger    *zap.Logger
}

// NewWordService creates a new word service
func NewWordService(wordRepo *repository.WordRepository, aiService *ai.Service, logger *zap.Logger) *WordService {
	return &WordService{
		wordRepo:  wordRepo,
		aiService: aiService,
		logger:    logger,
	}
}

// IsSingleWord checks if the input is a single word
func (s *WordService) IsSingleWord(text string) bool {
	trimmed := strings.TrimSpace(text)
	// Remove common punctuation
	trimmed = strings.Trim(trimmed, ".,!?;:()[]{}'\"")
	trimmed = strings.TrimSpace(trimmed)

	// Check if it's a single word (no spaces)
	parts := strings.Fields(trimmed)
	return len(parts) == 1 && len(trimmed) > 0
}

// NormalizeWord normalizes a word for storage/lookup
func (s *WordService) NormalizeWord(word string) string {
	return strings.TrimSpace(strings.ToLower(word))
}

// GetWordDefinition gets word definition from DB or AI
func (s *WordService) GetWordDefinition(ctx context.Context, userID int64, word string) (string, error) {
	normalizedWord := s.NormalizeWord(word)

	// First, try to get from database
	card, err := s.wordRepo.GetWordCard(normalizedWord)
	if err != nil {
		s.logger.Error("failed to get word card from DB", zap.Error(err))
		// Continue to AI service even if DB query fails
	} else if card != nil {
		s.logger.Info("word found in database",
			zap.String("word", normalizedWord),
		)

		// Record request history
		if err := s.wordRepo.AddWordRequestHistory(userID, normalizedWord); err != nil {
			s.logger.Warn("failed to add word request history", zap.Error(err))
		}

		return card.Definition, nil
	}

	// Word not found in DB, request from AI
	s.logger.Info("word not found in database, requesting from AI",
		zap.String("word", normalizedWord),
	)

	response, err := s.aiService.GenerateResponse(ctx, word)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	// Save word card to database
	if err := s.wordRepo.SaveWordCard(normalizedWord, response); err != nil {
		s.logger.Warn("failed to save word card to database", zap.Error(err))
		// Continue even if save fails, but don't record history if word wasn't saved
	} else {
		// Record request history only if word was successfully saved
		if err := s.wordRepo.AddWordRequestHistory(userID, normalizedWord); err != nil {
			s.logger.Warn("failed to add word request history", zap.Error(err))
		}
	}

	return response, nil
}

// GetWordCard retrieves a word card from database
func (s *WordService) GetWordCard(word string) (*models.WordCard, error) {
	return s.wordRepo.GetWordCard(word)
}

