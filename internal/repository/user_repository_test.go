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

const invalidDSN = "postgres://x:x@invalid.invalid:1/db?connect_timeout=1"

func setupUserTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewUserRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)
	if repo == nil {
		t.Error("NewUserRepository() should not return nil")
	}
}

func TestUserRepository_GetOrCreateUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)

	t.Run("Create new user", func(t *testing.T) {
		user, err := repo.GetOrCreateUser(12345)
		if err != nil {
			t.Fatalf("GetOrCreateUser() error = %v", err)
		}
		if user == nil {
			t.Fatal("GetOrCreateUser() should not return nil")
		}
		if user.TelegramID != 12345 {
			t.Errorf("Expected TelegramID 12345, got %d", user.TelegramID)
		}
		if user.ID == 0 {
			t.Error("User ID should be set")
		}
	})

	t.Run("Get existing user", func(t *testing.T) {
		user1, err := repo.GetOrCreateUser(67890)
		if err != nil {
			t.Fatalf("GetOrCreateUser() error = %v", err)
		}

		user2, err := repo.GetOrCreateUser(67890)
		if err != nil {
			t.Fatalf("GetOrCreateUser() error = %v", err)
		}
		if user1.ID != user2.ID {
			t.Errorf("Expected same user ID, got %d and %d", user1.ID, user2.ID)
		}
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)

	// Create a user first
	user, err := repo.GetOrCreateUser(11111)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("Get existing user", func(t *testing.T) {
		found, err := repo.GetUserByID(user.ID)
		if err != nil {
			t.Fatalf("GetUserByID() error = %v", err)
		}
		if found == nil {
			t.Fatal("GetUserByID() should not return nil")
		}
		if found.ID != user.ID {
			t.Errorf("Expected ID %d, got %d", user.ID, found.ID)
		}
	})

	t.Run("Get non-existent user", func(t *testing.T) {
		found, err := repo.GetUserByID(99999)
		if err != nil {
			t.Fatalf("GetUserByID() error = %v", err)
		}
		if found != nil {
			t.Error("GetUserByID() should return nil for non-existent user")
		}
	})
}

func TestUserRepository_GetUserByTelegramID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)

	// Create a user first
	_, err := repo.GetOrCreateUser(22222)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("Get existing user", func(t *testing.T) {
		user, err := repo.GetUserByTelegramID(22222)
		if err != nil {
			t.Fatalf("GetUserByTelegramID() error = %v", err)
		}
		if user == nil {
			t.Fatal("GetUserByTelegramID() should not return nil")
		}
		if user.TelegramID != 22222 {
			t.Errorf("Expected TelegramID 22222, got %d", user.TelegramID)
		}
	})

	t.Run("Get non-existent user", func(t *testing.T) {
		user, err := repo.GetUserByTelegramID(33333)
		if err != nil {
			t.Fatalf("GetUserByTelegramID() error = %v", err)
		}
		if user != nil {
			t.Error("GetUserByTelegramID() should return nil for non-existent user")
		}
	})
}

func TestUserRepository_UpdateUserSettings(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)

	// Create a user first
	user, err := repo.GetOrCreateUser(44444)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = repo.UpdateUserSettings(user.ID, `{"theme": "dark"}`)
	if err != nil {
		t.Fatalf("UpdateUserSettings() error = %v", err)
	}

	// Verify update
	updated, err := repo.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if updated.SettingsJSON != `{"theme": "dark"}` {
		t.Errorf("Expected settings %q, got %q", `{"theme": "dark"}`, updated.SettingsJSON)
	}
}

func TestUserRepository_UpdateUserPreferredTime(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)

	// Create a user first
	user, err := repo.GetOrCreateUser(55555)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = repo.UpdateUserPreferredTime(user.ID, "09:00")
	if err != nil {
		t.Fatalf("UpdateUserPreferredTime() error = %v", err)
	}

	// Verify update
	updated, err := repo.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if updated.PreferredTrainingTime != "09:00" {
		t.Errorf("Expected preferred time '09:00', got %q", updated.PreferredTrainingTime)
	}
}

