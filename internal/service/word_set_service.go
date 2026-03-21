package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// userCardRepoForWordSet is used by WordSetService for MarkKnown and EnsureUserCardsForWord (allows mocks in tests).
type userCardRepoForWordSet interface {
	CreateUserCard(card *models.UserCard) (int64, error)
	DeleteUserCardsByWordCardIDForUser(userID, wordCardID int64) (int64, error)
}

// wordRepoForWordSet is used by WordSetService for word card operations (allows mocks in tests).
type wordRepoForWordSet interface {
	GetWordCardByLemma(lemma string) (*models.WordCard, error)
	GetWordCardByID(id int64) (*models.WordCard, error)
	SaveWordCard(word, content string) error
	UpsertWordCardLemma(card *models.WordCard) (int64, error)
	UpsertWordFormMapping(form string, wordCardID int64) error
}

// WordSetService handles word set business logic
type WordSetService struct {
	wordSetRepo           *repository.WordSetRepository
	wordSetCategoryRepo   *repository.WordSetCategoryRepository
	wordRepo              wordRepoForWordSet
	trainingCardRepo      *repository.TrainingCardRepository
	userCardRepo          userCardRepoForWordSet
	userWordKnowledgeRepo *repository.UserWordKnowledgeRepository
	userWordMasteringRepo *repository.UserWordMasteringRepository
	aiService             *ai.Service
	learning              config.LearningConfig
	modelHigh             string
	logger                *zap.Logger
}

// NewWordSetService creates a new word set service
func NewWordSetService(
	wordSetRepo *repository.WordSetRepository,
	wordSetCategoryRepo *repository.WordSetCategoryRepository,
	wordRepo wordRepoForWordSet,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo userCardRepoForWordSet,
	userWordKnowledgeRepo *repository.UserWordKnowledgeRepository,
	aiService *ai.Service,
	learning config.LearningConfig,
	modelHigh string,
	logger *zap.Logger,
) *WordSetService {
	return NewWordSetServiceWithMastering(wordSetRepo, wordSetCategoryRepo, wordRepo, trainingCardRepo, userCardRepo, userWordKnowledgeRepo, nil, aiService, learning, modelHigh, logger)
}

// NewWordSetServiceWithMastering creates a word set service with optional UserWordMasteringRepository for storing mastering score on card creation.
func NewWordSetServiceWithMastering(
	wordSetRepo *repository.WordSetRepository,
	wordSetCategoryRepo *repository.WordSetCategoryRepository,
	wordRepo wordRepoForWordSet,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo userCardRepoForWordSet,
	userWordKnowledgeRepo *repository.UserWordKnowledgeRepository,
	userWordMasteringRepo *repository.UserWordMasteringRepository,
	aiService *ai.Service,
	learning config.LearningConfig,
	modelHigh string,
	logger *zap.Logger,
) *WordSetService {
	return &WordSetService{
		wordSetRepo:           wordSetRepo,
		wordSetCategoryRepo:   wordSetCategoryRepo,
		wordRepo:              wordRepo,
		trainingCardRepo:      trainingCardRepo,
		userCardRepo:          userCardRepo,
		userWordKnowledgeRepo: userWordKnowledgeRepo,
		userWordMasteringRepo: userWordMasteringRepo,
		aiService:             aiService,
		learning:              learning,
		modelHigh:             modelHigh,
		logger:                logger,
	}
}

// EnsureWordCardExistsMinimal ensures a word card exists with minimal data (only word)
// Returns the word card ID
// This is used for fast word set creation - word cards are created without LLM call
// Word cards will be filled asynchronously by TrainingWorker
func (s *WordSetService) EnsureWordCardExistsMinimal(word string) (int64, error) {
	normalizedWord := strings.TrimSpace(strings.ToLower(word))

	// Try to get existing word card
	wordCard, err := s.wordRepo.GetWordCardByLemma(normalizedWord)
	if err != nil {
		return 0, fmt.Errorf("failed to get word card: %w", err)
	}

	if wordCard != nil {
		return wordCard.ID, nil
	}

	// Word not found, create minimal word card (only word, no LLM call)
	// Word card data will be filled asynchronously by TrainingWorker
	wordCardModel := &models.WordCard{
		Word:       normalizedWord,
		Definition: "", // Empty - will be filled later
		// All other fields are nil - will be filled asynchronously
	}

	wordCardID, err := s.wordRepo.UpsertWordCardLemma(wordCardModel)
	if err != nil {
		return 0, fmt.Errorf("failed to create minimal word card: %w", err)
	}

	s.logger.Debug("created minimal word card",
		zap.String("word", normalizedWord),
		zap.Int64("word_card_id", wordCardID),
	)

	return wordCardID, nil
}

