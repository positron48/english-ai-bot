package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupWordSetCategoryTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewWordSetCategoryRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetCategoryTestDB(t)
	defer db.Close()

	repo := NewWordSetCategoryRepository(db, logger)
	if repo == nil {
		t.Error("NewWordSetCategoryRepository() should not return nil")
	}
}

func TestWordSetCategoryRepository_CreateCategory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetCategoryTestDB(t)
	defer db.Close()

	repo := NewWordSetCategoryRepository(db, logger)

	t.Run("Create category without parent", func(t *testing.T) {
		category := &models.WordSetCategory{
			Name:      "Test Category",
			SortOrder: 1,
		}

		id, err := repo.CreateCategory(category)
		if err != nil {
			t.Fatalf("CreateCategory() error = %v", err)
		}
		if id == 0 {
			t.Error("CreateCategory() should return non-zero ID")
		}
	})

	t.Run("Create category with parent", func(t *testing.T) {
		parentCategory := &models.WordSetCategory{
			Name:      "Parent Category",
			SortOrder: 0,
		}
		parentID, err := repo.CreateCategory(parentCategory)
		if err != nil {
			t.Fatalf("Failed to create parent category: %v", err)
		}

		category := &models.WordSetCategory{
			ParentID:  &parentID,
			Name:      "Child Category",
			SortOrder: 1,
		}

		id, err := repo.CreateCategory(category)
		if err != nil {
			t.Fatalf("CreateCategory() error = %v", err)
		}
		if id == 0 {
			t.Error("CreateCategory() should return non-zero ID")
		}
	})

	t.Run("Create category with description", func(t *testing.T) {
		desc := "Test description"
		category := &models.WordSetCategory{
			Name:        "Category with Description",
			Description: &desc,
			SortOrder:   2,
		}

		id, err := repo.CreateCategory(category)
		if err != nil {
			t.Fatalf("CreateCategory() error = %v", err)
		}
		if id == 0 {
			t.Error("CreateCategory() should return non-zero ID")
		}
	})
}


func TestWordSetCategoryRepository_GetCategory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetCategoryTestDB(t)
	defer db.Close()

	repo := NewWordSetCategoryRepository(db, logger)

	t.Run("Get non-existent category", func(t *testing.T) {
		category, err := repo.GetCategory(99999)
		if err != nil {
			t.Fatalf("GetCategory() error = %v", err)
		}
		if category != nil {
			t.Error("GetCategory() should return nil for non-existent category")
		}
	})

	t.Run("Get existing category", func(t *testing.T) {
		createdCategory := &models.WordSetCategory{
			Name:      "Get Test Category",
			SortOrder: 1,
		}
		id, err := repo.CreateCategory(createdCategory)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		category, err := repo.GetCategory(id)
		if err != nil {
			t.Fatalf("GetCategory() error = %v", err)
		}
		if category == nil {
			t.Fatal("GetCategory() should not return nil")
		}
		if category.ID != id {
			t.Errorf("Expected ID %d, got %d", id, category.ID)
		}
		if category.Name != "Get Test Category" {
			t.Errorf("Expected name 'Get Test Category', got %q", category.Name)
		}
	})
}

func TestWordSetCategoryRepository_GetAllCategories(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetCategoryTestDB(t)
	defer db.Close()

	repo := NewWordSetCategoryRepository(db, logger)

	t.Run("Get all categories when empty", func(t *testing.T) {
		categories, err := repo.GetAllCategories()
		if err != nil {
			t.Fatalf("GetAllCategories() error = %v", err)
		}
		if len(categories) != 0 {
			t.Errorf("Expected 0 categories, got %d", len(categories))
		}
	})

	t.Run("Get all categories", func(t *testing.T) {
		// Create multiple categories
		category1 := &models.WordSetCategory{
			Name:      "Category A",
			SortOrder: 2,
		}
		_, err := repo.CreateCategory(category1)
		if err != nil {
			t.Fatalf("Failed to create category 1: %v", err)
		}

		category2 := &models.WordSetCategory{
			Name:      "Category B",
			SortOrder: 1,
		}
		_, err = repo.CreateCategory(category2)
		if err != nil {
			t.Fatalf("Failed to create category 2: %v", err)
		}

		categories, err := repo.GetAllCategories()
		if err != nil {
			t.Fatalf("GetAllCategories() error = %v", err)
		}
		if len(categories) < 2 {
			t.Errorf("Expected at least 2 categories, got %d", len(categories))
		}
	})
}

