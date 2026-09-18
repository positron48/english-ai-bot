package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"tgbot-skeleton/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestGrammarSRSAnswerReturnsPersistenceFailure(t *testing.T) {
	svc, _, uid, _ := setupGrammarSRSServiceWithRepos(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc.SetSRSRepository(repository.NewGrammarSRSRepository(db, zap.NewNop()))
	mock.ExpectBegin().WillReturnError(errors.New("database unavailable"))
	result, err := svc.SubmitGrammarSrsAnswerWithClientAttemptID(context.Background(), uid, "ch1::b1::q1", "A", "pending", nil)
	if err == nil || result != nil || !strings.Contains(err.Error(), "database unavailable") {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