// EnsureWordCardExists ensures a word card exists for a word
// Returns the word card ID
func (s *WordSetService) EnsureWordCardExists(ctx context.Context, word string) (int64, error) {
	normalizedWord := strings.TrimSpace(strings.ToLower(word))

	// Try to get existing word card
	wordCard, err := s.wordRepo.GetWordCardByLemma(normalizedWord)
	if err != nil {
		return 0, fmt.Errorf("failed to get word card: %w", err)
	}

	if wordCard != nil {
		return wordCard.ID, nil
	}

	// Word not found, need to create it via AI
	if s.aiService == nil {
		return 0, fmt.Errorf("AI service not available")
	}

	response, err := s.aiService.GenerateResponse(ctx, word)
	if err != nil {
		return 0, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Parse JSON response
	var wordInfo models.WordInfoResponse
	if err := json.Unmarshal([]byte(response), &wordInfo); err != nil {
		// Not JSON, save as legacy format
		if err := s.wordRepo.SaveWordCard(normalizedWord, response); err != nil {
			return 0, fmt.Errorf("failed to save word card: %w", err)
		}
		// Get the saved word card
		wordCard, err := s.wordRepo.GetWordCardByLemma(normalizedWord)
		if err != nil {
			return 0, fmt.Errorf("failed to get saved word card: %w", err)
		}
		if wordCard == nil {
			return 0, fmt.Errorf("word card not found after save")
		}
		return wordCard.ID, nil
	}

	models.SyncWordInfoResponseNeutralAliases(&wordInfo)

	// Check for error from LLM
	if wordInfo.Error.IsTrue() {
		return 0, fmt.Errorf("word rejected by LLM: %s", wordInfo.Error.Message)
	}

	// Save structured word card (same logic as WordService)
	lemma := strings.ToLower(wordInfo.Lemma)
	if lemma == "" {
		lemma = normalizedWord
	}

	displayEN := lemma
	if wordInfo.POS == "verb" && wordInfo.VerbForms != nil && wordInfo.VerbForms.V1 != "" {
		displayEN = "to " + wordInfo.VerbForms.V1
	}

	// Marshal examples and verb forms
	var examplesJSON *string
	if len(wordInfo.Examples) > 0 {
		examplesBytes, _ := json.Marshal(wordInfo.Examples)
		examplesStr := string(examplesBytes)
		examplesJSON = &examplesStr
	}

	var verbFormsJSON *string
	if wordInfo.VerbForms != nil {
		verbFormsBytes, _ := json.Marshal(wordInfo.VerbForms)
		verbFormsStr := string(verbFormsBytes)
		verbFormsJSON = &verbFormsStr
	}

	pos := wordInfo.POS
	transcription := wordInfo.Transcription
	definitionRU := wordInfo.DefinitionRU

	wordCardModel := &models.WordCard{
		Word:          lemma,
		Definition:    "", // Legacy field
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU:  &definitionRU,
		ExamplesJSON:  examplesJSON,
		VerbFormsJSON: verbFormsJSON,
		DisplayEN:     &displayEN,
	}

	wordCardID, err := s.wordRepo.UpsertWordCardLemma(wordCardModel)
	if err != nil {
		return 0, fmt.Errorf("failed to save word card: %w", err)
	}

	// Get the saved word card
	wordCard, err = s.wordRepo.GetWordCardByID(wordCardID)
	if err != nil {
		return 0, fmt.Errorf("failed to get saved word card: %w", err)
	}
	if wordCard == nil {
		return 0, fmt.Errorf("word card not found after save")
	}

	// Create word form mapping if needed
	if normalizedWord != strings.ToLower(wordCard.Word) {
		if err := s.wordRepo.UpsertWordFormMapping(normalizedWord, wordCard.ID); err != nil {
			s.logger.Warn("failed to create word form mapping", zap.Error(err))
		}
	}

	return wordCard.ID, nil
}

// EnsureTrainingCardsExist ensures training cards exist for a word card
// Uses the same logic as training worker
func (s *WordSetService) EnsureTrainingCardsExist(ctx context.Context, wordCardID int64) error {
	// Check if training cards already exist
	trainingCards, err := s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to check training cards: %w", err)
	}

	if len(trainingCards) > 0 {
		// Training cards already exist
		return nil
	}

	// Get word card
	wordCard, err := s.wordRepo.GetWordCardByID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to get word card: %w", err)
	}
	if wordCard == nil {
		return fmt.Errorf("word card not found")
	}

	// Generate training card via LLM
	if s.aiService == nil {
		return fmt.Errorf("AI service not available")
	}

	// Try to generate training card, first with default model, then with high model if validation fails
	var trainingResp models.TrainingCardResponse
	var response string
	var validationError string

	// First attempt with default model
	response, err = s.aiService.GenerateTrainingCard(ctx, wordCard.Word)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse response
	if err := json.Unmarshal([]byte(response), &trainingResp); err != nil {
		return fmt.Errorf("failed to parse LLM response: %w", err)
	}

	models.SyncTrainingCardResponseNeutralAliases(&trainingResp)

	// Check for error from LLM
	if trainingResp.Error != "" {
		return fmt.Errorf("word rejected by LLM: %s", trainingResp.Error)
	}

	// Validate response
	if len(trainingResp.Senses) == 0 {
		return fmt.Errorf("no senses in LLM response")
	}

	// Validate distractors
	validationError = ValidateTrainingCardResponse(wordCard, &trainingResp)
	if validationError != "" {
		// Validation failed - try with high model if available
		if s.modelHigh != "" {
			s.logger.Info("validation failed with default model, trying with high model",
				zap.String("word", wordCard.Word),
				zap.String("error", validationError),
				zap.String("high_model", s.modelHigh),
			)

			// Try with high model
			response, err = s.aiService.GenerateTrainingCard(ctx, wordCard.Word, s.modelHigh)
			if err != nil {
				s.logger.Warn("LLM generation with high model failed, using original validation error",
					zap.String("word", wordCard.Word),
					zap.Error(err),
				)
				return fmt.Errorf("validation failed: %s", validationError)
			}

			// Parse response from high model
			var highTrainingResp models.TrainingCardResponse
			if err := json.Unmarshal([]byte(response), &highTrainingResp); err != nil {
				s.logger.Warn("failed to parse LLM response from high model, using original validation error",
					zap.String("word", wordCard.Word),
					zap.Error(err),
				)
				return fmt.Errorf("validation failed: %s", validationError)
			}

			models.SyncTrainingCardResponseNeutralAliases(&highTrainingResp)

			// Check for error from LLM
			if highTrainingResp.Error != "" {
				return fmt.Errorf("word rejected by high model LLM: %s", highTrainingResp.Error)
			}

			// Validate response from high model
			if len(highTrainingResp.Senses) == 0 {
				s.logger.Warn("no senses in high model LLM response, using original validation error",
					zap.String("word", wordCard.Word),
				)
				return fmt.Errorf("validation failed: %s", validationError)
			}

			// Validate distractors from high model
			highValidationError := ValidateTrainingCardResponse(wordCard, &highTrainingResp)
			if highValidationError == "" {
				// High model validation passed - use this response
				s.logger.Info("validation passed with high model",
					zap.String("word", wordCard.Word),
					zap.String("high_model", s.modelHigh),
				)
				trainingResp = highTrainingResp
				// Continue with creating training cards using high model response
			} else {
				s.logger.Warn("validation also failed with high model",
					zap.String("word", wordCard.Word),
					zap.String("error", highValidationError),
				)
				return fmt.Errorf("validation failed: %s", highValidationError)
			}
		} else {
			// High model not available, return original validation error
			return fmt.Errorf("validation failed: %s", validationError)
		}
	}

	// Create training cards
	for i, sense := range trainingResp.Senses {
		// Marshal distractors
		distractorsRU, _ := json.Marshal(sense.DistractorsRU)
		distractorsEN, _ := json.Marshal(sense.DistractorsEN)

		// Determine display_word
		displayWord := trainingResp.WordEN
		if sense.DisplayWord != "" {
			displayWord = sense.DisplayWord
		}

		// Get POS
		pos := sense.POS
		if pos == "" && wordCard.POS != nil {
			pos = *wordCard.POS
		}

		trainingCard := &models.TrainingCard{
			WordCardID:    wordCardID,
			WordEN:        displayWord,
			Transcription: trainingResp.Transcription,
			SenseIndex:    i,
			WordRU:        sense.WordRU,
			MeaningEN:     sense.MeaningEN,
			ExampleEN:     sense.ExampleEN,
			ExampleRU:     sense.ExampleRU,
			DistractorsRU: string(distractorsRU),
			DistractorsEN: string(distractorsEN),
			Hint:          sense.Hint,
		}

		if pos != "" {
			trainingCard.POS = &pos
		}
		if displayWord != "" {
			trainingCard.DisplayWord = &displayWord
		}

		_, err := s.trainingCardRepo.CreateTrainingCard(trainingCard)
		if err != nil {
			return fmt.Errorf("failed to create training card: %w", err)
		}
	}

	return nil
}

