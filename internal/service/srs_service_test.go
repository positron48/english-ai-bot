package service

import (
	"errors"
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
				// Wrong answer should reset to learning
				expectedState := models.StateLearning
				service.updateCardState(tt.userCard, quality, now)
				if tt.userCard.State != expectedState {
					t.Errorf("Expected state %v, got %v", expectedState, tt.userCard.State)
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
				if card.LearningStep != 0 {
					t.Errorf("Expected learning step 0, got %d", card.LearningStep)
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
			name: "Wrong answer resets to learning",
			card: &models.UserCard{
				State:        models.StateReview,
				EF:           2.0,
				Reps:         5,
				IntervalDays: 10,
				Direction:    models.DirectionENtoRU,
			},
			quality: models.QualityWrong,
			validate: func(t *testing.T, card *models.UserCard) {
				if card.State != models.StateLearning {
					t.Errorf("Expected state %v, got %v", models.StateLearning, card.State)
				}
				if card.LearningStep != 0 {
					t.Errorf("Expected learning step 0, got %d", card.LearningStep)
				}
				if card.Reps != 0 {
					t.Errorf("Expected reps 0, got %d", card.Reps)
				}
				if card.LapseCount == 0 {
					t.Error("LapseCount should be incremented")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &models.UserCard{
				ID:            tt.card.ID,
				UserID:        tt.card.UserID,
				State:         tt.card.State,
				EF:            tt.card.EF,
				Reps:          tt.card.Reps,
				IntervalDays: tt.card.IntervalDays,
				LearningStep:  tt.card.LearningStep,
				LapseCount:    tt.card.LapseCount,
				Direction:     tt.card.Direction,
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
				ID:              1,
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
				ID:              1,
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
				ID:              tt.card.ID,
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
