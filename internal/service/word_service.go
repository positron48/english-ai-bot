package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/utils"

	"go.uber.org/zap"
)

// WordService handles word-related business logic
type WordService struct {
	wordRepo         *repository.WordRepository
	trainingCardRepo *repository.TrainingCardRepository
	userCardRepo     *repository.UserCardRepository
	aiService        *ai.Service
	logger           *zap.Logger
}

// NewWordService creates a new word service
func NewWordService(
	wordRepo *repository.WordRepository,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo *repository.UserCardRepository,
	aiService *ai.Service,
	logger *zap.Logger,
) *WordService {
	return &WordService{
		wordRepo:         wordRepo,
		trainingCardRepo: trainingCardRepo,
		userCardRepo:     userCardRepo,
		aiService:        aiService,
		logger:           logger,
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
// New flow: resolve word form -> lemma -> render markdown from structured data
func (s *WordService) GetWordDefinition(ctx context.Context, userID int64, word string) (string, error) {
	normalizedWord := s.NormalizeWord(word)
	inputWord := word // Keep original for history

	// Check if word contains Cyrillic characters - don't save to DB
	if ContainsCyrillic(normalizedWord) {
		s.logger.Info("word contains Cyrillic, not saving to database",
			zap.String("word", normalizedWord),
			zap.Int64("user_id", userID),
		)
		return "💡 Пожалуйста, введите слово на английском языке.", nil
	}

	// Step 1: Try to resolve word form to lemma via word_forms table
	wordForm, err := s.wordRepo.GetWordFormMapping(normalizedWord)
	if err != nil {
		s.logger.Warn("failed to get word form mapping", zap.Error(err))
	}

	var wordCard *models.WordCard
	if wordForm != nil {
		// Found mapping, get lemma
		wordCard, err = s.wordRepo.GetWordCardByID(wordForm.WordCardID)
		if err != nil {
			s.logger.Warn("failed to get word card by ID", zap.Error(err))
		}
	}

	// Step 2: If no mapping found, try direct lookup by lemma
	if wordCard == nil {
		wordCard, err = s.wordRepo.GetWordCardByLemma(normalizedWord)
		if err != nil {
			s.logger.Warn("failed to get word card by lemma", zap.Error(err))
		}
	}

	// Step 3: If found in DB, render markdown and return
	if wordCard != nil {
		s.logger.Info("word found in database",
			zap.String("input", inputWord),
			zap.String("lemma", wordCard.Word),
		)

		// Create mapping if it doesn't exist (for word forms)
		if wordForm == nil && normalizedWord != strings.ToLower(wordCard.Word) {
			if err := s.wordRepo.UpsertWordFormMapping(normalizedWord, wordCard.ID); err != nil {
				s.logger.Warn("failed to create word form mapping", zap.Error(err))
			}
		}

		// Record request history
		wordCardID := wordCard.ID
		if err := s.wordRepo.AddWordRequestHistoryWithCard(userID, inputWord, &wordCardID, nil); err != nil {
			s.logger.Warn("failed to add word request history", zap.Error(err))
		}

		// Create user_cards for existing training_cards if they exist
		// This ensures that when a user requests a word that already has training cards,
		// they get linked to the user for training
		if s.trainingCardRepo != nil && s.userCardRepo != nil {
			if err := s.ensureUserCardsForWord(userID, wordCardID); err != nil {
				s.logger.Warn("failed to ensure user cards for word",
					zap.Int64("user_id", userID),
					zap.Int64("word_card_id", wordCardID),
					zap.Error(err),
				)
				// Don't fail the request if user cards creation fails
			}
		}

		// Render markdown from structured data
		markdown := s.renderWordCardMarkdown(wordCard)
		return markdown, nil
	}

	// Step 4: Word not found, request from AI (expecting JSON)
	s.logger.Info("word not found in database, requesting from AI",
		zap.String("word", normalizedWord),
	)

	if s.aiService == nil {
		return "", fmt.Errorf("AI service not available")
	}

	response, err := s.aiService.GenerateResponse(ctx, word)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	// Parse JSON response
	var wordInfo models.WordInfoResponse
	if err := json.Unmarshal([]byte(response), &wordInfo); err != nil {
		// Not JSON, might be old format - try to save as-is for backward compatibility
		s.logger.Warn("failed to parse AI response as JSON, saving as legacy format",
			zap.Error(err),
			zap.String("response", response[:min(100, len(response))]),
		)
	if err := s.wordRepo.SaveWordCard(normalizedWord, response); err != nil {
			s.logger.Warn("failed to save word card", zap.Error(err))
	} else {
		if err := s.wordRepo.AddWordRequestHistory(userID, normalizedWord); err != nil {
			s.logger.Warn("failed to add word request history", zap.Error(err))
		}
	}
	return response, nil
	}

	// Check for error from LLM
	// If definition_ru is present, ignore error field - we have valid data
	hasDefinitionRU := strings.TrimSpace(wordInfo.DefinitionRU) != ""
	
	// Priority 1: If error is true (bool or string "true"), check hint first
	// BUT skip if we have definition_ru (valid data)
	if !hasDefinitionRU && wordInfo.Error.IsTrue() {
		hint := strings.TrimSpace(wordInfo.Hint)
		if hint != "" {
			// If hint is present, return only hint
			message := fmt.Sprintf("💡 %s", hint)
			return message, nil
		} else {
			// If hint is empty, return default error message
			message := "💡 Это слово, скорее всего, опечатка или несуществующее английское слово."
			return message, nil
		}
	}
	
	// Priority 3: Legacy handling for string error messages
	// LLM sometimes puts non-error strings in error field (like "load", "master", "none", "valid English word")
	// Skip this check if we have definition_ru (valid data)
	if !hasDefinitionRU {
		errorMsg := strings.TrimSpace(wordInfo.Error.Message)
		hasValidData := wordInfo.Lemma != "" && wordInfo.POS != ""
		
		// List of known non-error strings that LLM sometimes puts in error field
		nonErrorStrings := []string{
			"null", "none", "false", "no",
			"load", "master", "slave", "tor", "corm",
			"valid english word", "valid English word",
		}
		
		// Keywords that indicate a real error (word doesn't exist, gibberish, etc.)
		errorKeywords := []string{
			"gibberish", "does not exist", "not exist", "non-standard", "not a valid",
			"not an english", "not english", "not recognized", "not a word",
			"doesn't exist", "not found", "invalid word", "not a real",
		}
		
		isNonErrorString := false
		errorMsgLower := strings.ToLower(errorMsg)
		for _, nonError := range nonErrorStrings {
			if errorMsgLower == strings.ToLower(nonError) {
				isNonErrorString = true
				break
			}
		}
		
		// Check if error message contains keywords indicating a real error
		isRealError := false
		if errorMsg != "" {
			for _, keyword := range errorKeywords {
				if strings.Contains(errorMsgLower, keyword) {
					isRealError = true
					break
				}
			}
		}
		
		// Treat as error if:
		// 1. Error field contains keywords indicating real error (word doesn't exist, etc.) OR
		// 2. Error field is not empty AND it's not a known non-error string AND we don't have valid data
		if errorMsg != "" && (isRealError || (!isNonErrorString && !hasValidData)) {
			// Check hint first
			hint := strings.TrimSpace(wordInfo.Hint)
			if hint != "" {
				// If hint is present, return only hint
				message := fmt.Sprintf("💡 %s", hint)
				return message, nil
			} else {
				// If hint is empty, return default error message
				message := "💡 Это слово, скорее всего, опечатка или несуществующее английское слово."
				return message, nil
			}
		}
	}

	// Step 5: Check if we have valid data (definition_ru is required for valid word card)
	// If no valid data, don't save to database
	// hasDefinitionRU already declared above, just check it again
	if !hasDefinitionRU {
		s.logger.Info("LLM did not return valid word card (no definition_ru), not saving to database",
			zap.String("word", normalizedWord),
			zap.Int64("user_id", userID),
		)
		// Return hint if available, otherwise default message
		hint := strings.TrimSpace(wordInfo.Hint)
		if hint != "" {
			return fmt.Sprintf("💡 %s", hint), nil
		}
		return "💡 Это слово, скорее всего, опечатка или несуществующее английское слово.", nil
	}

	// Step 6: Save structured data to word_cards (lemma)
	lemma := strings.ToLower(wordInfo.Lemma)
	if lemma == "" {
		lemma = normalizedWord
	}

	displayEN := lemma
	if wordInfo.POS == "verb" && wordInfo.VerbForms != nil && wordInfo.VerbForms.V1 != "" {
		displayEN = "to " + wordInfo.VerbForms.V1
	}

	wordCard = &models.WordCard{
		Word:          lemma,
		Definition:    "", // Legacy field, keep empty
		POS:           &wordInfo.POS,
		Transcription: &wordInfo.Transcription,
		DefinitionRU: &wordInfo.DefinitionRU,
		DisplayEN:     &displayEN,
	}

	// Serialize examples
	if len(wordInfo.Examples) > 0 {
		examplesJSON, _ := json.Marshal(wordInfo.Examples)
		examplesStr := string(examplesJSON)
		wordCard.ExamplesJSON = &examplesStr
	}

	// Serialize verb forms
	if wordInfo.VerbForms != nil {
		verbFormsJSON, _ := json.Marshal(wordInfo.VerbForms)
		verbFormsStr := string(verbFormsJSON)
		wordCard.VerbFormsJSON = &verbFormsStr
	}

	wordCardID, err := s.wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		s.logger.Warn("failed to save word card lemma", zap.Error(err))
		return "", fmt.Errorf("failed to save word card: %w", err)
	}

	// Step 7: Create word form mappings
	// Map input word to lemma
	if normalizedWord != lemma {
		if err := s.wordRepo.UpsertWordFormMapping(normalizedWord, wordCardID); err != nil {
			s.logger.Warn("failed to create word form mapping", zap.Error(err))
		}
	}

	// Map lemma to itself
	if err := s.wordRepo.UpsertWordFormMapping(lemma, wordCardID); err != nil {
		s.logger.Warn("failed to create lemma mapping", zap.Error(err))
	}

	// Map verb forms if present
	if wordInfo.VerbForms != nil {
		forms := []string{wordInfo.VerbForms.V1, wordInfo.VerbForms.V2, wordInfo.VerbForms.V3}
		if wordInfo.VerbForms.Gerund != "" {
			forms = append(forms, wordInfo.VerbForms.Gerund)
		}
		if wordInfo.VerbForms.ThirdPerson != "" {
			forms = append(forms, wordInfo.VerbForms.ThirdPerson)
		}
		for _, form := range forms {
			if form != "" && strings.ToLower(form) != lemma {
				if err := s.wordRepo.UpsertWordFormMapping(strings.ToLower(form), wordCardID); err != nil {
					s.logger.Warn("failed to create verb form mapping", zap.String("form", form), zap.Error(err))
				}
			}
		}
	}

	// Step 8: Record request history
	if err := s.wordRepo.AddWordRequestHistoryWithCard(userID, inputWord, &wordCardID, nil); err != nil {
		s.logger.Warn("failed to add word request history", zap.Error(err))
	}

	// Step 9: Render and return markdown
	markdown := s.renderWordCardMarkdown(wordCard)
	return markdown, nil
}

// renderWordCardMarkdown renders markdown from structured WordCard data
func (s *WordService) renderWordCardMarkdown(card *models.WordCard) string {
	var examples []models.WordInfoExample
	var verbForms *models.WordInfoVerbForms

	// Parse examples
	if card.ExamplesJSON != nil && *card.ExamplesJSON != "" {
		if err := json.Unmarshal([]byte(*card.ExamplesJSON), &examples); err != nil {
			s.logger.Warn("failed to parse examples JSON", zap.Error(err))
		}
	}

	// Parse verb forms
	if card.VerbFormsJSON != nil && *card.VerbFormsJSON != "" {
		var vf models.WordInfoVerbForms
		if err := json.Unmarshal([]byte(*card.VerbFormsJSON), &vf); err != nil {
			s.logger.Warn("failed to parse verb forms JSON", zap.Error(err))
		} else {
			verbForms = &vf
		}
	}

	return utils.RenderWordCardMarkdown(card, examples, verbForms)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetWordCard retrieves a word card from database
func (s *WordService) GetWordCard(word string) (*models.WordCard, error) {
	return s.wordRepo.GetWordCard(word)
}

// ensureUserCardsForWord creates user_cards for all training_cards of a word if they don't exist
// This is called when a user requests a word that already exists in the database with training cards
func (s *WordService) ensureUserCardsForWord(userID, wordCardID int64) error {
	// Get all training cards for this word
	trainingCards, err := s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to get training cards: %w", err)
	}

	if len(trainingCards) == 0 {
		// No training cards yet, nothing to do
		return nil
	}

	// Create user_cards for each training card (both directions)
	createdCount := 0
	for _, trainingCard := range trainingCards {
		// Create ru_en card
		ruEnCard := &models.UserCard{
			UserID:         userID,
			TrainingCardID: trainingCard.ID,
			Direction:      models.DirectionRUtoEN,
			State:          models.StateNew,
			EF:             models.InitialEF,
		}
		if _, err := s.userCardRepo.CreateUserCard(ruEnCard); err != nil {
			s.logger.Warn("failed to create ru_en user card",
				zap.Int64("user_id", userID),
				zap.Int64("training_card_id", trainingCard.ID),
				zap.Error(err),
			)
		} else {
			createdCount++
		}

		// Create en_ru card
		enRuCard := &models.UserCard{
			UserID:         userID,
			TrainingCardID: trainingCard.ID,
			Direction:      models.DirectionENtoRU,
			State:          models.StateNew,
			EF:             models.InitialEF,
		}
		if _, err := s.userCardRepo.CreateUserCard(enRuCard); err != nil {
			s.logger.Warn("failed to create en_ru user card",
				zap.Int64("user_id", userID),
				zap.Int64("training_card_id", trainingCard.ID),
				zap.Error(err),
			)
		} else {
			createdCount++
		}
	}

	if createdCount > 0 {
		s.logger.Info("created user cards for existing word",
			zap.Int64("user_id", userID),
			zap.Int64("word_card_id", wordCardID),
			zap.Int("training_cards", len(trainingCards)),
			zap.Int("user_cards_created", createdCount),
		)
	}

	return nil
}
