package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

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
