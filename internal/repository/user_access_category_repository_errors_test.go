package repository

import (
	"database/sql"
	"fmt"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const uacInvalidDSN = "postgres://x:x@invalid.invalid:1/db?connect_timeout=1"

// newInvalidDBRepo returns a UserAccessCategoryRepository backed by an invalid DSN
// to trigger error paths in all methods.
func newInvalidDBRepo(t *testing.T) *UserAccessCategoryRepository {
	t.Helper()
	// Ensure postgres_compat driver is registered
	testutil.SetupTestDB(t)
	logger, _ := zap.NewDevelopment()
	db, err := sql.Open("postgres_compat", uacInvalidDSN)
	if err != nil {
		t.Skipf("postgres_compat driver not registered or open failed: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewUserAccessCategoryRepository(db, logger)
}

func TestUserAccessCategoryRepository_CreateCategory_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	desc := "test"
	_, err := repo.CreateCategory(&models.UserAccessCategory{Name: "x", Description: &desc})
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_CreateCategory_NilDesc_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	_, err := repo.CreateCategory(&models.UserAccessCategory{Name: "x"})
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_GetCategory_NotFound(t *testing.T) {
	repo := setupUserAccessCategoryRepo(t)
	cat, err := repo.GetCategory(999999)
	if err != nil {
		t.Fatalf("GetCategory(nonexistent): %v", err)
	}
	if cat != nil {
		t.Fatal("expected nil for nonexistent category")
	}
}

func TestUserAccessCategoryRepository_GetCategory_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	_, err := repo.GetCategory(1)
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_GetAllCategories_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	_, err := repo.GetAllCategories()
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_UpdateCategory_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	err := repo.UpdateCategory(&models.UserAccessCategory{ID: 1, Name: "x"})
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_DeleteCategory_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	err := repo.DeleteCategory(1)
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

// TestUserAccessCategoryRepository_DeleteCategory_ExecError covers the DELETE exec error path
// (when count=0 but the DELETE itself fails).
func TestUserAccessCategoryRepository_DeleteCategory_ExecError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	err := repo.DeleteCategory(999)
	if err == nil {
		t.Fatal("expected error from invalid DB on DeleteCategory")
	}
}

func TestUserAccessCategoryRepository_GetCategoryPermissions_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	_, err := repo.GetCategoryPermissions(1)
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_SetCategoryPermissions_BeginError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	err := repo.SetCategoryPermissions(1, []string{"read"})
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_GetUserCategories_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	_, err := repo.GetUserCategories(1)
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_SetUserCategories_BeginError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	err := repo.SetUserCategories(1, []int64{1, 2})
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_GetUserPermissions_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	_, err := repo.GetUserPermissions(1)
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

func TestUserAccessCategoryRepository_GetUsersByCategory_DBError(t *testing.T) {
	repo := newInvalidDBRepo(t)
	_, err := repo.GetUsersByCategory(1)
	if err == nil {
		t.Fatal("expected error from invalid DB")
	}
}

// TestUserAccessCategoryRepository_SetCategoryPermissions_DeleteError covers the delete error path
// inside SetCategoryPermissions transaction.
func TestUserAccessCategoryRepository_SetCategoryPermissions_Errors(t *testing.T) {
	repo := setupUserAccessCategoryRepo(t)
	cat := &models.UserAccessCategory{Name: "errcat1"}
	id, err := repo.CreateCategory(cat)
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	// Set initial permissions
	if err := repo.SetCategoryPermissions(id, []string{"perm1"}); err != nil {
		t.Fatalf("SetCategoryPermissions: %v", err)
	}

	// Replace with new permissions (covers the insert loop)
	if err := repo.SetCategoryPermissions(id, []string{"perm2", "perm3"}); err != nil {
		t.Fatalf("SetCategoryPermissions replace: %v", err)
	}
	perms, _ := repo.GetCategoryPermissions(id)
	if len(perms) != 2 {
		t.Errorf("expected 2 perms, got %d", len(perms))
	}
}

// TestUserAccessCategoryRepository_SetUserCategories_Errors covers insert loop in SetUserCategories.
func TestUserAccessCategoryRepository_SetUserCategories_Errors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := setupUserAccessCategoryRepo(t).db
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(9901)
	repo := NewUserAccessCategoryRepository(conn, logger)

	c1 := &models.UserAccessCategory{Name: "errcat2"}
	id1, _ := repo.CreateCategory(c1)
	c2 := &models.UserAccessCategory{Name: "errcat3"}
	id2, _ := repo.CreateCategory(c2)

	// Set initial categories
	if err := repo.SetUserCategories(user.ID, []int64{id1}); err != nil {
		t.Fatalf("SetUserCategories: %v", err)
	}

	// Replace with multiple (covers insert loop)
	if err := repo.SetUserCategories(user.ID, []int64{id1, id2}); err != nil {
		t.Fatalf("SetUserCategories replace: %v", err)
	}
	cats, _ := repo.GetUserCategories(user.ID)
	if len(cats) != 2 {
		t.Errorf("expected 2 cats, got %d", len(cats))
	}
}

// TestUserAccessCategoryRepository_GetAllCategories_WithData verifies GetAllCategories returns multiple categories.
func TestUserAccessCategoryRepository_GetAllCategories_WithData(t *testing.T) {
	repo := setupUserAccessCategoryRepo(t)
	desc := "desc"
	for i := 0; i < 3; i++ {
		_, err := repo.CreateCategory(&models.UserAccessCategory{
			Name:        fmt.Sprintf("scancat-%d", i),
			Description: &desc,
		})
		if err != nil {
			t.Fatalf("CreateCategory: %v", err)
		}
	}
	cats, err := repo.GetAllCategories()
	if err != nil {
		t.Fatalf("GetAllCategories: %v", err)
	}
	if len(cats) < 3 {
		t.Errorf("expected at least 3 categories, got %d", len(cats))
	}
}

// TestUserAccessCategoryRepository_UpdateCategory_NilDescription covers nil description path.
func TestUserAccessCategoryRepository_UpdateCategory_NilDescription(t *testing.T) {
	repo := setupUserAccessCategoryRepo(t)
	id, err := repo.CreateCategory(&models.UserAccessCategory{Name: "updcat"})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	err = repo.UpdateCategory(&models.UserAccessCategory{ID: id, Name: "updcat2", Description: nil})
	if err != nil {
		t.Fatalf("UpdateCategory(nil desc): %v", err)
	}
}

// TestUserAccessCategoryRepository_GetAllCategories_ScanError covers the scan warning path inside rows.Next().
func TestUserAccessCategoryRepository_GetAllCategories_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row that will fail to scan (wrong column count)
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "cat1")
	mock.ExpectQuery(`SELECT id, name`).WillReturnRows(rows)

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	cats, err := repo.GetAllCategories()
	// scan error is logged as warning, not returned; rows.Err() is nil
	if err != nil {
		t.Fatalf("GetAllCategories scan error should be logged, not returned: %v", err)
	}
	// The row that failed to scan is skipped
	if len(cats) != 0 {
		t.Errorf("expected 0 categories (scan failed), got %d", len(cats))
	}
}