// EnsureUserCardsForWord creates user_cards for all training_cards of a word
// Similar to WordService.ensureUserCardsForWord
func (s *WordSetService) EnsureUserCardsForWord(userID, wordCardID int64) error {
	// Get all training cards for this word
	trainingCards, err := s.trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
	if err != nil {
		return fmt.Errorf("failed to get training cards: %w", err)
	}

	if len(trainingCards) == 0 {
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
		s.logger.Info("created user cards for word",
			zap.Int64("user_id", userID),
			zap.Int64("word_card_id", wordCardID),
			zap.Int("training_cards", len(trainingCards)),
			zap.Int("user_cards_created", createdCount),
		)
		if s.userWordMasteringRepo != nil {
			if err := s.userWordMasteringRepo.Upsert(userID, wordCardID, 0); err != nil {
				s.logger.Warn("failed to upsert initial mastering score", zap.Error(err))
			}
		}
	}

	return nil
}

// MarkKnown marks a word as known and removes user_cards
func (s *WordSetService) MarkKnown(userID, wordCardID int64) error {
	// Mark as known
	if err := s.userWordKnowledgeRepo.MarkKnown(userID, wordCardID); err != nil {
		return fmt.Errorf("failed to mark as known: %w", err)
	}

	// Delete user_cards for this word
	_, err := s.userCardRepo.DeleteUserCardsByWordCardIDForUser(userID, wordCardID)
	if err != nil {
		s.logger.Warn("failed to delete user cards",
			zap.Int64("user_id", userID),
			zap.Int64("word_card_id", wordCardID),
			zap.Error(err),
		)
		// Don't fail if deletion fails, knowledge is already marked
	}

	return nil
}

