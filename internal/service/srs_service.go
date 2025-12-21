package service

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// SRSService handles SRS algorithm and card state updates
type SRSService struct {
	userCardRepo *repository.UserCardRepository
	logger       *zap.Logger
}

// NewSRSService creates a new SRS service
func NewSRSService(userCardRepo *repository.UserCardRepository, logger *zap.Logger) *SRSService {
	return &SRSService{
		userCardRepo: userCardRepo,
		logger:       logger,
	}
}

// GradeCard grades a card based on attempt data and updates its state
func (s *SRSService) GradeCard(userCard *models.UserCard, attemptData models.AttemptData) error {
	// Calculate quality from attempt
	quality := models.CalculateQuality(attemptData)
	
	// Store state before update
	before := s.captureState(userCard)
	
	// Update card based on quality
	s.updateCardState(userCard, quality, time.Now())
	
	// Store state after update
	after := s.captureState(userCard)
	
	// Log the update
	s.logger.Info("graded card",
		zap.Int64("user_card_id", userCard.ID),
		zap.Int("quality", int(quality)),
		zap.String("state_before", string(before.State)),
		zap.String("state_after", string(after.State)),
		zap.Int("interval_before", before.IntervalDays),
		zap.Int("interval_after", after.IntervalDays),
	)
	
	// Update quality
	q := int(quality)
	userCard.LastQuality = &q
	
	// Save to database
	return s.userCardRepo.UpdateUserCard(userCard)
}

// updateCardState updates card state based on quality using SM-2 algorithm
func (s *SRSService) updateCardState(card *models.UserCard, quality models.Quality, now time.Time) {
	q := quality.ToSM2Quality()
	
	// Update last review time
	card.LastReviewAt = &now
	
	// If quality < 3 (wrong answer), reset to learning
	if q < 3 {
		s.handleLapse(card, now)
		return
	}
	
	// Handle based on current state
	switch card.State {
	case models.StateNew:
		s.handleNew(card, quality, now)
	case models.StateLearning:
		s.handleLearning(card, quality, now)
	case models.StateReview:
		s.handleReview(card, quality, q, now)
	}
}

// handleLapse handles a failed review
func (s *SRSService) handleLapse(card *models.UserCard, now time.Time) {
	card.LapseCount++
	card.State = models.StateLearning
	card.LearningStep = 0
	card.Reps = 0
	card.IntervalDays = 0
	
	// Reduce EF
	card.EF = math.Max(models.MinEF, card.EF-0.2)
	
	// Set next due to first learning step
	steps := models.LearningStepsDays(card.Direction)
	nextDue := now.Add(time.Duration(steps[0]) * 24 * time.Hour)
	card.NextDueAt = &nextDue
}

// handleNew handles a new card
func (s *SRSService) handleNew(card *models.UserCard, quality models.Quality, now time.Time) {
	card.State = models.StateLearning
	card.LearningStep = 0
	
	steps := models.LearningStepsDays(card.Direction)
	
	if quality == models.QualityHard {
		// Stay on step 0
		nextDue := now.Add(time.Duration(steps[0]) * 24 * time.Hour)
		card.NextDueAt = &nextDue
	} else {
		// Move to step 1 or graduation
		if len(steps) > 1 {
			card.LearningStep = 1
			nextDue := now.Add(time.Duration(steps[1]) * 24 * time.Hour)
			card.NextDueAt = &nextDue
		} else {
			// Graduate immediately
			s.graduate(card, now)
		}
	}
}

