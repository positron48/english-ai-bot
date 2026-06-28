package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupReadingProgressRepo(t *testing.T) (*ReadingTextProgressRepository, int64, *sql.DB) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	if _, err := db.Exec(`DELETE FROM reading_text_progress`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM reading_texts`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
		VALUES ('chapter-a', 'en_a1', 'Chapter A', 'A1', 'en', '{"segments":[]}'),
		       ('chapter-b', 'en_a1', 'Chapter B', 'A1', 'en', '{"segments":[]}'),
		       ('chapter-c', 'en_a1', 'Chapter C', 'A1', 'en', '{"segments":[]}')
	`); err != nil {
		t.Fatal(err)
	}

	userRepo := NewUserRepository(db, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(777001)
	if err != nil {
		t.Fatal(err)
	}
	return NewReadingTextProgressRepository(db), user.ID, db
}

func TestReadingTextProgressRepository_MarkReadAndGet(t *testing.T) {
	repo, userID, _ := setupReadingProgressRepo(t)

	before, err := repo.Get(userID, "chapter-a")
	if err != nil {
		t.Fatal(err)
	}
	if before != nil {
		t.Fatalf("expected nil progress before mark, got %#v", before)
	}

	if err := repo.MarkRead(userID, "chapter-a"); err != nil {
		t.Fatal(err)
	}

	after, err := repo.Get(userID, "chapter-a")
	if err != nil {
		t.Fatal(err)
	}
	if after == nil || after.UserID != userID || after.ChapterID != "chapter-a" {
		t.Fatalf("progress = %#v", after)
	}
	if after.ReadAt.IsZero() {
		t.Fatal("expected read_at to be set")
	}
}

func TestReadingTextProgressRepository_MarkRead_Idempotent(t *testing.T) {
	repo, userID, db := setupReadingProgressRepo(t)

	if err := repo.MarkRead(userID, "chapter-b"); err != nil {
		t.Fatal(err)
	}
	first, err := repo.Get(userID, "chapter-b")
	if err != nil || first == nil {
		t.Fatalf("first progress = %#v err=%v", first, err)
	}

	time.Sleep(10 * time.Millisecond)

	if err := repo.MarkRead(userID, "chapter-b"); err != nil {
		t.Fatal(err)
	}
	second, err := repo.Get(userID, "chapter-b")
	if err != nil || second == nil {
		t.Fatalf("second progress = %#v err=%v", second, err)
	}
	if !second.ReadAt.After(first.ReadAt) && !second.ReadAt.Equal(first.ReadAt) {
		t.Fatalf("read_at should be refreshed: first=%v second=%v", first.ReadAt, second.ReadAt)
	}

	var rowCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_text_progress WHERE user_id = ? AND chapter_id = 'chapter-b'`, userID).Scan(&rowCount); err != nil {
		t.Fatal(err)
	}
	if rowCount != 1 {
		t.Fatalf("progress rows = %d, want 1", rowCount)
	}
}

func TestReadingTextProgressRepository_CountReadInSet(t *testing.T) {
	repo, userID, _ := setupReadingProgressRepo(t)

	count, err := repo.CountReadInSet(userID, nil)
	if err != nil || count != 0 {
		t.Fatalf("empty set count = %d err=%v", count, err)
	}

	if err := repo.MarkRead(userID, "chapter-a"); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkRead(userID, "chapter-c"); err != nil {
		t.Fatal(err)
	}

	count, err = repo.CountReadInSet(userID, []string{"chapter-a", "chapter-b", "chapter-c"})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}
