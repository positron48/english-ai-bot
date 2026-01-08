package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// TrainingWorker handles background generation of training cards
type TrainingWorker struct {
	aiService        *ai.Service
	wordRepo         *repository.WordRepository
	trainingCardRepo *repository.TrainingCardRepository
	userCardRepo     *repository.UserCardRepository
	userRepo         *repository.UserRepository
	cbService        *CircuitBreakerService
	bot              *tgbotapi.BotAPI
	adminTelegramID  int64
	batchSize        int
	interval         time.Duration
	logger           *zap.Logger
	stopChan         chan struct{}
}

// NewTrainingWorker creates a new training worker
func NewTrainingWorker(
	aiService *ai.Service,
	wordRepo *repository.WordRepository,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo *repository.UserCardRepository,
	userRepo *repository.UserRepository,
	cbService *CircuitBreakerService,
	bot *tgbotapi.BotAPI,
	adminTelegramID int64,
	batchSize int,
	interval time.Duration,
	logger *zap.Logger,
) *TrainingWorker {
	return &TrainingWorker{
		aiService:        aiService,
		wordRepo:         wordRepo,
		trainingCardRepo: trainingCardRepo,
		userCardRepo:     userCardRepo,
		userRepo:         userRepo,
		cbService:        cbService,
		bot:              bot,
		adminTelegramID:  adminTelegramID,
		batchSize:        batchSize,
		interval:         interval,
		logger:           logger,
		stopChan:         make(chan struct{}),
	}
}

// Start starts the worker
func (w *TrainingWorker) Start(ctx context.Context) {
	w.logger.Info("starting training worker",
		zap.Int("batch_size", w.batchSize),
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

	// Get pending cards
	cards, err := w.trainingCardRepo.GetWordCardsWithoutTrainingCards(w.batchSize)
	if err != nil {
		w.logger.Error("failed to get pending cards", zap.Error(err))
		return
	}

	if len(cards) == 0 {
		return
	}

	w.logger.Info("processing word cards",
		zap.Int("count", len(cards)),
	)

	// Process each card
	for _, card := range cards {
		if err := w.processCard(ctx, card); err != nil {
			w.logger.Error("failed to process card",
				zap.String("word", card.Word),
				zap.Error(err),
			)
			
			// Record failure
			if err := w.cbService.RecordFailure(err.Error()); err != nil {
				w.logger.Error("failed to record circuit breaker failure", zap.Error(err))
			}
			
			// Check if circuit breaker opened
			if isOpen, _ := w.cbService.IsOpen(); isOpen {
				w.logger.Warn("circuit breaker opened, notifying admin")
				w.notifyAdmin(err.Error())
				return
			}
		} else {
			// Record success
			if err := w.cbService.RecordSuccess(); err != nil {
				w.logger.Error("failed to record circuit breaker success", zap.Error(err))
			}
		}
	}
}

// processCard processes a single word card
func (w *TrainingWorker) processCard(ctx context.Context, wordCard *models.WordCard) error {
	w.logger.Info("generating training card",
		zap.String("word", wordCard.Word),
		zap.Int64("word_card_id", wordCard.ID),
	)

	// Generate training card via LLM
	response, err := w.aiService.GenerateTrainingCard(ctx, wordCard.Word)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse response
	var trainingResp models.TrainingCardResponse
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
	if validationError := validateTrainingCardResponse(wordCard, &trainingResp); validationError != "" {
		w.logger.Warn("training card validation failed",
			zap.String("word", wordCard.Word),
			zap.Int64("word_card_id", wordCard.ID),
			zap.String("error", validationError),
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

