package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupWebSessionTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewWebSessionRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)

	repo := NewWebSessionRepository(db, logger)
	_ = repo // Verify repository is created
}

func TestWebSessionRepository_CreateSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(123)

	repo := NewWebSessionRepository(db, logger)

	session := &WebSession{
		UserID:    user.ID,
		Token:     "test-session-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if session.ID == 0 {
		t.Error("Session ID should be set after creation")
	}
}

func TestWebSessionRepository_GetSessionByToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(456)

	repo := NewWebSessionRepository(db, logger)

	// Create a session
	session := &WebSession{
		UserID:    user.ID,
		Token:     "get-session-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get the session
	found, err := repo.GetSessionByToken("get-session-token")
	if err != nil {
		t.Fatalf("GetSessionByToken() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetSessionByToken() should not return nil")
	}
	if found.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, found.UserID)
	}
	if found.Token != "get-session-token" {
		t.Errorf("Expected token 'get-session-token', got %q", found.Token)
	}
	if found.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if found.LastSeenAt.IsZero() {
		t.Error("LastSeenAt should be set")
	}
	if found.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}
}

func TestWebSessionRepository_GetSessionByToken_UnknownToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)
	repo := NewWebSessionRepository(db, logger)

	found, err := repo.GetSessionByToken("nonexistent-token-xyz")
	if err != nil {
		t.Fatalf("GetSessionByToken() error = %v", err)
	}
	if found != nil {
		t.Errorf("expected nil for unknown token, got %+v", found)
	}
}

func TestWebSessionRepository_GetSessionByToken_LongToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(888)

	repo := NewWebSessionRepository(db, logger)
	longToken := "abcdefghijklmnopqrstuvwxyz012345"
	session := &WebSession{
		UserID:    user.ID,
		Token:     longToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	found, err := repo.GetSessionByToken(longToken)
	if err != nil {
		t.Fatalf("GetSessionByToken() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetSessionByToken() should find session with long token")
	}
	if found.Token != longToken {
		t.Errorf("expected token %q, got %q", longToken, found.Token)
	}
}

func TestWebSessionRepository_DeleteSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(789)

	repo := NewWebSessionRepository(db, logger)

	// Create a session
	session := &WebSession{
		UserID:    user.ID,
		Token:     "delete-session-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Delete the session
	err = repo.DeleteSession("delete-session-token")
	if err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}

	// Verify it's deleted
	found, err := repo.GetSessionByToken("delete-session-token")
	if err != nil {
		t.Fatalf("GetSessionByToken() error = %v", err)
	}
	if found != nil {
		t.Error("GetSessionByToken() should return nil for deleted session")
	}
}

func TestWebSessionRepository_CleanupExpiredSessions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(999)

	repo := NewWebSessionRepository(db, logger)

	// Create an expired session
	session := &WebSession{
		UserID:    user.ID,
		Token:     "expired-session-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}

	err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Cleanup expired sessions
	err = repo.CleanupExpiredSessions()
	if err != nil {
		t.Fatalf("CleanupExpiredSessions() error = %v", err)
	}
}

