package repository

import (
	"testing"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func setupUserAccessCategoryRepo(t *testing.T) (*UserAccessCategoryRepository, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	repo := NewUserAccessCategoryRepository(db.GetConnection(), logger)

	cleanup := func() {
		_ = db.Close()
	}

	return repo, cleanup
}

func TestUserAccessCategoryRepository_CRUD(t *testing.T) {
	repo, cleanup := setupUserAccessCategoryRepo(t)
	defer cleanup()

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
	repo, cleanup := setupUserAccessCategoryRepo(t)
	defer cleanup()

	category := &models.UserAccessCategory{Name: "reader"}
	id, err := repo.CreateCategory(category)
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}

	if _, err := repo.db.Exec(`INSERT INTO user_access_category_permissions (category_id, permission) VALUES (?, ?)`, id, "stats:view"); err != nil {
		t.Fatalf("failed to insert permission: %v", err)
	}

	perms, err := repo.GetCategoryPermissions(id)
	if err != nil {
		t.Fatalf("GetCategoryPermissions error: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("expected permissions")
	}

	if _, err := repo.db.Exec(`INSERT INTO user_access_user_categories (user_id, category_id) VALUES (?, ?)`, 1, id); err != nil {
		t.Fatalf("failed to insert user category: %v", err)
	}

	if err := repo.DeleteCategory(id); err == nil {
		t.Fatalf("expected delete to fail with assigned users")
	}

	if _, err := repo.db.Exec(`DELETE FROM user_access_user_categories WHERE category_id = ?`, id); err != nil {
		t.Fatalf("failed to clear user categories: %v", err)
	}

	if err := repo.DeleteCategory(id); err != nil {
		t.Fatalf("DeleteCategory error: %v", err)
	}
}
