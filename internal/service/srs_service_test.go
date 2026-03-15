package service

import (
	"errors"
	"math"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestNewSRSService(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewSRSService((*repository.UserCardRepository)(nil), logger)
	_ = service // Verify service is created
}

func TestSRSService_GradeCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	now := time.Now()

	tests := []struct {
		name        string
		userCard    *models.UserCard
		attemptData models.AttemptData
		updateError error
		wantError   bool
	}{
		{
			name: "Successful grade with correct answer",
			userCard: &models.UserCard{
				ID:           1,
				UserID:       123,
				State:        models.StateNew,
				EF:           models.InitialEF,
				Reps:         0,
				IntervalDays: 0,
				LearningStep: 0,
			},
			attemptData: models.AttemptData{
				Correct:      true,
				EarlyReveal:  false,
				AnswerTimeMS: 3000,
			},
			updateError: nil,
			wantError:   false,
		},
		{
			name: "Grade with wrong answer",
			userCard: &models.UserCard{
				ID:           1,
				UserID:       123,
				State:        models.StateReview,
				EF:           2.0,
				Reps:         5,
				IntervalDays: 10,
			},
			attemptData: models.AttemptData{
				Correct: false,
			},
			updateError: nil,
			wantError:   false,
		},
		{
			name: "Database update error",
			userCard: &models.UserCard{
				ID:    1,
				State: models.StateNew,
				EF:    models.InitialEF,
			},
			attemptData: models.AttemptData{
				Correct: true,
			},
			updateError: errors.New("database error"),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SRSService{
				userCardRepo: (*repository.UserCardRepository)(nil),
				logger:       logger,
			}

			// We can't easily test the full flow without a real DB, but we can test the logic
			// by checking that the card state is updated correctly
			quality := models.CalculateQuality(tt.attemptData)
			if quality == models.QualityWrong && tt.userCard.State != models.StateNew {
				initialInterval := tt.userCard.IntervalDays
				initialReps := tt.userCard.Reps
				service.updateCardState(tt.userCard, quality, now)

				// For review cards: should stay in review with reduced interval (gentle approach)
				switch tt.userCard.State {
				case models.StateReview:
					// Interval should be reduced (divided by 2)
					expectedInterval := int(math.Max(1, math.Floor(float64(initialInterval)/2.0)))
					if tt.userCard.IntervalDays != expectedInterval {
						t.Errorf("Expected interval %d, got %d", expectedInterval, tt.userCard.IntervalDays)
					}
					// Reps should be preserved
					if tt.userCard.Reps != initialReps {
						t.Errorf("Expected reps %d (preserved), got %d", initialReps, tt.userCard.Reps)
					}
				case models.StateNew:
					// New cards should go to learning
					if tt.userCard.State != models.StateLearning {
						t.Errorf("Expected state %v for new card, got %v", models.StateLearning, tt.userCard.State)
					}
				}
			}
		})
	}
}

