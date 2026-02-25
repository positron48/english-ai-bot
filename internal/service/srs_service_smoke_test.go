package service

// Smoke-тест установки времени следующей тренировки.
// Запуск только этого теста:
//   go test -v -run 'Smoke' ./internal/service/

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// При правильном ответе интервал всегда увеличивается, при неправильном — сокращается.

func TestSRSService_Smoke_10CorrectInRow_IntervalAlwaysIncreases(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &SRSService{userCardRepo: nil, logger: logger}

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateReview,
		EF:           models.InitialEF,
		Reps:         2,
		IntervalDays: 6,
		Direction:    models.DirectionENtoRU,
	}
	card.NextDueAt = &now

	var prevInterval int
	var prevDue *time.Time
	for i := 0; i < 10; i++ {
		c := cloneUserCard(card)
		svc.updateCardState(c, models.QualityGood, now)
		card = c

		interval := card.IntervalDays
		if i > 0 && interval <= prevInterval {
			t.Errorf("шаг %d: после правильного ответа интервал должен расти: было %d, стало %d", i+1, prevInterval, interval)
		}
		if i > 0 && prevDue != nil && card.NextDueAt != nil && !card.NextDueAt.After(*prevDue) {
			t.Errorf("шаг %d: следующая дата тренировки должна быть позже предыдущей", i+1)
		}
		prevInterval = interval
		if card.NextDueAt != nil {
			due := *card.NextDueAt
			prevDue = &due
			now = due
		}
	}
}

func TestSRSService_Smoke_10CorrectThen10Wrong_IntervalDecreases(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &SRSService{userCardRepo: nil, logger: logger}

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateReview,
		EF:           models.InitialEF,
		Reps:         2,
		IntervalDays: 6,
		Direction:    models.DirectionENtoRU,
	}
	card.NextDueAt = &now

	// 10 правильных подряд
	for i := 0; i < 10; i++ {
		c := cloneUserCard(card)
		svc.updateCardState(c, models.QualityGood, now)
		card = c
		if card.NextDueAt != nil {
			now = *card.NextDueAt
		}
	}
	intervalAfterCorrect := card.IntervalDays

	// 10 неправильных подряд
	for i := 0; i < 10; i++ {
		c := cloneUserCard(card)
		svc.updateCardState(c, models.QualityWrong, now)
		card = c
		if card.NextDueAt != nil {
			now = *card.NextDueAt
		}
	}

	// После 10 неправильных интервал должен сократиться (либо карта ушла в learning — тогда интервал по шагам короткий)
	if card.State == models.StateReview {
		if card.IntervalDays >= intervalAfterCorrect {
			t.Errorf("после 10 неправильных интервал должен быть меньше: после правильных было %d, стало %d", intervalAfterCorrect, card.IntervalDays)
		}
	}
	// Если остались в review — интервал уже проверен выше; если перешли в learning — следующая показ через короткий шаг (сократилось)
	if card.State == models.StateLearning && intervalAfterCorrect > 0 {
		steps := models.LearningStepsDays(card.Direction)
		if len(steps) > 0 && card.LearningStep >= 0 && card.LearningStep < len(steps) {
			stepDays := steps[card.LearningStep]
			if stepDays >= intervalAfterCorrect {
				t.Errorf("после 10 неправильных карта в learning с шагом %d дней — должно быть меньше чем после правильных (%d)", stepDays, intervalAfterCorrect)
			}
		}
	}
}

func TestSRSService_Smoke_WrongAnswers_IntervalDecreases(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &SRSService{userCardRepo: nil, logger: logger}

	now := time.Now()
	card := &models.UserCard{
		State:        models.StateReview,
		EF:           2.0,
		Reps:         5,
		IntervalDays: 10,
		Direction:    models.DirectionENtoRU,
	}
	card.NextDueAt = &now

	prevInterval := card.IntervalDays
	for i := 0; i < 3; i++ {
		c := cloneUserCard(card)
		svc.updateCardState(c, models.QualityWrong, now)
		card = c
		if card.IntervalDays >= prevInterval && card.State == models.StateReview {
			t.Errorf("шаг %d: после неправильного ответа интервал должен уменьшаться: было %d, стало %d", i+1, prevInterval, card.IntervalDays)
		}
		prevInterval = card.IntervalDays
		if card.NextDueAt != nil {
			now = *card.NextDueAt
		}
	}
}

func cloneUserCard(c *models.UserCard) *models.UserCard {
	next := *c
	if c.NextDueAt != nil {
		t := *c.NextDueAt
		next.NextDueAt = &t
	}
	if c.LastReviewAt != nil {
		t := *c.LastReviewAt
		next.LastReviewAt = &t
	}
	if c.LastQuality != nil {
		q := *c.LastQuality
		next.LastQuality = &q
	}
	return &next
}
