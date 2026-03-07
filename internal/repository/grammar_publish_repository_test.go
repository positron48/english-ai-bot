package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
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

	// Cover branch where updatedByUserID.Valid is true in the loop
	db := testutil.SetupTestDB(t)
	var userID int64
	_ = db.QueryRow(`INSERT INTO users (telegram_id) VALUES (204) RETURNING id`).Scan(&userID)
	repo2 := NewGrammarPublishRepository(db, zap.NewNop())
	userIDPtr := &userID
	_ = repo2.SetPublished("section", "s-with-user", true, userIDPtr)
	items2, err := repo2.GetPublishedItemsByType("section")
	if err != nil {
		t.Fatalf("GetPublishedItemsByType() error = %v", err)
	}
	if items2["s-with-user"] == nil || items2["s-with-user"].UpdatedByUserID == nil || *items2["s-with-user"].UpdatedByUserID != userID {
		t.Errorf("expected s-with-user to have UpdatedByUserID %d, got %v", userID, items2["s-with-user"])
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
	t.Run("empty slice literal is no-op", func(t *testing.T) {
		if err := repo.BulkSetPublished("section", []string{}, false, nil); err != nil {
			t.Fatalf("BulkSetPublished([]) error = %v", err)
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

// --- Error-path tests using sqlmock ---

func TestGrammarPublishRepository_SetPublished_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO grammar_published_items").
		WithArgs("section", "sec-1", 1, sql.NullInt64{}).
		WillReturnError(fmt.Errorf("db exec failed"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	err = repo.SetPublished("section", "sec-1", true, nil)
	if err == nil {
		t.Fatal("SetPublished() expected error")
	}
	if !strings.Contains(err.Error(), "failed to set published status") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_SetName_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO grammar_published_items").
		WithArgs("section", "sec-1", nil, sql.NullInt64{}).
		WillReturnError(fmt.Errorf("db exec failed"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	err = repo.SetName("section", "sec-1", nil, nil)
	if err == nil {
		t.Fatal("SetName() expected error")
	}
	if !strings.Contains(err.Error(), "failed to set name") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_GetPublishedItem_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .+ FROM grammar_published_items WHERE item_type").
		WithArgs("section", "sec-1").
		WillReturnError(fmt.Errorf("connection lost"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	item, err := repo.GetPublishedItem("section", "sec-1")
	if err == nil {
		t.Fatal("GetPublishedItem() expected error")
	}
	if item != nil {
		t.Error("GetPublishedItem() should return nil item on error")
	}
	if !strings.Contains(err.Error(), "failed to get published item") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_IsPublished_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .+ FROM grammar_published_items WHERE item_type").
		WithArgs("section", "bad").
		WillReturnError(fmt.Errorf("db error"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	ok, err := repo.IsPublished("section", "bad")
	if err == nil {
		t.Fatal("IsPublished() expected error")
	}
	if ok {
		t.Error("IsPublished() should return false when GetPublishedItem fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_GetPublishedItemsByType_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .+ FROM grammar_published_items WHERE item_type").
		WithArgs("section").
		WillReturnError(fmt.Errorf("query failed"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	items, err := repo.GetPublishedItemsByType("section")
	if err == nil {
		t.Fatal("GetPublishedItemsByType() expected error")
	}
	if items != nil {
		t.Error("GetPublishedItemsByType() should return nil map on error")
	}
	if !strings.Contains(err.Error(), "failed to query published items") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_GetPublishedItemsByType_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	cols := []string{"id", "item_type", "item_id", "is_published", "name", "updated_at", "updated_by_user_id"}
	// Bad type in id column so Scan fails
	mock.ExpectQuery("SELECT .+ FROM grammar_published_items WHERE item_type").
		WithArgs("section").
		WillReturnRows(sqlmock.NewRows(cols).AddRow("not_an_int", "section", "s1", 1, nil, "2024-01-01 12:00:00", nil))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	items, err := repo.GetPublishedItemsByType("section")
	if err == nil {
		t.Fatal("GetPublishedItemsByType() expected error")
	}
	if items != nil {
		t.Error("GetPublishedItemsByType() should return nil map on scan error")
	}
	if !strings.Contains(err.Error(), "failed to scan published item") {
		t.Errorf("expected scan error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_GetPublishedItemsByType_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	cols := []string{"id", "item_type", "item_id", "is_published", "name", "updated_at", "updated_by_user_id"}
	rows := sqlmock.NewRows(cols).AddRow(1, "section", "s1", 1, nil, "2024-01-01 12:00:00", nil)
	rows.RowError(0, fmt.Errorf("row error"))
	mock.ExpectQuery("SELECT .+ FROM grammar_published_items WHERE item_type").
		WithArgs("section").
		WillReturnRows(rows)

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	items, err := repo.GetPublishedItemsByType("section")
	if err == nil {
		t.Fatal("GetPublishedItemsByType() expected error from rows.Err()")
	}
	// items may be populated before rows.Err() is returned
	_ = items
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_BulkSetPublished_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin failed"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	err = repo.BulkSetPublished("section", []string{"id1"}, true, nil)
	if err == nil {
		t.Fatal("BulkSetPublished() expected error")
	}
	if !strings.Contains(err.Error(), "failed to begin transaction") {
		t.Errorf("expected begin error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_BulkSetPublished_PrepareError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO grammar_published_items").
		WillReturnError(fmt.Errorf("prepare failed"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	err = repo.BulkSetPublished("section", []string{"id1"}, true, nil)
	if err == nil {
		t.Fatal("BulkSetPublished() expected error")
	}
	if !strings.Contains(err.Error(), "failed to prepare statement") {
		t.Errorf("expected prepare error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_BulkSetPublished_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO grammar_published_items").
		ExpectExec().
		WithArgs("section", "id1", 1, sql.NullInt64{}).
		WillReturnError(fmt.Errorf("exec failed"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	err = repo.BulkSetPublished("section", []string{"id1"}, true, nil)
	if err == nil {
		t.Fatal("BulkSetPublished() expected error")
	}
	if !strings.Contains(err.Error(), "failed to set published") {
		t.Errorf("expected exec error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}

func TestGrammarPublishRepository_BulkSetPublished_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO grammar_published_items").
		ExpectExec().
		WithArgs("section", "id1", 1, sql.NullInt64{}).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit failed"))

	repo := NewGrammarPublishRepository(db, zap.NewNop())
	err = repo.BulkSetPublished("section", []string{"id1"}, true, nil)
	if err == nil {
		t.Fatal("BulkSetPublished() expected error")
	}
	if !strings.Contains(err.Error(), "failed to commit transaction") {
		t.Errorf("expected commit error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock: %v", err)
	}
}
