package repository

import (
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestContentReportRepository_ListActiveGrammarReports_AndResolveBulk(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(db.GetConnection(), logger)
	user, err := userRepo.GetOrCreateUser(777001)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	repo := NewContentReportRepository(db.GetConnection(), logger)

	id1, err := repo.Create(CreateContentReportInput{
		UserID:            user.ID,
		SourceType:        "grammar_training",
		GrammarChapterID:  "es.chapter.1",
		TheoryBlockID:     "b1",
		GrammarQuestionID: "es.chapter.1::b1::q1",
	})
	if err != nil {
		t.Fatalf("create #1: %v", err)
	}
	id2, err := repo.Create(CreateContentReportInput{
		UserID:            user.ID,
		SourceType:        "grammar_training",
		GrammarChapterID:  "en.chapter.2",
		TheoryBlockID:     "b2",
		GrammarQuestionID: "en.chapter.2::b2::q1",
	})
	if err != nil {
		t.Fatalf("create #2: %v", err)
	}
	_, err = repo.Create(CreateContentReportInput{
		UserID:     user.ID,
		SourceType: "word_training",
		Word:       "hello",
	})
	if err != nil {
		t.Fatalf("create word report: %v", err)
	}

	list, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListActiveGrammarReports all: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 grammar active reports, got %d", len(list))
	}
	if list[0].ID <= list[1].ID {
		t.Fatalf("expected DESC order by id")
	}

	esOnly, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{Course: "es", Limit: 10})
	if err != nil {
		t.Fatalf("ListActiveGrammarReports es: %v", err)
	}
	if len(esOnly) != 1 || esOnly[0].ID != id1 {
		t.Fatalf("expected only es report id=%d, got %+v", id1, esOnly)
	}

	if _, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{Course: "xx"}); err == nil {
		t.Fatal("expected unsupported course filter error")
	}

	cursorList, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{CursorID: id2, Limit: 10})
	if err != nil {
		t.Fatalf("ListActiveGrammarReports cursor: %v", err)
	}
	if len(cursorList) != 1 || cursorList[0].ID != id1 {
		t.Fatalf("expected cursor to return id1 only, got %+v", cursorList)
	}

	affected, err := repo.ResolveBulk([]int64{id1, id2}, nil)
	if err != nil {
		t.Fatalf("ResolveBulk: %v", err)
	}
	if affected != 2 {
		t.Fatalf("expected affected=2, got %d", affected)
	}

	after, err := repo.ListActiveGrammarReports(ListGrammarReportsFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListActiveGrammarReports after resolve: %v", err)
	}
	if len(after) != 0 {
		t.Fatalf("expected 0 active grammar reports after resolve, got %d", len(after))
	}
}