// handleLearning handles a learning card
func (s *SRSService) handleLearning(card *models.UserCard, quality models.Quality, now time.Time) {
	steps := models.LearningStepsDays(card.Direction)
	
	if quality == models.QualityHard {
		// Repeat current step
		currentStep := card.LearningStep
		if currentStep >= len(steps) {
			currentStep = len(steps) - 1
		}
		nextDue := now.Add(time.Duration(steps[currentStep]) * 24 * time.Hour)
		card.NextDueAt = &nextDue
	} else {
		// Advance to next step
		card.LearningStep++
		
		if card.LearningStep >= len(steps) {
			// Graduate to review with improved initial interval
			// If we completed 2+ steps, use a more honest interval (3-4 days)
			// Otherwise use standard 1 day
			if card.LearningStep >= 2 {
				card.State = models.StateReview
				card.Reps = 1
				card.IntervalDays = 3 // Start with 3 days for better progression
				nextDue := now.Add(3 * 24 * time.Hour)
				card.NextDueAt = &nextDue
			} else {
				s.graduate(card, now)
			}
		} else {
			// Move to next learning step
			nextDue := now.Add(time.Duration(steps[card.LearningStep]) * 24 * time.Hour)
			card.NextDueAt = &nextDue
		}
	}
}

// handleReview handles a review card
func (s *SRSService) handleReview(card *models.UserCard, quality models.Quality, q int, now time.Time) {
	// Update EF using SM-2 formula
	// EF' = EF + (0.1 - (5-q)*(0.08 + (5-q)*0.02))
	delta := 0.1 - float64(5-q)*(0.08+float64(5-q)*0.02)
	card.EF = math.Max(models.MinEF, card.EF+delta)
	
	// Calculate new interval
	var newInterval int
	
	switch card.Reps {
	case 0:
		newInterval = 1
	case 1:
		newInterval = 6
	default:
		// interval = previous_interval * EF
		newInterval = int(math.Ceil(float64(card.IntervalDays) * card.EF))
	}
	
	card.Reps++
	card.IntervalDays = newInterval
	
	// Set next due date
	nextDue := now.Add(time.Duration(newInterval) * 24 * time.Hour)
	card.NextDueAt = &nextDue
}

// graduate moves a card from learning to review
func (s *SRSService) graduate(card *models.UserCard, now time.Time) {
	card.State = models.StateReview
	card.Reps = 0
	card.IntervalDays = 1
	
	nextDue := now.Add(24 * time.Hour)
	card.NextDueAt = &nextDue
}

// captureState captures current SRS state for logging
func (s *SRSService) captureState(card *models.UserCard) models.SRSState {
	return models.SRSState{
		State:        card.State,
		EF:           card.EF,
		Reps:         card.Reps,
		IntervalDays: card.IntervalDays,
		LearningStep: card.LearningStep,
		LapseCount:   card.LapseCount,
	}
}

// RecordWrongAnswer records a wrong answer choice for future distractor generation
func (s *SRSService) RecordWrongAnswer(card *models.UserCard, wrongOption string) error {
	type WrongAnswer struct {
		Option string    `json:"option"`
		TS     time.Time `json:"ts"`
		Count  int       `json:"count"`
	}
	
	var wrongAnswers []WrongAnswer
	if card.WrongAnswersJSON != "" {
		if err := json.Unmarshal([]byte(card.WrongAnswersJSON), &wrongAnswers); err != nil {
			s.logger.Warn("failed to parse wrong answers", zap.Error(err))
			wrongAnswers = []WrongAnswer{}
		}
	}
	
	// Check if this option already exists
	found := false
	for i := range wrongAnswers {
		if wrongAnswers[i].Option == wrongOption {
			wrongAnswers[i].Count++
			wrongAnswers[i].TS = time.Now()
			found = true
			break
		}
	}
	
	if !found {
		wrongAnswers = append(wrongAnswers, WrongAnswer{
			Option: wrongOption,
			TS:     time.Now(),
			Count:  1,
		})
	}
	
	// Keep only last 10 wrong answers
	if len(wrongAnswers) > 10 {
		wrongAnswers = wrongAnswers[len(wrongAnswers)-10:]
	}
	
	data, err := json.Marshal(wrongAnswers)
	if err != nil {
		return fmt.Errorf("failed to marshal wrong answers: %w", err)
	}
	
	card.WrongAnswersJSON = string(data)
	return nil
}