// TestUserAccessCategoryRepository_GetCategoryPermissions_ScanError covers the scan warning path.
func TestUserAccessCategoryRepository_GetCategoryPermissions_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return rows that fail to scan (wrong type)
	rows := sqlmock.NewRows([]string{"permission", "extra"}).
		AddRow("perm1", "extra")
	mock.ExpectQuery(`SELECT permission`).WillReturnRows(rows)

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	perms, err := repo.GetCategoryPermissions(1)
	if err != nil {
		t.Fatalf("GetCategoryPermissions scan error should be logged, not returned: %v", err)
	}
	if len(perms) != 0 {
		t.Errorf("expected 0 perms (scan failed), got %d", len(perms))
	}
}

// TestUserAccessCategoryRepository_GetUserCategories_ScanError covers the scan warning path.
func TestUserAccessCategoryRepository_GetUserCategories_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"category_id", "extra"}).
		AddRow(1, "extra")
	mock.ExpectQuery(`SELECT category_id`).WillReturnRows(rows)

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	cats, err := repo.GetUserCategories(1)
	if err != nil {
		t.Fatalf("GetUserCategories scan error should be logged, not returned: %v", err)
	}
	if len(cats) != 0 {
		t.Errorf("expected 0 cats (scan failed), got %d", len(cats))
	}
}

// TestUserAccessCategoryRepository_GetUserPermissions_ScanError covers the scan warning path.
func TestUserAccessCategoryRepository_GetUserPermissions_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"permission", "extra"}).
		AddRow("perm1", "extra")
	mock.ExpectQuery(`SELECT DISTINCT permission`).WillReturnRows(rows)

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	perms, err := repo.GetUserPermissions(1)
	if err != nil {
		t.Fatalf("GetUserPermissions scan error should be logged, not returned: %v", err)
	}
	if len(perms) != 0 {
		t.Errorf("expected 0 perms (scan failed), got %d", len(perms))
	}
}

