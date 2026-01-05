package repository

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestWebSessionRepository_UpdateLastSeen(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)
	defer db.Close()

	repo := NewWebSessionRepository(db, logger)

	// Create a session
	session := &WebSession{
		UserID:    111,
		Token:     "update-last-seen-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create a request with session cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "update-last-seen-token",
	})

	// Update last seen
	err = repo.UpdateLastSeen(req)
	if err != nil {
		t.Fatalf("UpdateLastSeen() error = %v", err)
	}
}

func TestWebSessionRepository_UpdateLastSeen_NoCookie(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebSessionTestDB(t)
	defer db.Close()

	repo := NewWebSessionRepository(db, logger)

	// Create a request without session cookie
	req := httptest.NewRequest("GET", "/test", nil)

	// Update last seen should return error
	err := repo.UpdateLastSeen(req)
	if err == nil {
		t.Error("UpdateLastSeen() should return error when no cookie")
	}
}
