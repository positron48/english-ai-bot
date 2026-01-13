package service

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestNotificationService_Start_ContextCancellation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	service := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	ctx, cancel := context.WithCancel(context.Background())

	// Start service in goroutine
	done := make(chan bool)
	go func() {
		service.Start(ctx)
		done <- true
	}()

	// Cancel context after short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for service to stop
	select {
	case <-done:
		// Service stopped, which is expected
	case <-time.After(1 * time.Second):
		t.Error("Service should stop when context is cancelled")
	}
}

func TestNotificationService_Start_StopChan(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	service := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	ctx := context.Background()

	// Start service in goroutine
	done := make(chan bool)
	go func() {
		service.Start(ctx)
		done <- true
	}()

	// Stop service after short delay
	time.Sleep(50 * time.Millisecond)
	service.Stop()

	// Wait for service to stop
	select {
	case <-done:
		// Service stopped, which is expected
	case <-time.After(1 * time.Second):
		t.Error("Service should stop when Stop() is called")
	}
}

func TestNotificationService_Stop(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	service := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)

	// Stop should not panic
	service.Stop()

	// Verify stopChan is closed
	select {
	case <-service.stopChan:
		// Channel is closed, which is expected
	default:
		t.Error("stopChan should be closed after Stop()")
	}
}