// ProcessWordSetItems processes a comma-separated list of words for a word set
// Creates word_cards with minimal data (only word) and creates word_set_items
// Word card data and training cards are created asynchronously by TrainingWorker
// This allows saving word sets quickly even with many words
func (s *WordSetService) ProcessWordSetItems(ctx context.Context, wordSetID int64, wordsStr string) error {
	// Split by comma and normalize
	words := strings.Split(wordsStr, ",")
	var wordCardIDs []int64

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		// Ensure word card exists with minimal data (no LLM call)
		// Word card data will be filled asynchronously by TrainingWorker
		wordCardID, err := s.EnsureWordCardExistsMinimal(word)
		if err != nil {
			s.logger.Warn("failed to ensure minimal word card exists",
				zap.String("word", word),
				zap.Error(err),
			)
			// Continue with other words
			continue
		}

		// Word card data and training cards will be created asynchronously by TrainingWorker
		// This allows saving word sets quickly even with many words

		wordCardIDs = append(wordCardIDs, wordCardID)
	}

	// Remove duplicates
	seen := make(map[int64]bool)
	var uniqueWordCardIDs []int64
	for _, id := range wordCardIDs {
		if !seen[id] {
			seen[id] = true
			uniqueWordCardIDs = append(uniqueWordCardIDs, id)
		}
	}

	// Set word set items
	if err := s.wordSetRepo.SetWordSetItems(wordSetID, uniqueWordCardIDs); err != nil {
		return fmt.Errorf("failed to set word set items: %w", err)
	}

	return nil
}
