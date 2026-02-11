package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// GrammarPublishRepository handles database operations for grammar course publishing
type GrammarPublishRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewGrammarPublishRepository creates a new grammar publish repository
func NewGrammarPublishRepository(db *sql.DB, logger *zap.Logger) *GrammarPublishRepository {
	return &GrammarPublishRepository{
		db:     db,
		logger: logger,
	}
}

// PublishedItem represents a published grammar item (section or chapter)
type PublishedItem struct {
	ID             int64
	ItemType       string // "section" or "chapter"
	ItemID         string
	IsPublished    bool
	Name           *string // Override name (nullable)
	UpdatedAt      time.Time
	UpdatedByUserID *int64
}

// SetPublished sets the published status for an item
func (r *GrammarPublishRepository) SetPublished(itemType, itemID string, isPublished bool, userID *int64) error {
	// DB column is INTEGER; pass 0/1 for PostgreSQL compatibility (driver cannot encode bool as int8)
	publishedInt := 0
	if isPublished {
		publishedInt = 1
	}
	query := `INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at, updated_by_user_id)
			  VALUES (?, ?, ?, CURRENT_TIMESTAMP, ?)
			  ON CONFLICT(item_type, item_id) DO UPDATE SET
			  	is_published = excluded.is_published,
			  	updated_at = CURRENT_TIMESTAMP,
			  	updated_by_user_id = excluded.updated_by_user_id`

	_, err := r.db.Exec(query, itemType, itemID, publishedInt, userID)
	if err != nil {
		return fmt.Errorf("failed to set published status: %w", err)
	}

	return nil
}

// SetName sets the override name for an item
func (r *GrammarPublishRepository) SetName(itemType, itemID string, name *string, userID *int64) error {
	query := `INSERT INTO grammar_published_items (item_type, item_id, name, updated_at, updated_by_user_id)
			  VALUES (?, ?, ?, CURRENT_TIMESTAMP, ?)
			  ON CONFLICT(item_type, item_id) DO UPDATE SET
			  	name = excluded.name,
			  	updated_at = CURRENT_TIMESTAMP,
			  	updated_by_user_id = excluded.updated_by_user_id`

	_, err := r.db.Exec(query, itemType, itemID, name, userID)
	if err != nil {
		return fmt.Errorf("failed to set name: %w", err)
	}

	return nil
}

// GetPublishedItem retrieves published item status
func (r *GrammarPublishRepository) GetPublishedItem(itemType, itemID string) (*PublishedItem, error) {
	query := `SELECT id, item_type, item_id, is_published, name, updated_at, updated_by_user_id
			  FROM grammar_published_items
			  WHERE item_type = ? AND item_id = ?`

	var item PublishedItem
	var updatedAt string
	var updatedByUserID sql.NullInt64

	err := r.db.QueryRow(query, itemType, itemID).Scan(
		&item.ID,
		&item.ItemType,
		&item.ItemID,
		&item.IsPublished,
		&item.Name,
		&updatedAt,
		&updatedByUserID,
	)

	if err == sql.ErrNoRows {
		// Default: not published
		return &PublishedItem{
			ItemType:    itemType,
			ItemID:       itemID,
			IsPublished: false,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get published item: %w", err)
	}

	item.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	if updatedByUserID.Valid {
		item.UpdatedByUserID = &updatedByUserID.Int64
	}

	return &item, nil
}

// IsPublished checks if an item is published (defaults to false if not in DB)
func (r *GrammarPublishRepository) IsPublished(itemType, itemID string) (bool, error) {
	item, err := r.GetPublishedItem(itemType, itemID)
	if err != nil {
		return false, err
	}
	return item.IsPublished, nil
}

// GetPublishedItemsByType retrieves all published items of a given type
func (r *GrammarPublishRepository) GetPublishedItemsByType(itemType string) (map[string]*PublishedItem, error) {
	query := `SELECT id, item_type, item_id, is_published, name, updated_at, updated_by_user_id
			  FROM grammar_published_items
			  WHERE item_type = ?`

	rows, err := r.db.Query(query, itemType)
	if err != nil {
		return nil, fmt.Errorf("failed to query published items: %w", err)
	}
	defer rows.Close()

	items := make(map[string]*PublishedItem)
	for rows.Next() {
		var item PublishedItem
		var updatedAt string
		var updatedByUserID sql.NullInt64

		err := rows.Scan(
			&item.ID,
			&item.ItemType,
			&item.ItemID,
			&item.IsPublished,
			&item.Name,
			&updatedAt,
			&updatedByUserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan published item: %w", err)
		}

		item.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		if updatedByUserID.Valid {
			item.UpdatedByUserID = &updatedByUserID.Int64
		}

		items[item.ItemID] = &item
	}

	return items, rows.Err()
}

// BulkSetPublished sets published status for multiple items
func (r *GrammarPublishRepository) BulkSetPublished(itemType string, itemIDs []string, isPublished bool, userID *int64) error {
	if len(itemIDs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at, updated_by_user_id)
			  VALUES (?, ?, ?, CURRENT_TIMESTAMP, ?)
			  ON CONFLICT(item_type, item_id) DO UPDATE SET
			  	is_published = excluded.is_published,
			  	updated_at = CURRENT_TIMESTAMP,
			  	updated_by_user_id = excluded.updated_by_user_id`

	// DB column is INTEGER; pass 0/1 for PostgreSQL compatibility
	publishedInt := 0
	if isPublished {
		publishedInt = 1
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, itemID := range itemIDs {
		if _, err := stmt.Exec(itemType, itemID, publishedInt, userID); err != nil {
			return fmt.Errorf("failed to set published for %s: %w", itemID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
