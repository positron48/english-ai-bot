package service

import (
	"errors"
	"reflect"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestGradeCardFailedWriteDoesNotAdvanceMemory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := NewSRSService(repository.NewUserCardRepository(db, zap.NewNop()), config.DefaultLearningConfig(), zap.NewNop())
	card := &models.UserCard{ID: 1, State: models.StateReview, EF: 2.5, Reps: 3, IntervalDays: 10, Direction: models.DirectionRUtoEN}
	before := *card
	attempt := models.AttemptData{Correct: true, AnswerTimeMS: 1000, OptionCount: 4}
	mock.ExpectExec("UPDATE user_cards").WillReturnError(errors.New("temporary write failure"))
	if err := svc.GradeCard(card, attempt); err == nil {
		t.Fatal("expected write failure")
	}
	if !reflect.DeepEqual(*card, before) {
		t.Fatal("failed save changed the session's card")
	}
	mock.ExpectExec("UPDATE user_cards").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.GradeCard(card, attempt); err != nil {
		t.Fatal(err)
	}
	if card.Reps != before.Reps+1 {
		t.Fatalf("retry counted twice: reps=%d", card.Reps)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
