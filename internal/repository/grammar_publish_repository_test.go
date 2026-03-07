package repository

import (
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupGrammarPublishRepo(t *testing.T) *GrammarPublishRepository {
	t.Helper()
	db := testutil.SetupTestDB(t)
	return NewGrammarPublishRepository(db, zap.NewNop())
}

func TestNewGrammarPublishRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewGrammarPublishRepository(db, zap.NewNop())
	if repo == nil {
		t.Fatal("NewGrammarPublishRepository() should not return nil")
	}
}

func TestGrammarPublishRepository_SetPublished(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var userID int64
	_ = db.QueryRow(`INSERT INTO users (telegram_id) VALUES (200) RETURNING id`).Scan(&userID)
	repo := NewGrammarPublishRepository(db, zap.NewNop())
	userIDPtr := &userID

	if err := repo.SetPublished("section", "sec-1", true, userIDPtr); err != nil {
		t.Fatalf("SetPublished() error = %v", err)
	}
	item, err := repo.GetPublishedItem("section", "sec-1")
	if err != nil {
		t.Fatalf("GetPublishedItem() error = %v", err)
	}
	if !item.IsPublished {
		t.Error("expected IsPublished true")
	}

	_ = repo.SetPublished("section", "sec-1", false, nil)
	item, _ = repo.GetPublishedItem("section", "sec-1")
	if item.IsPublished {
		t.Error("expected IsPublished false after update")
	}
}

func TestGrammarPublishRepository_SetName(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var userID int64
	_ = db.QueryRow(`INSERT INTO users (telegram_id) VALUES (201) RETURNING id`).Scan(&userID)
	repo := NewGrammarPublishRepository(db, zap.NewNop())
	name := "Custom Section Name"
	namePtr := &name

	if err := repo.SetName("section", "sec-name", namePtr, &userID); err != nil {
		t.Fatalf("SetName() error = %v", err)
	}
	item, err := repo.GetPublishedItem("section", "sec-name")
	if err != nil {
		t.Fatalf("GetPublishedItem() error = %v", err)
	}
	if item.Name == nil || *item.Name != name {
		t.Errorf("expected name %q, got %v", name, item.Name)
	}

	// Clear name
	if err := repo.SetName("section", "sec-name", nil, &userID); err != nil {
		t.Fatalf("SetName(nil) error = %v", err)
	}
	item, _ = repo.GetPublishedItem("section", "sec-name")
	if item.Name != nil {
		t.Errorf("expected nil name, got %v", item.Name)
	}
}

func TestGrammarPublishRepository_GetPublishedItem(t *testing.T) {
	repo := setupGrammarPublishRepo(t)

	t.Run("missing item returns default not published", func(t *testing.T) {
		item, err := repo.GetPublishedItem("chapter", "missing-id")
		if err != nil {
			t.Fatalf("GetPublishedItem() error = %v", err)
		}
		if item == nil {
			t.Fatal("GetPublishedItem() should not return nil for missing (returns default)")
		}
		if item.ItemType != "chapter" || item.ItemID != "missing-id" {
			t.Errorf("expected item_type=chapter item_id=missing-id, got %q %q", item.ItemType, item.ItemID)
		}
		if item.IsPublished {
			t.Error("expected IsPublished false for missing item")
		}
	})

	t.Run("existing item returns row", func(t *testing.T) {
		_ = repo.SetPublished("chapter", "ch-1", true, nil)
		item, err := repo.GetPublishedItem("chapter", "ch-1")
		if err != nil {
			t.Fatalf("GetPublishedItem() error = %v", err)
		}
		if !item.IsPublished {
			t.Error("expected IsPublished true")
		}
		if item.ItemID != "ch-1" {
			t.Errorf("expected ItemID ch-1, got %q", item.ItemID)
		}
	})

	t.Run("existing item with UpdatedByUserID returns it", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		var userID int64
		_ = db.QueryRow(`INSERT INTO users (telegram_id) VALUES (203) RETURNING id`).Scan(&userID)
		repo2 := NewGrammarPublishRepository(db, zap.NewNop())
		userIDPtr := &userID
		if err := repo2.SetPublished("section", "sec-with-user", true, userIDPtr); err != nil {
			t.Fatalf("SetPublished: %v", err)
		}
		item, err := repo2.GetPublishedItem("section", "sec-with-user")
		if err != nil {
			t.Fatalf("GetPublishedItem() error = %v", err)
		}
		if item.UpdatedByUserID == nil || *item.UpdatedByUserID != userID {
			t.Errorf("expected UpdatedByUserID %d, got %v", userID, item.UpdatedByUserID)
		}
	})
}

func TestGrammarPublishRepository_IsPublished(t *testing.T) {
	repo := setupGrammarPublishRepo(t)

	ok, err := repo.IsPublished("section", "unknown")
	if err != nil {
		t.Fatalf("IsPublished() error = %v", err)
	}
	if ok {
		t.Error("expected false for unknown item")
	}

	_ = repo.SetPublished("section", "pub-sec", true, nil)
	ok, err = repo.IsPublished("section", "pub-sec")
	if err != nil {
		t.Fatalf("IsPublished() error = %v", err)
	}
	if !ok {
		t.Error("expected true for published item")
	}
}

func TestGrammarPublishRepository_GetPublishedItemsByType(t *testing.T) {
	repo := setupGrammarPublishRepo(t)
	_ = repo.SetPublished("section", "s1", true, nil)
	_ = repo.SetPublished("section", "s2", false, nil)
	_ = repo.SetPublished("chapter", "c1", true, nil)

	items, err := repo.GetPublishedItemsByType("section")
	if err != nil {
		t.Fatalf("GetPublishedItemsByType() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 section items, got %d", len(items))
	}
	if items["s1"] == nil || !items["s1"].IsPublished {
		t.Error("s1 should be published")
	}
	if items["s2"] == nil || items["s2"].IsPublished {
		t.Error("s2 should not be published")
	}

	chItems, _ := repo.GetPublishedItemsByType("chapter")
	if len(chItems) != 1 || chItems["c1"] == nil {
		t.Errorf("expected 1 chapter item c1, got %d items", len(chItems))
	}
}

func TestGrammarPublishRepository_BulkSetPublished(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var userID int64
	_ = db.QueryRow(`INSERT INTO users (telegram_id) VALUES (202) RETURNING id`).Scan(&userID)
	repo := NewGrammarPublishRepository(db, zap.NewNop())

	t.Run("empty slice is no-op", func(t *testing.T) {
		if err := repo.BulkSetPublished("section", nil, true, &userID); err != nil {
			t.Fatalf("BulkSetPublished(nil) error = %v", err)
		}
	})
	if err := repo.BulkSetPublished("section", []string{"bulk-1", "bulk-2", "bulk-3"}, true, &userID); err != nil {
		t.Fatalf("BulkSetPublished() error = %v", err)
	}
	for _, id := range []string{"bulk-1", "bulk-2", "bulk-3"} {
		ok, _ := repo.IsPublished("section", id)
		if !ok {
			t.Errorf("expected %q to be published", id)
		}
	}
}
