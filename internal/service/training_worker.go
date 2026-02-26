package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// TrainingWorker handles background generation of training cards
type TrainingWorker struct {
	aiService            *ai.Service
	wordRepo             *repository.WordRepository
	trainingCardRepo     *repository.TrainingCardRepository
	userCardRepo         *repository.UserCardRepository
	userRepo             *repository.UserRepository
	pronunciationService *PronunciationService
	cbService            *CircuitBreakerService
	bot                  *tgbotapi.BotAPI
	adminTelegramID      int64
	batchSize            int
	llmWorkers           int
	interval             time.Duration
	modelHigh            string
	logger               *zap.Logger
	stopChan             chan struct{}
}

// NewTrainingWorker creates a new training worker
func NewTrainingWorker(
	aiService *ai.Service,
	wordRepo *repository.WordRepository,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo *repository.UserCardRepository,
	userRepo *repository.UserRepository,
	pronunciationService *PronunciationService,
	cbService *CircuitBreakerService,
	bot *tgbotapi.BotAPI,
	adminTelegramID int64,
	batchSize int,
	llmWorkers int,
	interval time.Duration,
	modelHigh string,
	logger *zap.Logger,
) *TrainingWorker {
	return &TrainingWorker{
		aiService:            aiService,
		wordRepo:             wordRepo,
		trainingCardRepo:     trainingCardRepo,
		userCardRepo:         userCardRepo,
		userRepo:             userRepo,
		pronunciationService: pronunciationService,
		cbService:            cbService,
		bot:                  bot,
		adminTelegramID:      adminTelegramID,
		batchSize:            batchSize,
		llmWorkers:           llmWorkers,
		interval:             interval,
		modelHigh:            modelHigh,
		logger:               logger,
		stopChan:             make(chan struct{}),
	}
}

// Start starts the worker
func (w *TrainingWorker) Start(ctx context.Context) {
	w.logger.Info("starting training worker",
		zap.Int("batch_size", w.batchSize),
		zap.Int("llm_workers", w.llmWorkers),
		zap.Duration("interval", w.interval),
	)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("training worker stopped")
			return
		case <-w.stopChan:
			w.logger.Info("training worker stopped")
			return
		case <-ticker.C:
			w.processCards(ctx)
		}
	}
}

// Stop stops the worker
func (w *TrainingWorker) Stop() {
	close(w.stopChan)
}

// processCards processes a batch of cards
func (w *TrainingWorker) processCards(ctx context.Context) {
	// Check circuit breaker
	isOpen, err := w.cbService.IsOpen()
	if err != nil {
		w.logger.Error("failed to check circuit breaker", zap.Error(err))
		return
	}
	if isOpen {
		w.logger.Debug("circuit breaker is open, skipping iteration")
		return
	}

	// Determine how many cards to fetch
	// We need at least as many cards as workers to utilize all workers
	// But also respect batchSize as a minimum to avoid too many small batches
	cardsToFetch := w.batchSize
	if w.llmWorkers > 0 && w.llmWorkers > cardsToFetch {
		cardsToFetch = w.llmWorkers
	}

	// Get pending cards
	cards, err := w.trainingCardRepo.GetWordCardsWithoutTrainingCards(cardsToFetch)
	if err != nil {
		w.logger.Error("failed to get pending cards", zap.Error(err))
		return
	}

	if len(cards) == 0 {
		return
	}

	w.logger.Info("processing word cards",
		zap.Int("count", len(cards)),
		zap.Int("workers", w.llmWorkers),
		zap.Int("cards_fetched", cardsToFetch),
	)

	// Use worker pool for parallel processing
	// Each worker processes cards in parallel, and for each card:
	// 1. Fills word card data via LLM (fillWordCardData -> GenerateResponse)
	// 2. Generates training cards via LLM (GenerateTrainingCard)
	// Both LLM requests are parallelized across different cards
	workers := w.llmWorkers
	if workers <= 0 {
		workers = 1 // Fallback to sequential if invalid
	}
	if workers > len(cards) {
		workers = len(cards) // Don't create more workers than cards
	}

	w.logger.Info("creating worker pool",
		zap.Int("requested_workers", w.llmWorkers),
		zap.Int("actual_workers", workers),
		zap.Int("cards_to_process", len(cards)),
	)

	// Create channels for work distribution
	cardChan := make(chan *models.WordCard, len(cards))
	resultChan := make(chan error, len(cards))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for card := range cardChan {
				err := w.processCard(ctx, card)
				resultChan <- err
			}
		}()
	}

	// Send cards to workers
	for _, card := range cards {
		cardChan <- card
	}
	close(cardChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	circuitBreakerOpened := false
	for err := range resultChan {
		if err != nil {
			w.logger.Error("failed to process card",
				zap.Error(err),
			)

			// Record failure
			if err := w.cbService.RecordFailure(err.Error()); err != nil {
				w.logger.Error("failed to record circuit breaker failure", zap.Error(err))
			}

			// Check if circuit breaker opened
			if isOpen, _ := w.cbService.IsOpen(); isOpen {
				if !circuitBreakerOpened {
					w.logger.Warn("circuit breaker opened, notifying admin")
					w.notifyAdmin(err.Error())
					circuitBreakerOpened = true
				}
				// Continue collecting results but don't process more
			}
		} else {
			// Record success only if circuit breaker is still closed
			if !circuitBreakerOpened {
				if err := w.cbService.RecordSuccess(); err != nil {
					w.logger.Error("failed to record circuit breaker success", zap.Error(err))
				}
			}
		}
	}
}

