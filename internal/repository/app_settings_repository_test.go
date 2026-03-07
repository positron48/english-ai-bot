package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupAppSettingsRepo(t *testing.T) *AppSettingsRepository {
	t.Helper()
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	return NewAppSettingsRepository(db, logger)
}

func TestNewAppSettingsRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	repo := NewAppSettingsRepository(db, logger)
	if repo == nil {
		t.Fatal("NewAppSettingsRepository() should not return nil")
	}
}

func TestAppSettingsRepository_GetSetting(t *testing.T) {
	repo := setupAppSettingsRepo(t)

	t.Run("missing key returns empty string", func(t *testing.T) {
		val, err := repo.GetSetting("nonexistent_key_xyz")
		if err != nil {
			t.Fatalf("GetSetting() error = %v", err)
		}
		if val != "" {
			t.Errorf("expected empty string for missing key, got %q", val)
		}
	})

	t.Run("existing key returns value", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		_, _ = db.Exec(`INSERT INTO app_settings (key, value) VALUES ('test_key', 'test_value') ON CONFLICT (key) DO UPDATE SET value = excluded.value`)
		repo2 := NewAppSettingsRepository(db, zap.NewNop())
		val, err := repo2.GetSetting("test_key")
		if err != nil {
			t.Fatalf("GetSetting() error = %v", err)
		}
		if val != "test_value" {
			t.Errorf("expected test_value, got %q", val)
		}
	})
}

func TestAppSettingsRepository_SetSetting(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var userID int64
	_ = db.QueryRow(`INSERT INTO users (telegram_id) VALUES (100) RETURNING id`).Scan(&userID)
	repo := NewAppSettingsRepository(db, zap.NewNop())

	if err := repo.SetSetting("my_setting", "my_value", userID); err != nil {
		t.Fatalf("SetSetting() error = %v", err)
	}
	val, err := repo.GetSetting("my_setting")
	if err != nil {
		t.Fatalf("GetSetting() after SetSetting error = %v", err)
	}
	if val != "my_value" {
		t.Errorf("expected my_value, got %q", val)
	}

	// Update same key
	if err := repo.SetSetting("my_setting", "updated_value", userID); err != nil {
		t.Fatalf("SetSetting() update error = %v", err)
	}
	val, _ = repo.GetSetting("my_setting")
	if val != "updated_value" {
		t.Errorf("expected updated_value, got %q", val)
	}
}

func TestAppSettingsRepository_GetBoolSetting(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, _ = db.Exec(`INSERT INTO app_settings (key, value) VALUES ('bool_true', 'true'), ('bool_false', 'false') ON CONFLICT (key) DO UPDATE SET value = excluded.value`)
	repo := NewAppSettingsRepository(db, zap.NewNop())

	t.Run("true", func(t *testing.T) {
		got, err := repo.GetBoolSetting("bool_true")
		if err != nil {
			t.Fatalf("GetBoolSetting() error = %v", err)
		}
		if !got {
			t.Error("expected true")
		}
	})
	t.Run("false", func(t *testing.T) {
		got, err := repo.GetBoolSetting("bool_false")
		if err != nil {
			t.Fatalf("GetBoolSetting() error = %v", err)
		}
		if got {
			t.Error("expected false")
		}
	})
	t.Run("missing key is false", func(t *testing.T) {
		got, err := repo.GetBoolSetting("missing_bool")
		if err != nil {
			t.Fatalf("GetBoolSetting() error = %v", err)
		}
		if got {
			t.Error("expected false for missing key")
		}
	})
}

func TestAppSettingsRepository_SetBoolSetting(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var userID int64
	_ = db.QueryRow(`INSERT INTO users (telegram_id) VALUES (101) RETURNING id`).Scan(&userID)
	repo := NewAppSettingsRepository(db, zap.NewNop())

	if err := repo.SetBoolSetting("flag_on", true, userID); err != nil {
		t.Fatalf("SetBoolSetting(true) error = %v", err)
	}
	got, _ := repo.GetBoolSetting("flag_on")
	if !got {
		t.Error("expected true after SetBoolSetting(true)")
	}

	if err := repo.SetBoolSetting("flag_off", false, userID); err != nil {
		t.Fatalf("SetBoolSetting(false) error = %v", err)
	}
	got, _ = repo.GetBoolSetting("flag_off")
	if got {
		t.Error("expected false after SetBoolSetting(false)")
	}
}

// TestAppSettingsRepository_GetSetting_Error covers the error path when QueryRow fails (invalid DSN).
func TestAppSettingsRepository_GetSetting_Error(t *testing.T) {
	testutil.SetupTestDB(t) // ensure postgres_compat driver is registered
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewAppSettingsRepository(db, zap.NewNop())
	_, err = repo.GetSetting("any")
	if err == nil {
		t.Error("GetSetting() expected error with invalid DSN")
	}
}

// TestAppSettingsRepository_SetSetting_Error covers the error path when Exec fails.
func TestAppSettingsRepository_SetSetting_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered:", err)
	}
	defer db.Close()
	repo := NewAppSettingsRepository(db, zap.NewNop())
	err = repo.SetSetting("k", "v", 1)
	if err == nil {
		t.Error("SetSetting() expected error with invalid DSN")
	}
}

// TestAppSettingsRepository_GetBoolSetting_Error covers GetBoolSetting when GetSetting returns error.
func TestAppSettingsRepository_GetBoolSetting_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered:", err)
	}
	defer db.Close()
	repo := NewAppSettingsRepository(db, zap.NewNop())
	_, err = repo.GetBoolSetting("any")
	if err == nil {
		t.Error("GetBoolSetting() expected error when GetSetting fails")
	}
}
