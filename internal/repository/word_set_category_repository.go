package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// WordSetCategoryRepository handles database operations for word set categories
type WordSetCategoryRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewWordSetCategoryRepository creates a new word set category repository
func NewWordSetCategoryRepository(db *sql.DB, logger *zap.Logger) *WordSetCategoryRepository {
	return &WordSetCategoryRepository{
		db:     db,
		logger: logger,
	}
}

// CreateCategory creates a new category
func (r *WordSetCategoryRepository) CreateCategory(category *models.WordSetCategory) (int64, error) {
	query := `INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	var parentID interface{}
	if category.ParentID != nil {
		parentID = *category.ParentID
	}

	var description interface{}
	if category.Description != nil {
		description = *category.Description
	}

	isPublished := 0
	if category.IsPublished {
		isPublished = 1
	}

	var courseCode interface{}
	if category.CourseCode != "" {
		courseCode = category.CourseCode
	}
	id, err := database.InsertAndReturnID(r.db, query, courseCode, parentID, category.Name, description, isPublished, category.SortOrder)
	if err != nil {
		return 0, fmt.Errorf("failed to create category: %w", err)
	}

	return id, nil
}

// parseTime parses datetime string with multiple format support
func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}

	// Try standard format first
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err == nil {
		return t, nil
	}

	// Try RFC3339 format
	if t2, err2 := time.Parse(time.RFC3339, timeStr); err2 == nil {
		return t2, nil
	}

	// Try ISO format without timezone
	if t3, err3 := time.Parse("2006-01-02T15:04:05", timeStr); err3 == nil {
		return t3, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// GetCategory retrieves a category by ID
func (r *WordSetCategoryRepository) GetCategory(id int64) (*models.WordSetCategory, error) {
	query := `SELECT id, parent_id, name, description, is_published, sort_order,
			  substr(CAST(created_at AS TEXT), 1, 19) as created_at,
			  substr(CAST(updated_at AS TEXT), 1, 19) as updated_at
			  FROM word_set_categories WHERE id = ?`

	var category models.WordSetCategory
	var createdAt, updatedAt string
	var parentID sql.NullInt64
	var descText sql.NullString
	var isPublished int

	err := r.db.QueryRow(query, id).Scan(
		&category.ID,
		&parentID,
		&category.Name,
		&descText,
		&isPublished,
		&category.SortOrder,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	category.IsPublished = isPublished == 1

	if parentID.Valid {
		pid := int64(parentID.Int64)
		category.ParentID = &pid
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

// GetCategoryForCourse retrieves a category only when it belongs to courseCode.
func (r *WordSetCategoryRepository) GetCategoryForCourse(id int64, courseCode string) (*models.WordSetCategory, error) {
	query := `SELECT id, COALESCE(course_code, ''), parent_id, name, description, is_published, sort_order,
			  substr(CAST(created_at AS TEXT), 1, 19) as created_at,
			  substr(CAST(updated_at AS TEXT), 1, 19) as updated_at
			  FROM word_set_categories WHERE id = ? AND course_code = ?`

	var category models.WordSetCategory
	var createdAt, updatedAt string
	var parentID sql.NullInt64
	var descText sql.NullString
	var isPublished int
	err := r.db.QueryRow(query, id, courseCode).Scan(
		&category.ID,
		&category.CourseCode,
		&parentID,
		&category.Name,
		&descText,
		&isPublished,
		&category.SortOrder,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	category.IsPublished = isPublished == 1
	if parentID.Valid {
		pid := parentID.Int64
		category.ParentID = &pid
	}
	if descText.Valid {
		category.Description = &descText.String
	}
	category.CreatedAt, _ = parseTime(createdAt)
	category.UpdatedAt, _ = parseTime(updatedAt)
	return &category, nil
}

// GetAllCategories retrieves all categories
func (r *WordSetCategoryRepository) GetAllCategories() ([]*models.WordSetCategory, error) {
	query := `SELECT id, parent_id, name, description, is_published, sort_order,
			  substr(CAST(created_at AS TEXT), 1, 19) as created_at,
			  substr(CAST(updated_at AS TEXT), 1, 19) as updated_at
			  FROM word_set_categories ORDER BY sort_order, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows, false)
}