// hasMissingData checks if word card has any missing data fields (POS, Transcription, or DefinitionRU)
func (w *TrainingWorker) hasMissingData(wordCard *models.WordCard) bool {
	return wordCard.POS == nil || wordCard.Transcription == nil || wordCard.DefinitionRU == nil
}

// fillWordCardData fills missing word card data via LLM
func (w *TrainingWorker) fillWordCardData(ctx context.Context, wordCard *models.WordCard) error {
	if !w.hasMissingData(wordCard) {
		// Word card already has all data, skip
		return nil
	}

	w.logger.Info("filling missing word card data",
		zap.String("word", wordCard.Word),
		zap.Int64("word_card_id", wordCard.ID),
		zap.Bool("missing_pos", wordCard.POS == nil),
		zap.Bool("missing_transcription", wordCard.Transcription == nil),
		zap.Bool("missing_definition_ru", wordCard.DefinitionRU == nil),
	)

	// Generate word card data via LLM
	response, err := w.aiService.GenerateResponse(ctx, wordCard.Word)
	if err != nil {
		return fmt.Errorf("failed to get AI response: %w", err)
	}

	// Parse JSON response
	var wordInfo models.WordInfoResponse
	if err := json.Unmarshal([]byte(response), &wordInfo); err != nil {
		// Not JSON, skip filling (word card will remain as is)
		w.logger.Warn("failed to parse AI response as JSON, skipping word card fill",
			zap.String("word", wordCard.Word),
			zap.Error(err),
		)
		return nil
	}

	// Check for error from LLM
	if wordInfo.Error.IsTrue() {
		// Word rejected by LLM, skip filling
		w.logger.Info("word rejected by LLM, skipping word card fill",
			zap.String("word", wordCard.Word),
		)
		return nil
	}

	// Prepare updated word card model, preserving existing values
	wordCardModel := &models.WordCard{
		ID:            wordCard.ID,
		Word:          wordCard.Word,          // Keep existing word
		Definition:    wordCard.Definition,    // Keep existing definition
		POS:           wordCard.POS,           // Will update if nil
		Transcription: wordCard.Transcription, // Will update if nil
		DefinitionRU:  wordCard.DefinitionRU,  // Will update if nil
		ExamplesJSON:  wordCard.ExamplesJSON,  // Will update if empty
		VerbFormsJSON: wordCard.VerbFormsJSON, // Will update if empty
		DisplayEN:     wordCard.DisplayEN,     // Will update if empty
	}

	// Fill only missing fields (preserve existing values)
	if wordCard.POS == nil && wordInfo.POS != "" {
		pos := wordInfo.POS
		wordCardModel.POS = &pos
	}

	if wordCard.Transcription == nil && wordInfo.Transcription != "" {
		transcription := wordInfo.Transcription
		wordCardModel.Transcription = &transcription
	}

	if wordCard.DefinitionRU == nil && wordInfo.DefinitionRU != "" {
		definitionRU := wordInfo.DefinitionRU
		wordCardModel.DefinitionRU = &definitionRU
	}

	// Update examples if missing and new data available
	if (wordCard.ExamplesJSON == nil || *wordCard.ExamplesJSON == "") && len(wordInfo.Examples) > 0 {
		examplesBytes, _ := json.Marshal(wordInfo.Examples)
		examplesStr := string(examplesBytes)
		wordCardModel.ExamplesJSON = &examplesStr
	}

	// Update verb forms if missing and new data available
	if (wordCard.VerbFormsJSON == nil || *wordCard.VerbFormsJSON == "") && wordInfo.VerbForms != nil {
		verbFormsBytes, _ := json.Marshal(wordInfo.VerbForms)
		verbFormsStr := string(verbFormsBytes)
		wordCardModel.VerbFormsJSON = &verbFormsStr
	}

	// Update display_en if missing and new data available
	if wordCard.DisplayEN == nil || *wordCard.DisplayEN == "" {
		lemma := strings.ToLower(wordInfo.Lemma)
		if lemma == "" {
			lemma = strings.ToLower(wordCard.Word)
		}

		displayEN := lemma
		if wordCardModel.POS != nil && *wordCardModel.POS == "verb" && wordInfo.VerbForms != nil && wordInfo.VerbForms.V1 != "" {
			displayEN = "to " + wordInfo.VerbForms.V1
		} else if wordInfo.POS == "verb" && wordInfo.VerbForms != nil && wordInfo.VerbForms.V1 != "" {
			displayEN = "to " + wordInfo.VerbForms.V1
		}

		wordCardModel.DisplayEN = &displayEN
	}

	// Update word card
	if err := w.wordRepo.UpdateWordCard(wordCardModel); err != nil {
		return fmt.Errorf("failed to update word card: %w", err)
	}

	w.logger.Info("filled missing word card data",
		zap.String("word", wordCard.Word),
		zap.Int64("word_card_id", wordCard.ID),
	)

	return nil
}

