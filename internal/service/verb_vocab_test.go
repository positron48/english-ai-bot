package service

import (
	"testing"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestEnsureVerbFormUserCardsForWordNounStopsBeforeVocabularyScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewVerbTrainingService(repository.NewVerbFormsRepository(db, zap.NewNop()), config.LearningConfig{TargetLang: "es"}, config.TrainingConfig{SpanishVerbFormsEnabled: true}, zap.NewNop())
	mock.ExpectQuery(`SELECT LOWER`).WithArgs(int64(123)).WillReturnRows(sqlmock.NewRows([]string{"word"}).AddRow("casa"))
	mock.ExpectQuery(`SELECT id FROM verb_lemmas`).WithArgs("casa", "es").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	if err := s.EnsureVerbFormUserCardsForWord(1, 123, []string{"es.presente.indicativo"}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