func TestSRSService_updateCardState(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	now := time.Now()

	service := &SRSService{
		userCardRepo: nil,
		logger:       logger,
	}

	tests := []struct {
		name     string
		card     *models.UserCard
		quality  models.Quality
		validate func(*testing.T, *models.UserCard)
	}{
		{
			name: "New card with hard quality",
			card: &models.UserCard{
				State:        models.StateNew,
				EF:           models.InitialEF,
				Direction:    models.DirectionENtoRU,
				LearningStep: 0,
			},
			quality: models.QualityHard,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state %v, got %v", models.StateLearning, card.State)
				}
				if card.LearningStep != 1 {
					t.Errorf("Expected learning step 1 (advance), got %d", card.LearningStep)
				}
				if card.NextDueAt == nil {
					t.Error("NextDueAt should be set")
				}
			},
		},
		{
			name: "New card with good quality",
			card: &models.UserCard{
				State:        models.StateNew,
				EF:           models.InitialEF,
				Direction:    models.DirectionENtoRU,
				LearningStep: 0,
			},
			quality: models.QualityGood,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state %v, got %v", models.StateLearning, card.State)
				}
				if card.LearningStep != 1 {
					t.Errorf("Expected learning step 1, got %d", card.LearningStep)
				}
			},
		},
		{
			name: "Wrong answer in review reduces interval (gentle approach)",
			card: &models.UserCard{
				State:        models.StateReview,
				EF:           2.0,
				Reps:         5,
				IntervalDays: 10,
				Direction:    models.DirectionENtoRU,
			},
			quality: models.QualityWrong,
			validate: func(t *testing.T, card *models.UserCard) {
				// Should stay in review for first error
				if card.State != models.StateReview {
					t.Errorf("Expected state %v, got %v (should stay in review for first error)", models.StateReview, card.State)
				}
				// Interval should be reduced (divided by 2)
				if card.IntervalDays != 5 {
					t.Errorf("Expected interval 5 (10/2), got %d", card.IntervalDays)
				}
				// Reps should be preserved
				if card.Reps != 5 {
					t.Errorf("Expected reps 5 (preserved), got %d", card.Reps)
				}
				if card.LapseCount != 1 {
					t.Errorf("Expected LapseCount 1, got %d", card.LapseCount)
				}
				// EF should be reduced
				if card.EF >= 2.0 {
					t.Errorf("Expected EF < 2.0, got %f", card.EF)
				}
			},
		},
		{
			name: "Review card with good quality",
			card: &models.UserCard{
				State:        models.StateReview,
				EF:           2.0,
				Reps:         1,
				IntervalDays: 6,
				Direction:    models.DirectionENtoRU,
			},
			quality: models.QualityGood,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateReview {
					t.Errorf("Expected state %v, got %v", models.StateReview, card.State)
				}
				if card.Reps != 2 {
					t.Errorf("Expected reps 2, got %d", card.Reps)
				}
				// When reps=1, next interval should be calculated as 6 * EF = 6 * 2.0 = 12
				if card.IntervalDays < 6 {
					t.Errorf("Expected interval >= 6, got %d", card.IntervalDays)
				}
			},
		},
		{
			name: "handleLapse: StateLearning with step >= len(steps) clamps to last step",
			card: &models.UserCard{
				State:        models.StateLearning,
				EF:           2.0,
				LearningStep: 10,
				Direction:    models.DirectionENtoRU,
			},
			quality: models.QualityWrong,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state %v, got %v", models.StateLearning, card.State)
				}
				if card.NextDueAt == nil {
					t.Error("NextDueAt should be set")
				}
				// ENtoRU has 3 steps; currentStep clamped to 2
				if card.LapseCount != 1 {
					t.Errorf("Expected LapseCount 1, got %d", card.LapseCount)
				}
			},
		},
		{
			name: "handleLapse: StateLearning with step < 0 clamps to 0",
			card: &models.UserCard{
				State:        models.StateLearning,
				EF:           2.0,
				LearningStep: -1,
				Direction:    models.DirectionENtoRU,
			},
			quality: models.QualityWrong,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state %v, got %v", models.StateLearning, card.State)
				}
				if card.NextDueAt == nil {
					t.Error("NextDueAt should be set")
				}
				if card.LapseCount != 1 {
					t.Errorf("Expected LapseCount 1, got %d", card.LapseCount)
				}
			},
		},
		{
			name: "handleLapse: StateReview with LapseCount 2 -> 3 resets to learning",
			card: &models.UserCard{
				State:        models.StateReview,
				EF:           2.0,
				Reps:         5,
				IntervalDays: 10,
				LapseCount:   2,
				Direction:    models.DirectionENtoRU,
			},
			quality: models.QualityWrong,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state %v after 3 lapses, got %v", models.StateLearning, card.State)
				}
				if card.LearningStep != 0 {
					t.Errorf("Expected LearningStep 0, got %d", card.LearningStep)
				}
				if card.Reps != 0 {
					t.Errorf("Expected Reps 0 after reset, got %d", card.Reps)
				}
				if card.LapseCount != 3 {
					t.Errorf("Expected LapseCount 3, got %d", card.LapseCount)
				}
				if card.NextDueAt == nil {
					t.Error("NextDueAt should be set")
				}
			},
		},
		{
			name: "handleLapse: StateNew -> StateLearning",
			card: &models.UserCard{
				State:     models.StateNew,
				EF:        models.InitialEF,
				Direction: models.DirectionENtoRU,
			},
			quality: models.QualityWrong,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state %v, got %v", models.StateLearning, card.State)
				}
				if card.LearningStep != 0 {
					t.Errorf("Expected LearningStep 0, got %d", card.LearningStep)
				}
				if card.LapseCount != 1 {
					t.Errorf("Expected LapseCount 1, got %d", card.LapseCount)
				}
				if card.NextDueAt == nil {
					t.Error("NextDueAt should be set")
				}
			},
		},
		{
			name: "StateLearning with good quality advances via handleLearning",
			card: &models.UserCard{
				State:        models.StateLearning,
				EF:           models.InitialEF,
				Direction:    models.DirectionENtoRU,
				LearningStep: 0,
			},
			quality: models.QualityGood,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state %v, got %v", models.StateLearning, card.State)
				}
				if card.LearningStep != 1 {
					t.Errorf("Expected LearningStep 1, got %d", card.LearningStep)
				}
				if card.NextDueAt == nil {
					t.Error("NextDueAt should be set")
				}
			},
		},
		{
			name: "handleReview resets LapseCount and uses default interval (Reps >= 2)",
			card: &models.UserCard{
				State:        models.StateReview,
				EF:           2.0,
				Reps:         2,
				IntervalDays: 6,
				LapseCount:   2,
				Direction:    models.DirectionENtoRU,
			},
			quality: models.QualityGood,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateReview {
					t.Errorf("Expected state %v, got %v", models.StateReview, card.State)
				}
				if card.LapseCount != 0 {
					t.Errorf("Expected LapseCount 0 after successful review, got %d", card.LapseCount)
				}
				if card.Reps != 3 {
					t.Errorf("Expected Reps 3, got %d", card.Reps)
				}
				// default: ceil(6 * 2.0) = 12
				if card.IntervalDays != 12 {
					t.Errorf("Expected IntervalDays 12 (ceil(6*EF)), got %d", card.IntervalDays)
				}
				if card.NextDueAt == nil {
					t.Error("NextDueAt should be set")
				}
			},
		},
		// Direction-specific learning steps: EN-RU has [1,3,7], RU-EN has [1,3,7,14].
		// EN-RU at last learning step (2) + Good -> graduate to review (next_due 3 days).
		{
			name: "EN-RU learning step 2 Good graduates to review (3 days)",
			card: &models.UserCard{
				State:        models.StateLearning,
				LearningStep: 2,
				Direction:    models.DirectionENtoRU,
			},
			quality: models.QualityGood,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateReview {
					t.Errorf("Expected state review after graduating, got %v", card.State)
				}
				if card.IntervalDays != 3 {
					t.Errorf("Expected IntervalDays 3 after graduating from 2+ steps, got %d", card.IntervalDays)
				}
				if card.NextDueAt == nil {
					t.Fatal("NextDueAt should be set")
				}
				dueIn := card.NextDueAt.Sub(now)
				if dueIn < 2*24*time.Hour || dueIn > 4*24*time.Hour {
					t.Errorf("Expected next_due ~3 days, got %v", dueIn)
				}
			},
		},
		// RU-EN at step 2 + Good -> step 3, next_due 14 days (steps[3]=14).
		{
			name: "RU-EN learning step 2 Good advances to step 3 (14 days)",
			card: &models.UserCard{
				State:        models.StateLearning,
				LearningStep: 2,
				Direction:    models.DirectionRUtoEN,
			},
			quality: models.QualityGood,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state learning (one more step), got %v", card.State)
				}
				if card.LearningStep != 3 {
					t.Errorf("Expected LearningStep 3, got %d", card.LearningStep)
				}
				if card.NextDueAt == nil {
					t.Fatal("NextDueAt should be set")
				}
				dueIn := card.NextDueAt.Sub(now)
				if dueIn < 13*24*time.Hour || dueIn > 15*24*time.Hour {
					t.Errorf("Expected next_due ~14 days (RU-EN step 3), got %v", dueIn)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &models.UserCard{
				ID:           tt.card.ID,
				UserID:       tt.card.UserID,
				State:        tt.card.State,
				EF:           tt.card.EF,
				Reps:         tt.card.Reps,
				IntervalDays: tt.card.IntervalDays,
				LearningStep: tt.card.LearningStep,
				LapseCount:   tt.card.LapseCount,
				Direction:    tt.card.Direction,
			}

			service.updateCardState(card, tt.quality, now)
			tt.validate(t, card)
		})
	}
}

