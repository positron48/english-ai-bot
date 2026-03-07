package repository

import (
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupUserAccessCategoryRepo(t *testing.T) *UserAccessCategoryRepository {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	return NewUserAccessCategoryRepository(conn, logger)
}

func TestUserAccessCategoryRepository_CRUD(t *testing.T) {
	repo := setupUserAccessCategoryRepo(t)

	desc := "Admins"
	category := &models.UserAccessCategory{Name: "admin", Description: &desc}

	id, err := repo.CreateCategory(category)
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}

	loaded, err := repo.GetCategory(id)
	if err != nil {
		t.Fatalf("GetCategory error: %v", err)
	}
	if loaded == nil || loaded.Name != "admin" {
		t.Fatalf("expected category to be loaded")
	}

	loaded.Name = "admins"
	if err := repo.UpdateCategory(loaded); err != nil {
		t.Fatalf("UpdateCategory error: %v", err)
	}

	categories, err := repo.GetAllCategories()
	if err != nil {
		t.Fatalf("GetAllCategories error: %v", err)
	}
	if len(categories) == 0 {
		t.Fatalf("expected categories")
	}
}

func TestUserAccessCategoryRepository_PermissionsAndDelete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(100)
	repo := NewUserAccessCategoryRepository(conn, logger)

	category := &models.UserAccessCategory{Name: "reader"}
	id, err := repo.CreateCategory(category)
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}

	if _, err := repo.db.Exec(`INSERT INTO user_access_category_permissions (category_id, permission) VALUES ($1, $2)`, id, "stats:view"); err != nil {
		t.Fatalf("failed to insert permission: %v", err)
	}

	perms, err := repo.GetCategoryPermissions(id)
	if err != nil {
		t.Fatalf("GetCategoryPermissions error: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("expected permissions")
	}

	if _, err := repo.db.Exec(`INSERT INTO user_access_user_categories (user_id, category_id) VALUES ($1, $2)`, user.ID, id); err != nil {
		t.Fatalf("failed to insert user category: %v", err)
	}

	if err := repo.DeleteCategory(id); err == nil {
		t.Fatalf("expected delete to fail with assigned users")
	}

	if _, err := repo.db.Exec(`DELETE FROM user_access_user_categories WHERE category_id = $1`, id); err != nil {
		t.Fatalf("failed to clear user categories: %v", err)
	}

	if err := repo.DeleteCategory(id); err != nil {
		t.Fatalf("DeleteCategory error: %v", err)
	}
}

func TestUserAccessCategoryRepository_SetCategoryPermissions(t *testing.T) {
	repo := setupUserAccessCategoryRepo(t)

	cat := &models.UserAccessCategory{Name: "perm-cat"}
	id, err := repo.CreateCategory(cat)
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}

	// set permissions (replaces existing)
	if err := repo.SetCategoryPermissions(id, []string{"read", "write"}); err != nil {
		t.Fatalf("SetCategoryPermissions error: %v", err)
	}
	perms, err := repo.GetCategoryPermissions(id)
	if err != nil {
		t.Fatalf("GetCategoryPermissions error: %v", err)
	}
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(perms))
	}

	// replace with one permission
	if err := repo.SetCategoryPermissions(id, []string{"admin"}); err != nil {
		t.Fatalf("SetCategoryPermissions(replace) error: %v", err)
	}
	perms, _ = repo.GetCategoryPermissions(id)
	if len(perms) != 1 || perms[0] != "admin" {
		t.Fatalf("expected [admin], got %v", perms)
	}

	// empty list clears permissions
	if err := repo.SetCategoryPermissions(id, []string{}); err != nil {
		t.Fatalf("SetCategoryPermissions(empty) error: %v", err)
	}
	perms, _ = repo.GetCategoryPermissions(id)
	if len(perms) != 0 {
		t.Fatalf("expected no permissions, got %v", perms)
	}
}

