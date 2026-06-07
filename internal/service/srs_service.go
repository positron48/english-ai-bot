package service

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// SRSService handles SRS algorithm and card state updates
type SRSService struct {
	userCardRepo      *repository.UserCardRepository
	logger            *zap.Logger
	learning          config.LearningConfig
	learningStepsFunc func(models.CardDirection) []int  // optional; if nil, models.LearningStepsDays is used (for tests to cover single-step paths)
	marshalJSONFunc   func(interface{}) ([]byte, error) // optional; if nil, json.Marshal is used (for tests to cover error path)
}

func (s *SRSService) learningSteps(d models.CardDirection) []int {
	if s.learningStepsFunc != nil {
		return s.learningStepsFunc(d)
	}
	return models.LearningStepsDays(d)
}

// NewSRSService creates a new SRS service
func NewSRSService(userCardRepo *repository.UserCardRepository, learning config.LearningConfig, logger *zap.Logger) *SRSService {
	return &SRSService{
		userCardRepo: userCardRepo,
		learning:     learning,
		logger:       logger,
	}
}

// GradeCard grades a card based on attempt data and updates its state
func (s *SRSService) GradeCard(userCard *models.UserCard, attemptData models.AttemptData) error {
	// Calculate quality from attempt
	quality := models.CalculateQuality(attemptData)

	// Store state before update
	before := s.captureState(userCard)

	gradedAt := time.Now()
	if attemptData.GradedAt != nil && !attemptData.GradedAt.IsZero() {
		gradedAt = attemptData.GradedAt.UTC()
	}

	// Update card based on quality
	s.updateCardState(userCard, quality, gradedAt)

	// Store state after update
	after := s.captureState(userCard)

	// Log the update
	s.logger.Info("graded card",
		zap.String("learning_pair", s.learning.Pair),
		zap.Int64("user_card_id", userCard.ID),
		zap.Int("quality", int(quality)),
		zap.String("state_before", string(before.State)),
		zap.String("state_after", string(after.State)),
		zap.Int("reps_before", before.Reps),
		zap.Int("reps_after", after.Reps),
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

	// If quality < 3 (wrong answer), handle lapse
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
// Uses a gentler approach: reduces interval and EF instead of full reset
// Only resets to learning if there are multiple consecutive errors
func (s *SRSService) handleLapse(card *models.UserCard, now time.Time) {
	card.LapseCount++

	// Reduce EF
	card.EF = math.Max(models.MinEF, card.EF-0.2)

	// Check if we should reset to learning (after 3+ consecutive errors)
	// or if card is already in learning
	if card.State == models.StateLearning {
		// Already in learning - stay on current step
		steps := s.learningSteps(card.Direction)
		currentStep := card.LearningStep
		if currentStep >= len(steps) {
			currentStep = len(steps) - 1
		}
		if currentStep < 0 {
			currentStep = 0
		}
		nextDue := now.Add(time.Duration(steps[currentStep]) * 24 * time.Hour)
		card.NextDueAt = &nextDue
		return
	}

	// For review cards: use gentle approach - reduce interval instead of full reset
	if card.State == models.StateReview {
		// Reduce interval by dividing by 2 (minimum 1 day)
		// This preserves progress while still making the card appear more frequently
		newInterval := int(math.Max(1, math.Floor(float64(card.IntervalDays)/2.0)))
		card.IntervalDays = newInterval

		// Don't reset reps - keep the progress
		// Don't change state - stay in review

		nextDue := now.Add(time.Duration(newInterval) * 24 * time.Hour)
		card.NextDueAt = &nextDue

		// Only reset to learning if there are 3+ consecutive errors
		// This handles cases where the word is genuinely forgotten
		if card.LapseCount >= 3 {
			// Multiple errors - reset to learning
			card.State = models.StateLearning
			card.LearningStep = 0
			card.Reps = 0
			steps := s.learningSteps(card.Direction)
			nextDue = now.Add(time.Duration(steps[0]) * 24 * time.Hour)
			card.NextDueAt = &nextDue
		}
		return
	}

	// For new cards: start learning
	if card.State == models.StateNew {
		card.State = models.StateLearning
		card.LearningStep = 0
		steps := s.learningSteps(card.Direction)
		nextDue := now.Add(time.Duration(steps[0]) * 24 * time.Hour)
		card.NextDueAt = &nextDue
		return
	}
}

// handleNew handles a new card. Correct answer always advances (to step 1 or graduate); Hard = 1 day.
func (s *SRSService) handleNew(card *models.UserCard, quality models.Quality, now time.Time) {
	card.State = models.StateLearning
	steps := s.learningSteps(card.Direction)

	if len(steps) > 1 {
		card.LearningStep = 1 // advance from "just seen" to step 1
		if quality == models.QualityHard {
			nextDue := now.Add(24 * time.Hour)
			card.NextDueAt = &nextDue
		} else {
			nextDue := now.Add(time.Duration(steps[1]) * 24 * time.Hour)
			card.NextDueAt = &nextDue
		}
	} else {
		s.graduate(card, now)
	}
}

// handleLearning handles a learning card.
// Correct answer always advances the step; quality only affects how far (interval).
// Hard = advance but next review in 1 day; Good/Easy = advance with full step interval.
func (s *SRSService) handleLearning(card *models.UserCard, quality models.Quality, now time.Time) {
	steps := s.learningSteps(card.Direction)

	// Reset lapse count on successful answer in learning phase
	if quality != models.QualityWrong {
		card.LapseCount = 0
	}

	// Always advance on correct answer
	card.LearningStep++

	if card.LearningStep >= len(steps) {
		// Graduate to review
		if card.LearningStep >= 2 {
			card.State = models.StateReview
			card.Reps = 1
			card.IntervalDays = 3
			nextDue := now.Add(3 * 24 * time.Hour)
			card.NextDueAt = &nextDue
		} else {
			s.graduate(card, now)
		}
	} else {
		// Still in learning: quality determines interval length
		if quality == models.QualityHard {
			nextDue := now.Add(24 * time.Hour)
			card.NextDueAt = &nextDue
		} else {
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

	// Reset lapse count on successful answer (only consecutive errors count)
	// This ensures that random errors don't accumulate
	if card.LapseCount > 0 {
		card.LapseCount = 0
	}

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

	marshal := json.Marshal
	if s.marshalJSONFunc != nil {
		marshal = s.marshalJSONFunc
	}
	data, err := marshal(wrongAnswers)
	if err != nil {
		return fmt.Errorf("failed to marshal wrong answers: %w", err)
	}

	card.WrongAnswersJSON = string(data)
	return nil
}