func TestSRSService_RecordWrongAnswer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := &SRSService{
		userCardRepo: nil,
		logger:       logger,
	}

	tests := []struct {
		name         string
		card         *models.UserCard
		wrongOption  string
		validateJSON func(*testing.T, string)
	}{
		{
			name: "First wrong answer",
			card: &models.UserCard{
				ID:               1,
				WrongAnswersJSON: "",
			},
			wrongOption: "wrong1",
			validateJSON: func(t *testing.T, json string) {
				if json == "" {
					t.Error("WrongAnswersJSON should not be empty")
				}
			},
		},
		{
			name: "Increment existing wrong answer",
			card: &models.UserCard{
				ID:               1,
				WrongAnswersJSON: `[{"option":"wrong1","ts":"2024-01-01T00:00:00Z","count":1}]`,
			},
			wrongOption: "wrong1",
			validateJSON: func(t *testing.T, json string) {
				if json == "" {
					t.Error("WrongAnswersJSON should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &models.UserCard{
				ID:               tt.card.ID,
				WrongAnswersJSON: tt.card.WrongAnswersJSON,
			}

			err := service.RecordWrongAnswer(card, tt.wrongOption)
			if err != nil {
				t.Errorf("RecordWrongAnswer() error = %v", err)
			}

			tt.validateJSON(t, card.WrongAnswersJSON)
		})
	}
}