func TestUserAccessCategoryRepository_GetUserCategories_SetUserCategories(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(101)
	repo := NewUserAccessCategoryRepository(conn, logger)

	c1 := &models.UserAccessCategory{Name: "ucat1"}
	id1, _ := repo.CreateCategory(c1)
	c2 := &models.UserAccessCategory{Name: "ucat2"}
	id2, _ := repo.CreateCategory(c2)

	// initially no categories
	cats, err := repo.GetUserCategories(user.ID)
	if err != nil {
		t.Fatalf("GetUserCategories error: %v", err)
	}
	if len(cats) != 0 {
		t.Fatalf("expected 0 categories, got %d", len(cats))
	}

	// set user categories
	if err := repo.SetUserCategories(user.ID, []int64{id1, id2}); err != nil {
		t.Fatalf("SetUserCategories error: %v", err)
	}
	cats, err = repo.GetUserCategories(user.ID)
	if err != nil {
		t.Fatalf("GetUserCategories error: %v", err)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}

	// replace with one
	if err := repo.SetUserCategories(user.ID, []int64{id2}); err != nil {
		t.Fatalf("SetUserCategories(replace) error: %v", err)
	}
	cats, _ = repo.GetUserCategories(user.ID)
	if len(cats) != 1 || cats[0] != id2 {
		t.Fatalf("expected [id2], got %v", cats)
	}

	// empty list clears
	if err := repo.SetUserCategories(user.ID, []int64{}); err != nil {
		t.Fatalf("SetUserCategories(empty) error: %v", err)
	}
	cats, _ = repo.GetUserCategories(user.ID)
	if len(cats) != 0 {
		t.Fatalf("expected 0 categories, got %d", len(cats))
	}
}

func TestUserAccessCategoryRepository_GetUserPermissions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(102)
	repo := NewUserAccessCategoryRepository(conn, logger)

	c1 := &models.UserAccessCategory{Name: "pcat1"}
	id1, _ := repo.CreateCategory(c1)
	c2 := &models.UserAccessCategory{Name: "pcat2"}
	id2, _ := repo.CreateCategory(c2)

	if err := repo.SetCategoryPermissions(id1, []string{"read", "write"}); err != nil {
		t.Fatalf("SetCategoryPermissions error: %v", err)
	}
	if err := repo.SetCategoryPermissions(id2, []string{"write", "admin"}); err != nil {
		t.Fatalf("SetCategoryPermissions error: %v", err)
	}
	if err := repo.SetUserCategories(user.ID, []int64{id1, id2}); err != nil {
		t.Fatalf("SetUserCategories error: %v", err)
	}

	perms, err := repo.GetUserPermissions(user.ID)
	if err != nil {
		t.Fatalf("GetUserPermissions error: %v", err)
	}
	// union: read, write, admin (distinct, ordered)
	if len(perms) < 3 {
		t.Fatalf("expected at least 3 distinct permissions, got %v", perms)
	}
}

func TestUserAccessCategoryRepository_GetUsersByCategory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	u1, _ := userRepo.GetOrCreateUser(103)
	u2, _ := userRepo.GetOrCreateUser(104)
	repo := NewUserAccessCategoryRepository(conn, logger)

	cat := &models.UserAccessCategory{Name: "bycat"}
	id, _ := repo.CreateCategory(cat)

	if err := repo.SetUserCategories(u1.ID, []int64{id}); err != nil {
		t.Fatalf("SetUserCategories error: %v", err)
	}
	if err := repo.SetUserCategories(u2.ID, []int64{id}); err != nil {
		t.Fatalf("SetUserCategories error: %v", err)
	}

	userIDs, err := repo.GetUsersByCategory(id)
	if err != nil {
		t.Fatalf("GetUsersByCategory error: %v", err)
	}
	if len(userIDs) != 2 {
		t.Fatalf("expected 2 users, got %d", len(userIDs))
	}
	seen := make(map[int64]bool)
	for _, uid := range userIDs {
		seen[uid] = true
	}
	if !seen[u1.ID] || !seen[u2.ID] {
		t.Fatalf("expected both user IDs in result, got %v", userIDs)
	}
}
