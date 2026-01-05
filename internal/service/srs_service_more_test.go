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
	card := &models.UserCard{
		State:        models.StateReview,
		EF:           initialEF,
		Reps:         5,
		IntervalDays: 10,
		LapseCount:   1,
		Direction:    models.DirectionENtoRU,
	}

	service.handleLapse(card, now)

	if card.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, card.State)
	}
	if card.LearningStep != 0 {
		t.Errorf("Expected LearningStep 0, got %d", card.LearningStep)
	}
	if card.Reps != 0 {
		t.Errorf("Expected Reps 0, got %d", card.Reps)
	}
	if card.LapseCount != 2 {
		t.Errorf("Expected LapseCount 2, got %d", card.LapseCount)
	}
	if card.EF >= initialEF {
		t.Errorf("Expected EF < %f, got %f", initialEF, card.EF)
	}
}