func TestUserRepository_GetUserByUsernameOrID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)

	// Create a user with username
	user, err := repo.GetOrCreateUser(66666)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	repo.UpdateUsername(66666, "testuser")

	t.Run("Get by telegram ID as string", func(t *testing.T) {
		found, err := repo.GetUserByUsernameOrID("66666")
		if err != nil {
			t.Fatalf("GetUserByUsernameOrID() error = %v", err)
		}
		if found == nil {
			t.Fatal("GetUserByUsernameOrID() should not return nil")
		}
		if found.ID != user.ID {
			t.Errorf("Expected ID %d, got %d", user.ID, found.ID)
		}
	})

	t.Run("Get by username", func(t *testing.T) {
		found, err := repo.GetUserByUsernameOrID("testuser")
		if err != nil {
			t.Fatalf("GetUserByUsernameOrID() error = %v", err)
		}
		if found == nil {
			t.Fatal("GetUserByUsernameOrID() should not return nil")
		}
		if found.TelegramUsername != "testuser" {
			t.Errorf("Expected username 'testuser', got %q", found.TelegramUsername)
		}
	})

	t.Run("Get by username with @", func(t *testing.T) {
		found, err := repo.GetUserByUsernameOrID("@testuser")
		if err != nil {
			t.Fatalf("GetUserByUsernameOrID() error = %v", err)
		}
		if found == nil {
			t.Fatal("GetUserByUsernameOrID() should not return nil")
		}
	})

	t.Run("Get non-existent numeric id returns nil", func(t *testing.T) {
		found, err := repo.GetUserByUsernameOrID("999999999")
		if err != nil {
			t.Fatalf("GetUserByUsernameOrID() error = %v", err)
		}
		if found != nil {
			t.Error("GetUserByUsernameOrID() should return nil for non-existent id")
		}
	})

	t.Run("Get non-existent username returns nil", func(t *testing.T) {
		found, err := repo.GetUserByUsernameOrID("nonexistentuser")
		if err != nil {
			t.Fatalf("GetUserByUsernameOrID() error = %v", err)
		}
		if found != nil {
			t.Error("GetUserByUsernameOrID() should return nil for non-existent username")
		}
	})

	t.Run("Get by username case-insensitive", func(t *testing.T) {
		for _, input := range []string{"TestUser", "TESTUSER", "TeStUsEr"} {
			found, err := repo.GetUserByUsernameOrID(input)
			if err != nil {
				t.Fatalf("GetUserByUsernameOrID(%q) error = %v", input, err)
			}
			if found == nil {
				t.Fatalf("GetUserByUsernameOrID(%q) should find user", input)
			}
			if found.TelegramUsername != "testuser" {
				t.Errorf("GetUserByUsernameOrID(%q): expected username 'testuser', got %q", input, found.TelegramUsername)
			}
		}
	})

	t.Run("Get by username with surrounding spaces", func(t *testing.T) {
		found, err := repo.GetUserByUsernameOrID("  testuser  ")
		if err != nil {
			t.Fatalf("GetUserByUsernameOrID() error = %v", err)
		}
		if found == nil {
			t.Fatal("GetUserByUsernameOrID() should find user when username has spaces")
		}
		if found.TelegramUsername != "testuser" {
			t.Errorf("Expected username 'testuser', got %q", found.TelegramUsername)
		}
	})
}

func TestUserRepository_UpdateUsername(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)

	// Create a user first
	_, err := repo.GetOrCreateUser(77777)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = repo.UpdateUsername(77777, "newusername")
	if err != nil {
		t.Fatalf("UpdateUsername() error = %v", err)
	}

	// Verify update
	user, err := repo.GetUserByTelegramID(77777)
	if err != nil {
		t.Fatalf("GetUserByTelegramID() error = %v", err)
	}
	if user.TelegramUsername != "newusername" {
		t.Errorf("Expected username 'newusername', got %q", user.TelegramUsername)
	}
}

func TestUserRepository_GetAllUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserTestDB(t)

	repo := NewUserRepository(db, logger)

	// Create multiple users
	repo.GetOrCreateUser(100)
	repo.GetOrCreateUser(200)
	repo.GetOrCreateUser(300)

	users, err := repo.GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers() error = %v", err)
	}
	if len(users) < 3 {
		t.Errorf("Expected at least 3 users, got %d", len(users))
	}
}

