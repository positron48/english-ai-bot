package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// TrainingService handles training session management
type TrainingService struct {
	userCardRepo     *repository.UserCardRepository
	trainingCardRepo *repository.TrainingCardRepository
	sessionRepo      *repository.SessionRepository
	logger           *zap.Logger
}

// NewTrainingService creates a new training service
func NewTrainingService(
	userCardRepo *repository.UserCardRepository,
	trainingCardRepo *repository.TrainingCardRepository,
	sessionRepo *repository.SessionRepository,
	logger *zap.Logger,
) *TrainingService {
	return &TrainingService{
		userCardRepo:     userCardRepo,
		trainingCardRepo: trainingCardRepo,
		sessionRepo:      sessionRepo,
		logger:           logger,
	}
}

// SessionConfig holds session configuration
type SessionConfig struct {
	MaxCardsPerSession int
	MaxNewPerSession   int
	AlgoVersion        string
}

// StartSession starts a new training session
func (s *TrainingService) StartSession(userID int64, source models.SessionSource) (*models.TrainingSession, []*models.UserCardWithTraining, error) {
	// Check if there's already an active session
	activeSession, err := s.sessionRepo.GetActiveSession(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check active session: %w", err)
	}
	if activeSession != nil {
		// Finish the old session
		s.logger.Info("finishing old active session", zap.Int64("session_id", activeSession.ID))
		if err := s.sessionRepo.FinishSession(activeSession.ID, activeSession.DoneCount); err != nil {
			s.logger.Warn("failed to finish old session", zap.Error(err))
		}
	}

	// Get session configuration
	config := SessionConfig{
		MaxCardsPerSession: models.DefaultMaxCardsPerSession,
		MaxNewPerSession:   models.DefaultMaxNewPerSession,
		AlgoVersion:        "srs_v2_delayed_mcq_sm2_autoquality",
	}

	// Generate card queue
	queue, err := s.generateQueue(userID, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate queue: %w", err)
	}

	if len(queue) == 0 {
		return nil, nil, fmt.Errorf("no cards available for training")
	}

	// Create session
	configJSON, _ := json.Marshal(config)
	session := &models.TrainingSession{
		UserID:       userID,
		Source:       source,
		PlannedCount: len(queue),
		DoneCount:    0,
		SessionJSON:  string(configJSON),
	}

	sessionID, err := s.sessionRepo.CreateSession(session)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	session.ID = sessionID

	s.logger.Info("started training session",
		zap.Int64("session_id", sessionID),
		zap.Int64("user_id", userID),
		zap.String("source", string(source)),
		zap.Int("planned_count", len(queue)),
	)

	return session, queue, nil
}

// FinishSession finishes a training session
func (s *TrainingService) FinishSession(sessionID int64, doneCount int) error {
	return s.sessionRepo.FinishSession(sessionID, doneCount)
}