func TestWordSetCategoryRepository_UpdateCategory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetCategoryTestDB(t)
	defer db.Close()

	repo := NewWordSetCategoryRepository(db, logger)

	t.Run("Update category", func(t *testing.T) {
		category := &models.WordSetCategory{
			Name:      "Original Name",
			SortOrder: 1,
		}
		id, err := repo.CreateCategory(category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		updatedCategory := &models.WordSetCategory{
			ID:        id,
			Name:      "Updated Name",
			SortOrder: 2,
		}
		err = repo.UpdateCategory(updatedCategory)
		if err != nil {
			t.Fatalf("UpdateCategory() error = %v", err)
		}

		// Verify update
		got, err := repo.GetCategory(id)
		if err != nil {
			t.Fatalf("GetCategory() error = %v", err)
		}
		if got.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %q", got.Name)
		}
		if got.SortOrder != 2 {
			t.Errorf("Expected sort order 2, got %d", got.SortOrder)
		}
	})

	t.Run("Update category with description", func(t *testing.T) {
		category := &models.WordSetCategory{
			Name:      "Category to Update",
			SortOrder: 1,
		}
		id, err := repo.CreateCategory(category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		desc := "Updated description"
		updatedCategory := &models.WordSetCategory{
			ID:          id,
			Name:        "Category to Update",
			Description: &desc,
			SortOrder:   1,
		}
		err = repo.UpdateCategory(updatedCategory)
		if err != nil {
			t.Fatalf("UpdateCategory() error = %v", err)
		}

		// Verify update
		got, err := repo.GetCategory(id)
		if err != nil {
			t.Fatalf("GetCategory() error = %v", err)
		}
		if got.Description == nil || *got.Description != "Updated description" {
			t.Errorf("Expected description 'Updated description', got %v", got.Description)
		}
	})
}

func TestWordSetCategoryRepository_DeleteCategory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetCategoryTestDB(t)
	defer db.Close()

	repo := NewWordSetCategoryRepository(db, logger)

	t.Run("Delete category without children", func(t *testing.T) {
		category := &models.WordSetCategory{
			Name:      "Category to Delete",
			SortOrder: 1,
		}
		id, err := repo.CreateCategory(category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		err = repo.DeleteCategory(id)
		if err != nil {
			t.Fatalf("DeleteCategory() error = %v", err)
		}

		// Verify deletion
		got, err := repo.GetCategory(id)
		if err != nil {
			t.Fatalf("GetCategory() error = %v", err)
		}
		if got != nil {
			t.Error("Category should be deleted")
		}
	})

	t.Run("Delete category with children should fail", func(t *testing.T) {
		parentCategory := &models.WordSetCategory{
			Name:      "Parent Category",
			SortOrder: 0,
		}
		parentID, err := repo.CreateCategory(parentCategory)
		if err != nil {
			t.Fatalf("Failed to create parent category: %v", err)
		}

		childCategory := &models.WordSetCategory{
			ParentID:  &parentID,
			Name:      "Child Category",
			SortOrder: 1,
		}
		_, err = repo.CreateCategory(childCategory)
		if err != nil {
			t.Fatalf("Failed to create child category: %v", err)
		}

		err = repo.DeleteCategory(parentID)
		if err == nil {
			t.Error("DeleteCategory() should fail when category has children")
		}
	})
}