func TestUserRepository_GetOrCreateUser_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", invalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserRepository(db, zap.NewNop())
	_, err = repo.GetOrCreateUser(12345)
	if err == nil {
		t.Error("GetOrCreateUser() expected error with invalid DSN")
	}
}

// TestUserRepository_GetOrCreateUser_InsertError covers the branch when GetUserByTelegramID
// returns nil,nil but InsertAndReturnID fails.
func TestUserRepository_GetOrCreateUser_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	telegramID := int64(88888)
	// GetUserByTelegramID: no existing user
	mock.ExpectQuery("SELECT .+ FROM users WHERE telegram_id").
		WithArgs(telegramID).
		WillReturnError(sql.ErrNoRows)
	// InsertAndReturnID: INSERT fails
	mock.ExpectQuery("INSERT INTO users .+ RETURNING id").
		WithArgs(telegramID).
		WillReturnError(fmt.Errorf("duplicate key"))

	repo := NewUserRepository(db, zap.NewNop())
	_, err = repo.GetOrCreateUser(telegramID)
	if err == nil {
		t.Error("GetOrCreateUser() expected error when insert fails")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to create user") {
		t.Errorf("expected 'failed to create user' wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestUserRepository_GetUserByID_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", invalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserRepository(db, zap.NewNop())
	_, err = repo.GetUserByID(1)
	if err == nil {
		t.Error("GetUserByID() expected error with invalid DSN")
	}
}

func TestUserRepository_GetUserByTelegramID_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", invalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserRepository(db, zap.NewNop())
	_, err = repo.GetUserByTelegramID(12345)
	if err == nil {
		t.Error("GetUserByTelegramID() expected error with invalid DSN")
	}
}

func TestUserRepository_UpdateUserSettings_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", invalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserRepository(db, zap.NewNop())
	err = repo.UpdateUserSettings(1, `{}`)
	if err == nil {
		t.Error("UpdateUserSettings() expected error with invalid DSN")
	}
}

func TestUserRepository_UpdateUserPreferredTime_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", invalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserRepository(db, zap.NewNop())
	err = repo.UpdateUserPreferredTime(1, "09:00")
	if err == nil {
		t.Error("UpdateUserPreferredTime() expected error with invalid DSN")
	}
}

func TestUserRepository_GetUserByUsernameOrID_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", invalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserRepository(db, zap.NewNop())

	t.Run("by_numeric_id", func(t *testing.T) {
		_, err := repo.GetUserByUsernameOrID("12345")
		if err == nil {
			t.Error("GetUserByUsernameOrID(numeric) expected error with invalid DSN")
		}
	})
	t.Run("by_username", func(t *testing.T) {
		_, err := repo.GetUserByUsernameOrID("someuser")
		if err == nil {
			t.Error("GetUserByUsernameOrID(username) expected error with invalid DSN")
		}
	})
}

func TestUserRepository_UpdateUsername_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", invalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserRepository(db, zap.NewNop())
	err = repo.UpdateUsername(12345, "user")
	if err == nil {
		t.Error("UpdateUsername() expected error with invalid DSN")
	}
}

func TestUserRepository_GetAllUsers_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", invalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserRepository(db, zap.NewNop())
	_, err = repo.GetAllUsers()
	if err == nil {
		t.Error("GetAllUsers() expected error with invalid DSN")
	}
}

// TestUserRepository_GetAllUsers_ScanError covers the branch when rows.Scan fails.
func TestUserRepository_GetAllUsers_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Query returns rows with bad type so Scan fails (e.g. text in id column)
	cols := []string{"id", "telegram_id", "telegram_username", "timezone", "preferred_training_time", "settings_json", "created_at", "updated_at"}
	mock.ExpectQuery("SELECT .+ FROM users ORDER BY id").
		WillReturnRows(sqlmock.NewRows(cols).AddRow("not_an_int", int64(1), "", "", "", "", "2024-01-01 12:00:00", "2024-01-01 12:00:00"))

	repo := NewUserRepository(db, zap.NewNop())
	_, err = repo.GetAllUsers()
	if err == nil {
		t.Error("GetAllUsers() expected error when scan fails")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to scan user") {
		t.Errorf("expected 'failed to scan user' error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}