// generateQueue generates a queue of cards for training
func (s *TrainingService) generateQueue(userID int64, config SessionConfig) ([]*models.UserCardWithTraining, error) {
	now := time.Now()

	// Clean up orphaned user_cards before getting cards
	if _, err := s.userCardRepo.DeleteOrphanedUserCards(); err != nil {
		s.logger.Warn("failed to clean up orphaned user cards", zap.Error(err))
	}

	// Get due cards (learning + review)
	dueCards, err := s.userCardRepo.GetDueCards(userID, now, config.MaxCardsPerSession)
	if err != nil {
		return nil, fmt.Errorf("failed to get due cards: %w", err)
	}

	s.logger.Info("got due cards for user",
		zap.Int64("user_id", userID),
		zap.Int("due_count", len(dueCards)),
	)

	// Get new cards if we have space
	remainingSlots := config.MaxCardsPerSession - len(dueCards)
	var newCards []*models.UserCard
	if remainingSlots > 0 {
		maxNew := config.MaxNewPerSession
		if remainingSlots < maxNew {
			maxNew = remainingSlots
		}
		newCards, err = s.userCardRepo.GetNewCards(userID, maxNew)
		if err != nil {
			return nil, fmt.Errorf("failed to get new cards: %w", err)
		}
		s.logger.Info("got new cards for user",
			zap.Int64("user_id", userID),
			zap.Int("new_count", len(newCards)),
		)
	}

	// Combine cards
	allCards := append(dueCards, newCards...)

	if len(allCards) == 0 {
		return nil, nil
	}

	// Sort: learning first, then review, then new
	s.sortCards(allCards, now)

	// Fetch training card data
	queue := make([]*models.UserCardWithTraining, 0, len(allCards))
	for _, userCard := range allCards {
		trainingCard, err := s.trainingCardRepo.GetTrainingCard(userCard.TrainingCardID)
		if err != nil {
			s.logger.Warn("failed to get training card", zap.Error(err))
			continue
		}
		if trainingCard == nil {
			s.logger.Warn("training card not found", zap.Int64("id", userCard.TrainingCardID))
			continue
		}

		queue = append(queue, &models.UserCardWithTraining{
			UserCard:     *userCard,
			TrainingCard: *trainingCard,
		})
	}

	// First, shuffle all cards randomly
	rand.Shuffle(len(queue), func(i, j int) {
		queue[i], queue[j] = queue[j], queue[i]
	})

	// Then apply algorithm to prevent same words appearing close together
	queue = s.shufflePreventDuplicates(queue)

	return queue, nil
}

// sortCards sorts cards by priority
func (s *TrainingService) sortCards(cards []*models.UserCard, now time.Time) {
	// Sort: learning first, then by overdue time
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			// Learning cards come first
			if cards[i].State != models.StateLearning && cards[j].State == models.StateLearning {
				cards[i], cards[j] = cards[j], cards[i]
				continue
			}
			if cards[i].State == models.StateLearning && cards[j].State != models.StateLearning {
				continue
			}

			// For same state, sort by overdue time
			if cards[i].NextDueAt != nil && cards[j].NextDueAt != nil {
				if cards[i].NextDueAt.After(*cards[j].NextDueAt) {
					cards[i], cards[j] = cards[j], cards[i]
				}
			}
		}
	}
}

// shufflePreventDuplicates shuffles cards while preventing same word appearing close together
func (s *TrainingService) shufflePreventDuplicates(queue []*models.UserCardWithTraining) []*models.UserCardWithTraining {
	if len(queue) <= 1 {
		return queue
	}

	// Group by word_card_id (same word, different senses/directions)
	wordGroups := make(map[int64][]*models.UserCardWithTraining)
	for _, card := range queue {
		wordGroups[card.TrainingCard.WordCardID] = append(wordGroups[card.TrainingCard.WordCardID], card)
	}

	// If no duplicates, already shuffled, return as is
	if len(wordGroups) == len(queue) {
		return queue
	}

	// Build new queue spreading duplicates apart with larger minimum distance
	result := make([]*models.UserCardWithTraining, 0, len(queue))
	
	// Calculate minimum distance based on queue size
	// For larger queues, use larger distance to better spread words
	minDistance := 5
	if len(queue) < 10 {
		minDistance = 3
	} else if len(queue) < 20 {
		minDistance = 4
	}

	// Shuffle groups to randomize which word we try first
	groupKeys := make([]int64, 0, len(wordGroups))
	for k := range wordGroups {
		groupKeys = append(groupKeys, k)
	}
	rand.Shuffle(len(groupKeys), func(i, j int) {
		groupKeys[i], groupKeys[j] = groupKeys[j], groupKeys[i]
	})

	for len(result) < len(queue) {
		added := false
		
		// Try each word group in random order
		for _, wordCardID := range groupKeys {
			cards := wordGroups[wordCardID]
			if len(cards) == 0 {
				continue
			}

			// Check if we can add this word (not used in last N positions)
			canAdd := true
			checkDistance := minDistance
			if len(result) < checkDistance {
				checkDistance = len(result)
			}
			
			for i := len(result) - checkDistance; i < len(result); i++ {
				if result[i].TrainingCard.WordCardID == wordCardID {
					canAdd = false
					break
				}
			}

			if canAdd || len(result) == 0 {
				// Shuffle cards within this word group to randomize which card we take
				rand.Shuffle(len(cards), func(i, j int) {
					cards[i], cards[j] = cards[j], cards[i]
				})
				
				// Add first card from this group
				result = append(result, cards[0])
				wordGroups[wordCardID] = cards[1:]
				added = true
			}
		}

		// If we couldn't add any card (all words are too recent), add the next available
		// This should rarely happen with proper distance
		if !added {
			for _, wordCardID := range groupKeys {
				cards := wordGroups[wordCardID]
				if len(cards) > 0 {
					rand.Shuffle(len(cards), func(i, j int) {
						cards[i], cards[j] = cards[j], cards[i]
					})
					result = append(result, cards[0])
					wordGroups[wordCardID] = cards[1:]
					break
				}
			}
		}

		// Clean up empty groups
		for i := len(groupKeys) - 1; i >= 0; i-- {
			if len(wordGroups[groupKeys[i]]) == 0 {
				groupKeys = append(groupKeys[:i], groupKeys[i+1:]...)
			}
		}
	}

	return result
}

