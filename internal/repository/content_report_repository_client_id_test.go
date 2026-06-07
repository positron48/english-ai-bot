package repository

import (
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestContentReportRepository_HasClientReport(t *testing.T) {
	db := testutil.SetupTestDatabase(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(db.GetConnection(), logger)
	user, err := userRepo.GetOrCreateUser(888001)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	repo := NewContentReportRepository(db.GetConnection(), logger)

	exists, err := repo.HasClientReport(user.ID, "missing-id")
	if err != nil {
		t.Fatalf("HasClientReport missing: %v", err)
	}
	if exists {
		t.Fatal("expected false for missing client report")
	}

	clientID := "client-report-test-1"
	_, err = repo.Create(CreateContentReportInput{
		UserID:         user.ID,
		SourceType:     "grammar_chapter",
		ClientReportID: clientID,
		GrammarChapterID: "en.chapter.1",
		ReportCategory: "typo",
		CommentText:    "typo",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	exists, err = repo.HasClientReport(user.ID, clientID)
	if err != nil {
		t.Fatalf("HasClientReport existing: %v", err)
	}
	if !exists {
		t.Fatal("expected true for existing client report")
	}
}