// GetAllCategoriesForCourse retrieves categories for one course.
func (r *WordSetCategoryRepository) GetAllCategoriesForCourse(courseCode string) ([]*models.WordSetCategory, error) {
	query := `SELECT id, COALESCE(course_code, ''), parent_id, name, description, is_published, sort_order,
			  substr(CAST(created_at AS TEXT), 1, 19) as created_at,
			  substr(CAST(updated_at AS TEXT), 1, 19) as updated_at
			  FROM word_set_categories`
	args := []interface{}{}
	if courseCode != "" {
		query += ` WHERE course_code = ?`
		args = append(args, courseCode)
	}
	query += ` ORDER BY sort_order, name`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows, true)
}

func (r *WordSetCategoryRepository) scanCategories(rows *sql.Rows, includeCourse bool) ([]*models.WordSetCategory, error) {
	var categories []*models.WordSetCategory
	for rows.Next() {
		var category models.WordSetCategory
		var createdAt, updatedAt string
		var parentID sql.NullInt64
		var descText sql.NullString
		var isPublished int

		var err error
		if includeCourse {
			err = rows.Scan(&category.ID, &category.CourseCode, &parentID, &category.Name, &descText, &isPublished, &category.SortOrder, &createdAt, &updatedAt)
		} else {
			err = rows.Scan(&category.ID, &parentID, &category.Name, &descText, &isPublished, &category.SortOrder, &createdAt, &updatedAt)
		}
		if err != nil {
			r.logger.Warn("failed to scan category", zap.Error(err))
			continue
		}

		category.IsPublished = isPublished == 1

		if parentID.Valid {
			pid := int64(parentID.Int64)
			category.ParentID = &pid
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

	return categories, nil
}

// GetPublishedCategories retrieves only published categories
func (r *WordSetCategoryRepository) GetPublishedCategories() ([]*models.WordSetCategory, error) {
	query := `SELECT id, parent_id, name, description, is_published, sort_order,
			  substr(CAST(created_at AS TEXT), 1, 19) as created_at,
			  substr(CAST(updated_at AS TEXT), 1, 19) as updated_at
			  FROM word_set_categories WHERE is_published = 1 ORDER BY sort_order, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get published categories: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows, false)
}

// UpdateCategory updates a category
func (r *WordSetCategoryRepository) UpdateCategory(category *models.WordSetCategory) error {
	return r.updateCategory(category, false)
}

// UpdateCategoryForCourse updates a category only within its course.
func (r *WordSetCategoryRepository) UpdateCategoryForCourse(category *models.WordSetCategory) error {
	return r.updateCategory(category, true)
}

func (r *WordSetCategoryRepository) updateCategory(category *models.WordSetCategory, scoped bool) error {
	query := `UPDATE word_set_categories 
			  SET parent_id = ?, name = ?, description = ?, is_published = ?, sort_order = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	var parentID interface{}
	if category.ParentID != nil {
		parentID = *category.ParentID
	}

	var description interface{}
	if category.Description != nil {
		description = *category.Description
	}

	isPublished := 0
	if category.IsPublished {
		isPublished = 1
	}

	args := []interface{}{parentID, category.Name, description, isPublished, category.SortOrder, category.ID}
	if scoped {
		query += ` AND course_code = ?`
		args = append(args, category.CourseCode)
	}
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// DeleteCategory deletes a category
func (r *WordSetCategoryRepository) DeleteCategory(id int64) error {
	return r.DeleteCategoryForCourse(id, "")
}

// DeleteCategoryForCourse deletes a category only when it belongs to courseCode.
func (r *WordSetCategoryRepository) DeleteCategoryForCourse(id int64, courseCode string) error {
	// Check if category has children
	var childCount int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM word_set_categories WHERE parent_id = ? AND (? = '' OR course_code = ?)`, id, courseCode, courseCode).Scan(&childCount)
	if err != nil {
		return fmt.Errorf("failed to check category children: %w", err)
	}
	if childCount > 0 {
		return fmt.Errorf("cannot delete category: it has %d child categories", childCount)
	}

	// Check if category has sets
	var setCount int
	err = r.db.QueryRow(`SELECT COUNT(*) FROM word_sets WHERE category_id = ? AND (? = '' OR course_code = ?)`, id, courseCode, courseCode).Scan(&setCount)
	if err != nil {
		return fmt.Errorf("failed to check category sets: %w", err)
	}
	if setCount > 0 {
		return fmt.Errorf("cannot delete category: it has %d word sets", setCount)
	}

	result, err := r.db.Exec(`DELETE FROM word_set_categories WHERE id = ? AND (? = '' OR course_code = ?)`, id, courseCode, courseCode)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