// GetDueCount gets the count of due cards for a user
func (s *TrainingService) GetDueCount(userID int64) (int, error) {
	return s.userCardRepo.GetDueCount(userID, time.Now())
}

// GetSession gets a session by ID
func (s *TrainingService) GetSession(sessionID int64) (*models.TrainingSession, error) {
	return s.sessionRepo.GetSession(sessionID)
}

// GetActiveSession gets the active session for a user
func (s *TrainingService) GetActiveSession(userID int64) (*models.TrainingSession, error) {
	return s.sessionRepo.GetActiveSession(userID)
}

// UpdateSessionState updates session state in database
func (s *TrainingService) UpdateSessionState(sessionID int64, sessionJSON string) error {
	session, err := s.sessionRepo.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}

	session.SessionJSON = sessionJSON
	return s.sessionRepo.UpdateSession(session)
}

// RestoreQueue restores queue from user card IDs
func (s *TrainingService) RestoreQueue(userID int64, userCardIDs []int64) ([]*models.UserCardWithTraining, error) {
	if len(userCardIDs) == 0 {
		return nil, nil
	}

	queue := make([]*models.UserCardWithTraining, 0, len(userCardIDs))

	for _, userCardID := range userCardIDs {
		// Get user card
		userCard, err := s.userCardRepo.GetUserCard(userCardID)
		if err != nil {
			s.logger.Warn("failed to get user card during restore",
				zap.Int64("user_card_id", userCardID),
				zap.Error(err),
			)
			continue
		}
		if userCard == nil {
			s.logger.Warn("user card not found during restore",
				zap.Int64("user_card_id", userCardID),
			)
			continue
		}

		// Verify user owns this card
		if userCard.UserID != userID {
			s.logger.Warn("user card belongs to different user",
				zap.Int64("user_card_id", userCardID),
				zap.Int64("expected_user_id", userID),
				zap.Int64("actual_user_id", userCard.UserID),
			)
			continue
		}

		// Get training card
		trainingCard, err := s.trainingCardRepo.GetTrainingCard(userCard.TrainingCardID)
		if err != nil {
			s.logger.Warn("failed to get training card during restore",
				zap.Int64("training_card_id", userCard.TrainingCardID),
				zap.Error(err),
			)
			continue
		}
		if trainingCard == nil {
			s.logger.Warn("training card not found during restore",
				zap.Int64("training_card_id", userCard.TrainingCardID),
			)
			continue
		}

		queue = append(queue, &models.UserCardWithTraining{
			UserCard:     *userCard,
			TrainingCard: *trainingCard,
		})
	}

	return queue, nil
}

