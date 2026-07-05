package repository

import (
	"context"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestPictureQuestRepository_ImportQuestPreservesImageURL(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewPictureQuestRepository(conn, zap.NewNop())

	in := AdminPictureQuestInput{
		CourseCode:       "es_ru",
		CEFRLevel:        "A0",
		Code:             "import_picture_test",
		Title:            "Тест импорта",
		ImageURL:         "https://example.test/picture-quest.jpg",
		ImageDescription: "A test picture with one red cup.",
		MaxTurns:         30,
		TokenBudget:      40000,
		SortOrder:        0,
		Status:           "active",
	}
	tasks := []AdminPictureTaskInput{
		{Code: "main_objects", Title: "Назови главное", CompletionCriteria: "names cup", IsRequired: true, SortOrder: 0},
	}

	id, created, err := repo.ImportQuest(ctx, in, tasks)
	if err != nil || !created {
		t.Fatalf("import create: id=%d created=%v err=%v", id, created, err)
	}

	in.ImageURL = ""
	in.Title = "Тест импорта v2"
	in.ImageDescription = "Updated description."
	id2, created2, err := repo.ImportQuest(ctx, in, tasks)
	if err != nil || created2 || id2 != id {
		t.Fatalf("import update: id=%d (want %d) created=%v err=%v", id2, id, created2, err)
	}

	rows, err := repo.ListQuestsForCourseAdmin(ctx, mustCourseID(t, repo, ctx, "es_ru"))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for i := range rows {
		if rows[i].ID != id {
			continue
		}
		if rows[i].ImageURL != "https://example.test/picture-quest.jpg" {
			t.Fatalf("image_url cleared on re-import: got %q", rows[i].ImageURL)
		}
		if rows[i].Title != "Тест импорта v2" {
			t.Fatalf("title not updated: got %q", rows[i].Title)
		}
		return
	}
	t.Fatal("imported quest not found in admin list")
}

func mustCourseID(t *testing.T, repo *PictureQuestRepository, ctx context.Context, code string) int64 {
	t.Helper()
	id, err := repo.CourseIDByCode(ctx, code)
	if err != nil {
		t.Fatalf("course id: %v", err)
	}
	return id
}
