package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// GetOrCreateUser gets a user by telegram ID or creates a new one
func (r *UserRepository) GetOrCreateUser(telegramID int64) (*models.User, error) {
	// Try to get existing user
	user, err := r.GetUserByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	// Create new user
	query := `INSERT INTO users (telegram_id) VALUES (?)`
	userID, err := database.InsertAndReturnID(r.db, query, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("created new user",
		zap.Int64("user_id", userID),
		zap.Int64("telegram_id", telegramID),
	)

	return r.GetUserByID(userID)
}

// userSelectColumns is the standard users projection including subscription_tier.
const userSelectColumns = `id, telegram_id, COALESCE(telegram_username, ''), timezone, preferred_training_time,
			  COALESCE(settings_json, ''), COALESCE(subscription_tier, 'free'), created_at, updated_at`

func scanUser(row interface {
	Scan(dest ...interface{}) error
}) (*models.User, error) {
	var user models.User
	var createdAt, updatedAt string
	var tier string
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.Timezone,
		&user.PreferredTrainingTime,
		&user.SettingsJSON,
		&tier,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.SubscriptionTier = models.ParseUserTier(tier)
	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &user, nil
}

// GetUserByID gets a user by internal ID
func (r *UserRepository) GetUserByID(userID int64) (*models.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE id = ?`
	user, err := scanUser(r.db.QueryRow(query, userID))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByTelegramID gets a user by telegram ID
func (r *UserRepository) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE telegram_id = ?`
	user, err := scanUser(r.db.QueryRow(query, telegramID))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateUserSettings updates user settings
func (r *UserRepository) UpdateUserSettings(userID int64, settingsJSON string) error {
	query := `UPDATE users SET settings_json = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Exec(query, settingsJSON, userID)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}
	return nil
}

// UpdateUserPreferredTime updates user's preferred training time
func (r *UserRepository) UpdateUserPreferredTime(userID int64, preferredTime string) error {
	query := `UPDATE users SET preferred_training_time = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Exec(query, preferredTime, userID)
	if err != nil {
		return fmt.Errorf("failed to update preferred time: %w", err)
	}
	return nil
}

// GetUserByUsernameOrID gets a user by username (without @) or telegram_id.
// Username search is case-insensitive; input is trimmed of surrounding spaces.
func (r *UserRepository) GetUserByUsernameOrID(usernameOrID string) (*models.User, error) {
	usernameOrID = strings.TrimSpace(usernameOrID)
	// Try to parse as telegram_id first
	if telegramID, err := strconv.ParseInt(usernameOrID, 10, 64); err == nil {
		return r.GetUserByTelegramID(telegramID)
	}

	// Try to find by username (remove @ if present, trim spaces); case-insensitive
	username := strings.TrimSpace(strings.TrimPrefix(usernameOrID, "@"))
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE LOWER(telegram_username) = LOWER(?)`
	user, err := scanUser(r.db.QueryRow(query, username))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateSubscriptionTier sets the user's subscription tier.
func (r *UserRepository) UpdateSubscriptionTier(userID int64, tier models.UserTier) error {
	if !models.IsValidUserTier(string(tier)) {
		return fmt.Errorf("invalid subscription tier: %q", tier)
	}
	query := `UPDATE users SET subscription_tier = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Exec(query, string(tier), userID)
	if err != nil {
		return fmt.Errorf("failed to update subscription tier: %w", err)
	}
	return nil
}

// UpdateUsername updates user's telegram username
func (r *UserRepository) UpdateUsername(telegramID int64, username string) error {
	query := `UPDATE users SET telegram_username = ?, updated_at = CURRENT_TIMESTAMP WHERE telegram_id = ?`
	_, err := r.db.Exec(query, username, telegramID)
	if err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}
	return nil
}

// GetAllUsers returns all users (for notification scheduler)
func (r *UserRepository) GetAllUsers() ([]*models.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users ORDER BY id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