// TestUserAccessCategoryRepository_GetUsersByCategory_ScanError covers the scan warning path.
func TestUserAccessCategoryRepository_GetUsersByCategory_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "extra"}).
		AddRow(1, "extra")
	mock.ExpectQuery(`SELECT user_id`).WillReturnRows(rows)

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	userIDs, err := repo.GetUsersByCategory(1)
	if err != nil {
		t.Fatalf("GetUsersByCategory scan error should be logged, not returned: %v", err)
	}
	if len(userIDs) != 0 {
		t.Errorf("expected 0 userIDs (scan failed), got %d", len(userIDs))
	}
}

// TestUserAccessCategoryRepository_SetCategoryPermissions_DeleteError covers the delete error inside tx.
func TestUserAccessCategoryRepository_SetCategoryPermissions_DeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM user_access_category_permissions`).
		WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.SetCategoryPermissions(1, []string{"read"})
	if err == nil {
		t.Fatal("expected error when DELETE fails in SetCategoryPermissions")
	}
}

// TestUserAccessCategoryRepository_SetCategoryPermissions_PrepareError covers the prepare error inside tx.
func TestUserAccessCategoryRepository_SetCategoryPermissions_PrepareError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM user_access_category_permissions`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(`INSERT INTO user_access_category_permissions`).
		WillReturnError(fmt.Errorf("prepare error"))
	mock.ExpectRollback()

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.SetCategoryPermissions(1, []string{"read"})
	if err == nil {
		t.Fatal("expected error when PREPARE fails in SetCategoryPermissions")
	}
}

// TestUserAccessCategoryRepository_SetCategoryPermissions_InsertError covers the insert error inside tx.
func TestUserAccessCategoryRepository_SetCategoryPermissions_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM user_access_category_permissions`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(`INSERT INTO user_access_category_permissions`)
	mock.ExpectExec(`INSERT INTO user_access_category_permissions`).
		WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.SetCategoryPermissions(1, []string{"read"})
	if err == nil {
		t.Fatal("expected error when INSERT fails in SetCategoryPermissions")
	}
}

// TestUserAccessCategoryRepository_SetCategoryPermissions_CommitError covers the commit error.
func TestUserAccessCategoryRepository_SetCategoryPermissions_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM user_access_category_permissions`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(`INSERT INTO user_access_category_permissions`)
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.SetCategoryPermissions(1, []string{})
	if err == nil {
		t.Fatal("expected error when COMMIT fails in SetCategoryPermissions")
	}
}

// TestUserAccessCategoryRepository_SetUserCategories_DeleteError covers the delete error inside tx.
func TestUserAccessCategoryRepository_SetUserCategories_DeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM user_access_user_categories`).
		WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.SetUserCategories(1, []int64{1})
	if err == nil {
		t.Fatal("expected error when DELETE fails in SetUserCategories")
	}
}

// TestUserAccessCategoryRepository_SetUserCategories_PrepareError covers the prepare error inside tx.
func TestUserAccessCategoryRepository_SetUserCategories_PrepareError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM user_access_user_categories`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(`INSERT INTO user_access_user_categories`).
		WillReturnError(fmt.Errorf("prepare error"))
	mock.ExpectRollback()

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.SetUserCategories(1, []int64{1})
	if err == nil {
		t.Fatal("expected error when PREPARE fails in SetUserCategories")
	}
}

// TestUserAccessCategoryRepository_SetUserCategories_InsertError covers the insert error inside tx.
func TestUserAccessCategoryRepository_SetUserCategories_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM user_access_user_categories`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(`INSERT INTO user_access_user_categories`)
	mock.ExpectExec(`INSERT INTO user_access_user_categories`).
		WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.SetUserCategories(1, []int64{1})
	if err == nil {
		t.Fatal("expected error when INSERT fails in SetUserCategories")
	}
}

// TestUserAccessCategoryRepository_SetUserCategories_CommitError covers the commit error.
func TestUserAccessCategoryRepository_SetUserCategories_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM user_access_user_categories`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(`INSERT INTO user_access_user_categories`)
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.SetUserCategories(1, []int64{})
	if err == nil {
		t.Fatal("expected error when COMMIT fails in SetUserCategories")
	}
}

// TestUserAccessCategoryRepository_DeleteCategory_ExecDeleteError covers the exec error on DELETE.
func TestUserAccessCategoryRepository_DeleteCategory_ExecDeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// COUNT returns 0 (no users), but DELETE fails
	mock.ExpectQuery(`SELECT COUNT\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`DELETE FROM user_access_categories`).
		WillReturnError(fmt.Errorf("delete exec error"))

	repo := NewUserAccessCategoryRepository(db, zap.NewNop())
	err = repo.DeleteCategory(1)
	if err == nil {
		t.Fatal("expected error when DELETE exec fails in DeleteCategory")
	}
}
