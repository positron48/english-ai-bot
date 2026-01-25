package repository

import (
	"database/sql"
	"fmt"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// UserAccessCategoryRepository handles database operations for user access categories
type UserAccessCategoryRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserAccessCategoryRepository creates a new user access category repository
func NewUserAccessCategoryRepository(db *sql.DB, logger *zap.Logger) *UserAccessCategoryRepository {
	return &UserAccessCategoryRepository{
		db:     db,
		logger: logger,
	}
}

// CreateCategory creates a new category
func (r *UserAccessCategoryRepository) CreateCategory(category *models.UserAccessCategory) (int64, error) {
	query := `INSERT INTO user_access_categories (name, description, created_at, updated_at)
			  VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	var description interface{}
	if category.Description != nil {
		description = *category.Description
	}

	result, err := r.db.Exec(query, category.Name, description)
	if err != nil {
		return 0, fmt.Errorf("failed to create category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get category ID: %w", err)
	}

	return id, nil
}

// GetCategory retrieves a category by ID
func (r *UserAccessCategoryRepository) GetCategory(id int64) (*models.UserAccessCategory, error) {
	query := `SELECT id, name, description, 
			  substr(created_at, 1, 19) as created_at, 
			  substr(updated_at, 1, 19) as updated_at
			  FROM user_access_categories WHERE id = ?`

	var category models.UserAccessCategory
	var createdAt, updatedAt string
	var descText sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&category.ID,
		&category.Name,
		&descText,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if descText.Valid {
		category.Description = &descText.String
	}

	if createdAt != "" {
		if t, err := parseTime(createdAt); err == nil {
			category.CreatedAt = t
		}
	}
	if updatedAt != "" {
		if t, err := parseTime(updatedAt); err == nil {
			category.UpdatedAt = t
		}
	}

	return &category, nil
}

// GetAllCategories retrieves all categories
func (r *UserAccessCategoryRepository) GetAllCategories() ([]*models.UserAccessCategory, error) {
	query := `SELECT id, name, description, 
			  substr(created_at, 1, 19) as created_at, 
			  substr(updated_at, 1, 19) as updated_at
			  FROM user_access_categories ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.UserAccessCategory
	for rows.Next() {
		var category models.UserAccessCategory
		var createdAt, updatedAt string
		var descText sql.NullString

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&descText,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			r.logger.Warn("failed to scan category", zap.Error(err))
			continue
		}

		if descText.Valid {
			category.Description = &descText.String
		}

		if createdAt != "" {
			if t, err := parseTime(createdAt); err == nil {
				category.CreatedAt = t
			}
		}
		if updatedAt != "" {
			if t, err := parseTime(updatedAt); err == nil {
				category.UpdatedAt = t
			}
		}

		categories = append(categories, &category)
	}

	return categories, rows.Err()
}

// UpdateCategory updates a category
func (r *UserAccessCategoryRepository) UpdateCategory(category *models.UserAccessCategory) error {
	query := `UPDATE user_access_categories 
			  SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	var description interface{}
	if category.Description != nil {
		description = *category.Description
	}

	_, err := r.db.Exec(query, category.Name, description, category.ID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// DeleteCategory deletes a category
func (r *UserAccessCategoryRepository) DeleteCategory(id int64) error {
	// Check if category has users
	var userCount int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM user_access_user_categories WHERE category_id = ?`, id).Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to check category users: %w", err)
	}
	if userCount > 0 {
		return fmt.Errorf("cannot delete category: it has %d users", userCount)
	}

	_, err = r.db.Exec(`DELETE FROM user_access_categories WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// GetCategoryPermissions retrieves all permissions for a category
func (r *UserAccessCategoryRepository) GetCategoryPermissions(categoryID int64) ([]string, error) {
	query := `SELECT permission FROM user_access_category_permissions WHERE category_id = ? ORDER BY permission`

	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			r.logger.Warn("failed to scan permission", zap.Error(err))
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}

// SetCategoryPermissions sets permissions for a category (replaces existing)
func (r *UserAccessCategoryRepository) SetCategoryPermissions(categoryID int64, permissions []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing permissions
	_, err = tx.Exec(`DELETE FROM user_access_category_permissions WHERE category_id = ?`, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// Insert new permissions
	stmt, err := tx.Prepare(`INSERT INTO user_access_category_permissions (category_id, permission) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, perm := range permissions {
		if _, err := stmt.Exec(categoryID, perm); err != nil {
			return fmt.Errorf("failed to insert permission: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUserCategories retrieves all category IDs for a user
func (r *UserAccessCategoryRepository) GetUserCategories(userID int64) ([]int64, error) {
	query := `SELECT category_id FROM user_access_user_categories WHERE user_id = ? ORDER BY category_id`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user categories: %w", err)
	}
	defer rows.Close()

	var categories []int64
	for rows.Next() {
		var catID int64
		if err := rows.Scan(&catID); err != nil {
			r.logger.Warn("failed to scan category ID", zap.Error(err))
			continue
		}
		categories = append(categories, catID)
	}

	return categories, rows.Err()
}

// SetUserCategories sets categories for a user (replaces existing)
func (r *UserAccessCategoryRepository) SetUserCategories(userID int64, categoryIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing categories
	_, err = tx.Exec(`DELETE FROM user_access_user_categories WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete existing categories: %w", err)
	}

	// Insert new categories
	stmt, err := tx.Prepare(`INSERT INTO user_access_user_categories (user_id, category_id) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, catID := range categoryIDs {
		if _, err := stmt.Exec(userID, catID); err != nil {
			return fmt.Errorf("failed to insert user category: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUserPermissions calculates union of all permissions for a user's categories
func (r *UserAccessCategoryRepository) GetUserPermissions(userID int64) ([]string, error) {
	query := `SELECT DISTINCT permission 
			  FROM user_access_category_permissions 
			  WHERE category_id IN (
				  SELECT category_id FROM user_access_user_categories WHERE user_id = ?
			  )
			  ORDER BY permission`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			r.logger.Warn("failed to scan permission", zap.Error(err))
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}

// GetUsersByCategory retrieves all user IDs for a category
func (r *UserAccessCategoryRepository) GetUsersByCategory(categoryID int64) ([]int64, error) {
	query := `SELECT user_id FROM user_access_user_categories WHERE category_id = ? ORDER BY user_id`

	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by category: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			r.logger.Warn("failed to scan user ID", zap.Error(err))
			continue
		}
		userIDs = append(userIDs, uid)
	}

	return userIDs, rows.Err()
}