// processCard processes a single word card
func (w *TrainingWorker) processCard(ctx context.Context, wordCard *models.WordCard) error {
	w.logger.Info("generating training card",
		zap.String("word", wordCard.Word),
		zap.Int64("word_card_id", wordCard.ID),
	)

	// Fill word card data if it has minimal data
	if err := w.fillWordCardData(ctx, wordCard); err != nil {
		w.logger.Warn("failed to fill word card data, continuing anyway",
			zap.String("word", wordCard.Word),
			zap.Int64("word_card_id", wordCard.ID),
			zap.Error(err),
		)
		// Continue anyway - training cards can still be created
	}

	// Reload word card to get updated data
	updatedWordCard, err := w.wordRepo.GetWordCardByID(wordCard.ID)
	if err != nil {
		w.logger.Warn("failed to reload word card, using original",
			zap.String("word", wordCard.Word),
			zap.Error(err),
		)
		// Continue with original wordCard
	} else if updatedWordCard != nil {
		wordCard = updatedWordCard
	}

	// Try to generate training card, first with default model, then with high model if validation fails
	var trainingResp models.TrainingCardResponse
	var response string
	var validationError string
	triedHighModel := false

	// First attempt with default model
	response, err = w.aiService.GenerateTrainingCard(ctx, wordCard.Word)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse response
	if err := json.Unmarshal([]byte(response), &trainingResp); err != nil {
		// Parsing error - don't mark as processed, allow retry
		return fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Check for error from LLM (e.g., word is not English, proper noun, non-existent)
	if trainingResp.Error != "" {
		// LLM explicitly rejected the word - mark as processed with error
		// This is not a transient error, so we mark it to prevent infinite retries
		// Don't trigger circuit breaker as this is expected behavior
		err := w.wordRepo.MarkWordCardProcessedError(wordCard.ID, trainingResp.Error)
		if err != nil {
			w.logger.Error("failed to mark word card as processed with error",
				zap.Int64("word_card_id", wordCard.ID),
				zap.String("error", trainingResp.Error),
				zap.Error(err),
			)
			// Still return nil to avoid circuit breaker, but log the error
		} else {
			w.logger.Info("word card rejected by LLM",
				zap.String("word", wordCard.Word),
				zap.String("error", trainingResp.Error),
			)
		}
		// Return nil (not an error) to avoid triggering circuit breaker
		return nil
	}

	// Validate response
	if len(trainingResp.Senses) == 0 {
		return fmt.Errorf("no senses in LLM response")
	}

	// Validate distractors according to rules
	validationError = ValidateTrainingCardResponse(wordCard, &trainingResp)
	if validationError != "" {
		// Validation failed - try with high model if available
		if w.modelHigh != "" {
			w.logger.Info("validation failed with default model, trying with high model",
				zap.String("word", wordCard.Word),
				zap.String("error", validationError),
				zap.String("high_model", w.modelHigh),
			)

			// Try with high model
			response, err = w.aiService.GenerateTrainingCard(ctx, wordCard.Word, w.modelHigh)
			if err != nil {
				w.logger.Warn("LLM generation with high model failed, using original validation error",
					zap.String("word", wordCard.Word),
					zap.Error(err),
				)
			} else {
				// Parse response from high model
				var highTrainingResp models.TrainingCardResponse
				if err := json.Unmarshal([]byte(response), &highTrainingResp); err != nil {
					w.logger.Warn("failed to parse LLM response from high model, using original validation error",
						zap.String("word", wordCard.Word),
						zap.Error(err),
					)
				} else {
					// Check for error from LLM
					if highTrainingResp.Error != "" {
						w.logger.Info("word rejected by high model LLM",
							zap.String("word", wordCard.Word),
							zap.String("error", highTrainingResp.Error),
						)
						err := w.wordRepo.MarkWordCardProcessedError(wordCard.ID, highTrainingResp.Error)
						if err != nil {
							w.logger.Error("failed to mark word card as processed with error", zap.Error(err))
						}
						return nil
					}

					// Validate response from high model
					if len(highTrainingResp.Senses) == 0 {
						w.logger.Warn("no senses in high model LLM response, using original validation error",
							zap.String("word", wordCard.Word),
						)
					} else {
						// Validate distractors from high model
						highValidationError := ValidateTrainingCardResponse(wordCard, &highTrainingResp)
						if highValidationError == "" {
							// High model validation passed - use this response
							w.logger.Info("validation passed with high model",
								zap.String("word", wordCard.Word),
								zap.String("high_model", w.modelHigh),
							)
							trainingResp = highTrainingResp
							validationError = ""
							triedHighModel = true
						} else {
							w.logger.Warn("validation also failed with high model",
								zap.String("word", wordCard.Word),
								zap.String("error", highValidationError),
							)
							validationError = highValidationError
						}
					}
				}
			}
		}

		// If validation still failed after trying high model (or high model not available)
		if validationError != "" {
			w.logger.Warn("training card validation failed",
				zap.String("word", wordCard.Word),
				zap.Int64("word_card_id", wordCard.ID),
				zap.String("error", validationError),
				zap.Bool("tried_high_model", triedHighModel),
			)
			// Mark as processed with error - don't create cards, don't trigger circuit breaker
			err := w.wordRepo.MarkWordCardProcessedError(wordCard.ID, validationError)
			if err != nil {
				w.logger.Error("failed to mark word card as processed with error",
					zap.Int64("word_card_id", wordCard.ID),
					zap.String("error", validationError),
					zap.Error(err),
				)
				// Still return nil to avoid circuit breaker
			} else {
				w.logger.Info("word card rejected due to validation failure",
					zap.String("word", wordCard.Word),
					zap.String("error", validationError),
				)
			}
			// Notify admin about validation error
			w.notifyAdminValidationError(wordCard.Word, validationError)
			// Return nil (not an error) to avoid triggering circuit breaker
			return nil
		}
	}

	w.logger.Info("parsed training card response",
		zap.String("word", wordCard.Word),
		zap.Int("senses", len(trainingResp.Senses)),
	)

	// Create training cards
	trainingCardIDs := make([]int64, 0, len(trainingResp.Senses))

	for i, sense := range trainingResp.Senses {
		// Marshal distractors
		distractorsRU, _ := json.Marshal(sense.DistractorsRU)
		distractorsEN, _ := json.Marshal(sense.DistractorsEN)

		// Determine display_word for this sense: use from sense, or fallback to word_en
		displayWord := trainingResp.WordEN
		if sense.DisplayWord != "" {
			displayWord = sense.DisplayWord
		}

		// Get POS for this sense: use from sense, or fallback to word_card if available
		pos := sense.POS
		if pos == "" {
			// Try to get from word_card if available
			if wordCard.POS != nil {
				pos = *wordCard.POS
			}
		}

		trainingCard := &models.TrainingCard{
			WordCardID:    wordCard.ID,
			WordEN:        displayWord, // For backward compatibility
			Transcription: trainingResp.Transcription,
			SenseIndex:    i, // Use index from loop instead of sense.Index
			WordRU:        sense.WordRU,
			MeaningEN:     sense.MeaningEN,
			ExampleEN:     sense.ExampleEN,
			ExampleRU:     sense.ExampleRU,
			DistractorsRU: string(distractorsRU),
			DistractorsEN: string(distractorsEN),
			Hint:          sense.Hint,
		}

		// Set POS and display_word from sense
		if pos != "" {
			trainingCard.POS = &pos
		}
		if displayWord != "" {
			trainingCard.DisplayWord = &displayWord
		}

		id, err := w.trainingCardRepo.CreateTrainingCard(trainingCard)
		if err != nil {
			return fmt.Errorf("failed to create training card: %w", err)
		}

		if w.pronunciationService != nil {
			// Prefetch pronunciation for display form and base word in background.
			w.pronunciationService.ScheduleWords(displayWord, trainingResp.WordEN)
		}

		trainingCardIDs = append(trainingCardIDs, id)
	}

	// Get users who requested this word
	users, err := w.getUsersForWord(wordCard.Word)
	if err != nil {
		w.logger.Warn("failed to get users for word", zap.Error(err))
		// Don't fail the whole process
		users = []*models.User{}
	}

	w.logger.Info("found users for word",
		zap.String("word", wordCard.Word),
		zap.Int("user_count", len(users)),
	)

	// Create user_cards for each user and training card
	createdCount := 0
	for _, user := range users {
		for _, trainingCardID := range trainingCardIDs {
			// Create ru_en card
			ruEnCard := &models.UserCard{
				UserID:         user.ID,
				TrainingCardID: trainingCardID,
				Direction:      models.DirectionRUtoEN,
				State:          models.StateNew,
				EF:             models.InitialEF,
			}
			if _, err := w.userCardRepo.CreateUserCard(ruEnCard); err != nil {
				w.logger.Warn("failed to create ru_en user card",
					zap.Int64("user_id", user.ID),
					zap.Error(err),
				)
			} else {
				createdCount++
			}

			// Create en_ru card
			enRuCard := &models.UserCard{
				UserID:         user.ID,
				TrainingCardID: trainingCardID,
				Direction:      models.DirectionENtoRU,
				State:          models.StateNew,
				EF:             models.InitialEF,
			}
			if _, err := w.userCardRepo.CreateUserCard(enRuCard); err != nil {
				w.logger.Warn("failed to create en_ru user card",
					zap.Int64("user_id", user.ID),
					zap.Error(err),
				)
			} else {
				createdCount++
			}
		}
	}

	w.logger.Info("successfully created training cards",
		zap.String("word", wordCard.Word),
		zap.Int("training_cards", len(trainingCardIDs)),
		zap.Int("users", len(users)),
		zap.Int("user_cards_created", createdCount),
	)

	return nil
}

// getUsersForWord gets users who requested this word
func (w *TrainingWorker) getUsersForWord(word string) ([]*models.User, error) {
	// Get user IDs who requested this word from word_request_history
	// Note: user_id in word_request_history can be either:
	// - Internal user_id (from web chat via JWT)
	// - Telegram_id (from telegram bot - legacy)
	userIDs, err := w.wordRepo.GetUserIDsByWord(word)
	if err != nil {
		return nil, fmt.Errorf("failed to get user IDs: %w", err)
	}

	// Get users by internal ID first, then fallback to telegram_id for backward compatibility
	users := make([]*models.User, 0, len(userIDs))
	seenUserIDs := make(map[int64]bool) // Track already added users to avoid duplicates

	for _, id := range userIDs {
		// Skip if we already added this user
		if seenUserIDs[id] {
			continue
		}

		// First, try to get user by internal ID (for web chat users)
		user, err := w.userRepo.GetUserByID(id)
		if err != nil {
			w.logger.Warn("failed to get user by ID",
				zap.Int64("user_id", id),
				zap.Error(err),
			)
			// Fallback: try as telegram_id (for legacy telegram bot entries)
			user, err = w.userRepo.GetOrCreateUser(id)
			if err != nil {
				w.logger.Warn("failed to get/create user by telegram_id",
					zap.Int64("id", id),
					zap.Error(err),
				)
				continue
			}
		}

		if user != nil {
			users = append(users, user)
			seenUserIDs[user.ID] = true
			// Also mark telegram_id as seen to avoid duplicates
			if user.TelegramID > 0 {
				seenUserIDs[user.TelegramID] = true
			}
		}
	}

	return users, nil
}

// notifyAdmin sends a notification to the admin about circuit breaker
func (w *TrainingWorker) notifyAdmin(errorMessage string) {
	if w.adminTelegramID == 0 {
		w.logger.Debug("admin telegram ID not set, skipping notification")
		return
	}

	_, failureCount, lastError, err := w.cbService.GetState()
	if err != nil {
		w.logger.Error("failed to get circuit breaker state", zap.Error(err))
		return
	}

	message := fmt.Sprintf(
		"⚠️ Circuit Breaker ОТКРЫТ\n\n"+
			"Воркер генерации карточек остановлен.\n"+
			"Причина: %d последовательных ошибок LLM.\n\n"+
			"Последняя ошибка: %s\n"+
			"Время: %s\n\n"+
			"Для сброса используйте /reset_circuit",
		failureCount,
		lastError,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if w.bot == nil {
		w.logger.Warn("cannot send admin notification: Telegram bot not initialized",
			zap.Int64("admin_id", w.adminTelegramID),
		)
		return
	}

	msg := tgbotapi.NewMessage(w.adminTelegramID, message)
	if _, err := w.bot.Send(msg); err != nil {
		w.logger.Error("failed to send admin notification", zap.Error(err))
	} else {
		w.logger.Info("sent circuit breaker notification to admin",
			zap.Int64("admin_id", w.adminTelegramID),
		)
	}
}

// notifyAdminValidationError sends a notification to the admin about validation error
func (w *TrainingWorker) notifyAdminValidationError(word, validationError string) {
	if w.adminTelegramID == 0 {
		w.logger.Debug("admin telegram ID not set, skipping validation error notification")
		return
	}

	// Truncate error message if too long for Telegram
	errorMsg := validationError
	if len(errorMsg) > 500 {
		errorMsg = errorMsg[:500] + "..."
	}

	message := fmt.Sprintf(
		"⚠️ Ошибка валидации карточки\n\n"+
			"Слово: %s\n"+
			"Ошибка валидации distractors:\n%s\n\n"+
			"Время: %s",
		word,
		errorMsg,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if w.bot == nil {
		w.logger.Warn("cannot send admin notification: Telegram bot not initialized",
			zap.Int64("admin_id", w.adminTelegramID),
		)
		return
	}

	msg := tgbotapi.NewMessage(w.adminTelegramID, message)
	if _, err := w.bot.Send(msg); err != nil {
		w.logger.Error("failed to send validation error notification to admin", zap.Error(err))
	} else {
		w.logger.Info("sent validation error notification to admin",
			zap.Int64("admin_id", w.adminTelegramID),
			zap.String("word", word),
		)
	}
}
