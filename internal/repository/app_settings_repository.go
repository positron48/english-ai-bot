package repository

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// AppSettingsRepository handles database operations for app settings
type AppSettingsRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAppSettingsRepository creates a new app settings repository
func NewAppSettingsRepository(db *sql.DB, logger *zap.Logger) *AppSettingsRepository {
	return &AppSettingsRepository{
		db:     db,
		logger: logger,
	}
}

// GetSetting gets a setting value by key
func (r *AppSettingsRepository) GetSetting(key string) (string, error) {
	var value string
	query := `SELECT value FROM app_settings WHERE key = ?`
	err := r.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get setting %s: %w", key, err)
	}
	return value, nil
}

// SetSetting sets a setting value by key
func (r *AppSettingsRepository) SetSetting(key, value string, userID int64) error {
	query := `
		INSERT INTO app_settings (key, value, updated_by_user_id, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_by_user_id = excluded.updated_by_user_id,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(query, key, value, userID)
	if err != nil {
		return fmt.Errorf("failed to set setting %s: %w", key, err)
	}
	return nil
}

// GetBoolSetting gets a boolean setting value by key
func (r *AppSettingsRepository) GetBoolSetting(key string) (bool, error) {
	value, err := r.GetSetting(key)
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

// SetBoolSetting sets a boolean setting value by key
func (r *AppSettingsRepository) SetBoolSetting(key string, value bool, userID int64) error {
	strValue := "false"
	if value {
		strValue = "true"
	}
	return r.SetSetting(key, strValue, userID)
}
