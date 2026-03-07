package service

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestSRSService_graduate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateLearning,
		EF:           2.0,
		Reps:         0,
		IntervalDays: 0,
		LearningStep: 2,
	}

	service.graduate(card, now)

	if card.State != models.StateReview {
		t.Errorf("Expected State %v, got %v", models.StateReview, card.State)
	}
	if card.Reps != 0 {
		t.Errorf("Expected Reps 0, got %d", card.Reps)
	}
	if card.IntervalDays != 1 {
		t.Errorf("Expected IntervalDays 1, got %d", card.IntervalDays)
	}
	if card.NextDueAt == nil {
		t.Error("NextDueAt should be set")
	}
}

func TestSRSService_captureState(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	card := &models.UserCard{
		State:        models.StateReview,
		EF:           2.2,
		Reps:         5,
		IntervalDays: 10,
		LearningStep: 0,
		LapseCount:   2,
	}

	state := service.captureState(card)

	if state.State != models.StateReview {
		t.Errorf("Expected State %v, got %v", models.StateReview, state.State)
	}
	if state.EF != 2.2 {
		t.Errorf("Expected EF 2.2, got %f", state.EF)
	}
	if state.Reps != 5 {
		t.Errorf("Expected Reps 5, got %d", state.Reps)
	}
	if state.IntervalDays != 10 {
		t.Errorf("Expected IntervalDays 10, got %d", state.IntervalDays)
	}
	if state.LapseCount != 2 {
		t.Errorf("Expected LapseCount 2, got %d", state.LapseCount)
	}
}

func TestSRSService_handleNew_EasyQuality(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateNew,
		EF:           models.InitialEF,
		Direction:    models.DirectionENtoRU,
		LearningStep: 0,
	}

	service.handleNew(card, models.QualityEasy, now)

	if card.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, card.State)
	}
	// With Easy quality, should move to next step or graduate
	if card.LearningStep == 0 {
		t.Error("LearningStep should advance with Easy quality")
	}
}

func TestSRSService_handleLearning_Graduate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateLearning,
		EF:           models.InitialEF,
		Direction:    models.DirectionENtoRU,
		LearningStep: 2, // Last step for EN->RU
	}

	service.handleLearning(card, models.QualityGood, now)

	// Should graduate to review
	if card.State != models.StateReview {
		t.Errorf("Expected State %v, got %v", models.StateReview, card.State)
	}
}

func TestSRSService_handleLearning_QualityHard_RepeatsStep(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateLearning,
		EF:           models.InitialEF,
		Direction:    models.DirectionENtoRU,
		LearningStep: 1,
	}

	service.handleLearning(card, models.QualityHard, now)

	// Should stay in learning, same step, NextDueAt set
	if card.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, card.State)
	}
	if card.LearningStep != 1 {
		t.Errorf("Expected LearningStep 1 (repeat), got %d", card.LearningStep)
	}
	if card.NextDueAt == nil {
		t.Error("NextDueAt should be set")
	}
}

func TestSRSService_handleLearning_QualityGood_AdvancesStep(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateLearning,
		EF:           models.InitialEF,
		Direction:    models.DirectionENtoRU,
		LearningStep: 0,
	}

	service.handleLearning(card, models.QualityGood, now)

	if card.LearningStep != 1 {
		t.Errorf("Expected LearningStep 1, got %d", card.LearningStep)
	}
	if card.NextDueAt == nil {
		t.Error("NextDueAt should be set")
	}
}

func TestSRSService_handleReview_FirstRep(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateReview,
		EF:           2.0,
		Reps:         0,
		IntervalDays: 1,
	}

	service.handleReview(card, models.QualityGood, 4, now)

	if card.Reps != 1 {
		t.Errorf("Expected Reps 1, got %d", card.Reps)
	}
	// When Reps=0, interval should be 1 (before incrementing)
	if card.IntervalDays != 1 {
		t.Errorf("Expected IntervalDays 1, got %d", card.IntervalDays)
	}
}

func TestSRSService_handleReview_SecondRep(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateReview,
		EF:           2.0,
		Reps:         1,
		IntervalDays: 6,
	}

	service.handleReview(card, models.QualityGood, 4, now)

	if card.Reps != 2 {
		t.Errorf("Expected Reps 2, got %d", card.Reps)
	}
	// Interval should be calculated as 6 * 2.0 = 12
	if card.IntervalDays < 6 {
		t.Errorf("Expected IntervalDays >= 6, got %d", card.IntervalDays)
	}
}

func TestSRSService_handleLapse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewSRSService(nil, logger)

	now := time.Now()
	initialEF := 2.0
	
	t.Run("First error in review - gentle approach", func(t *testing.T) {
		card := &models.UserCard{
			State:        models.StateReview,
			EF:           initialEF,
			Reps:         5,
			IntervalDays: 10,
			LapseCount:   0,
			Direction:    models.DirectionENtoRU,
		}

		service.handleLapse(card, now)

		// Should stay in review
		if card.State != models.StateReview {
			t.Errorf("Expected State %v, got %v", models.StateReview, card.State)
		}
		// Interval should be reduced
		if card.IntervalDays != 5 {
			t.Errorf("Expected IntervalDays 5 (10/2), got %d", card.IntervalDays)
		}
		// Reps should be preserved
		if card.Reps != 5 {
			t.Errorf("Expected Reps 5 (preserved), got %d", card.Reps)
		}
		if card.LapseCount != 1 {
			t.Errorf("Expected LapseCount 1, got %d", card.LapseCount)
		}
		if card.EF >= initialEF {
			t.Errorf("Expected EF < %f, got %f", initialEF, card.EF)
		}
	})

	t.Run("Third consecutive error - reset to learning", func(t *testing.T) {
		card := &models.UserCard{
			State:        models.StateReview,
			EF:           initialEF,
			Reps:         5,
			IntervalDays: 10,
			LapseCount:   2, // Already 2 errors
			Direction:    models.DirectionENtoRU,
		}

		service.handleLapse(card, now)

		// Should reset to learning after 3 errors
		if card.State != models.StateLearning {
			t.Errorf("Expected State %v, got %v", models.StateLearning, card.State)
		}
		if card.LearningStep != 0 {
			t.Errorf("Expected LearningStep 0, got %d", card.LearningStep)
		}
		if card.Reps != 0 {
			t.Errorf("Expected Reps 0, got %d", card.Reps)
		}
		if card.LapseCount != 3 {
			t.Errorf("Expected LapseCount 3, got %d", card.LapseCount)
		}
	})

	t.Run("Error in learning - stay in learning", func(t *testing.T) {
		card := &models.UserCard{
			State:        models.StateLearning,
			EF:           initialEF,
			Reps:         0,
			IntervalDays: 0,
			LearningStep: 2,
			LapseCount:   0,
			Direction:    models.DirectionENtoRU,
		}

		service.handleLapse(card, now)

		// Should stay in learning
		if card.State != models.StateLearning {
			t.Errorf("Expected State %v, got %v", models.StateLearning, card.State)
		}
		// Should stay on current step (or step 0 if invalid)
		if card.LearningStep < 0 || card.LearningStep > 2 {
			t.Errorf("Expected LearningStep 0-2, got %d", card.LearningStep)
		}
		if card.LapseCount != 1 {
			t.Errorf("Expected LapseCount 1, got %d", card.LapseCount)
		}
	})
}
