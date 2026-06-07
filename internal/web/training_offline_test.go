package web

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupTrainingOfflineTest(t *testing.T) (*Router, *sql.DB, int64, *repository.UserCardRepository, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(991001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)

	cfg := &config.Config{
		Learning: config.DefaultLearningConfig(),
		Training: config.TrainingConfig{OptionsDelayMS: 0, WrongAnswerDelaySeconds: 1},
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, cfg.Learning, logger)
	srsService := service.NewSRSService(userCardRepo, cfg.Learning, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger, "en")
	router := NewRouter(logger, cfg, db, trainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	return router, db, user.ID, userCardRepo, func() {}
}

func seedUserCardForOfflineSync(t *testing.T, db *sql.DB, userCardRepo *repository.UserCardRepository, trainingCardRepo *repository.TrainingCardRepository, userID int64) int64 {
	t.Helper()
	var wordCardID int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id`, "offline", "offline").Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "offline",
		SenseIndex: 0,
		WordRU:     "офлайн",
		MeaningEN:  "offline",
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}
	past := time.Now().Add(-time.Hour)
	userCardID, err := userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         userID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             2.5,
		NextDueAt:      &past,
	})
	if err != nil {
		t.Fatalf("CreateUserCard: %v", err)
	}
	return userCardID
}

func TestHandleTrainingOfflineSyncAttempts_PreservesAnsweredAt(t *testing.T) {
	router, _, userID, userCardRepo, cleanup := setupTrainingOfflineTest(t)
	defer cleanup()
	trainingCardRepo := repository.NewTrainingCardRepository(router.db, router.logger)
	userCardID := seedUserCardForOfflineSync(t, router.db, userCardRepo, trainingCardRepo, userID)

	yesterday := time.Now().UTC().Add(-26 * time.Hour).Truncate(time.Second)
	payload := map[string]interface{}{
		"attempts": []map[string]interface{}{
			{
				"client_attempt_id": "offline-card-yesterday",
				"user_card_id":      userCardID,
				"direction":         string(models.DirectionRUtoEN),
				"mode":              "card",
				"shown_at":          yesterday.Format(time.RFC3339),
				"options_shown_at":  yesterday.Add(500 * time.Millisecond).Format(time.RFC3339),
				"answered_at":       yesterday.Add(2 * time.Second).Format(time.RFC3339),
				"answer_time_ms":    1500,
				"t_delay_ms":        500,
				"options":           []string{"offline", "wrong1", "wrong2", "wrong3"},
				"chosen_option":     "offline",
				"correct_answer":    "offline",
			},
		},
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/training/offline/sync-attempts", bytes.NewReader(raw))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w := httptest.NewRecorder()
	router.handleTrainingOfflineSyncAttempts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	userCard, err := userCardRepo.GetUserCard(userCardID)
	if err != nil {
		t.Fatalf("GetUserCard: %v", err)
	}
	if userCard.LastReviewAt == nil {
		t.Fatal("expected last_review_at to be set")
	}
	diff := userCard.LastReviewAt.UTC().Sub(yesterday.Add(2 * time.Second))
	if diff < -time.Second || diff > time.Second {
		t.Fatalf("expected last_review_at near offline answered_at, got %v diff %v", userCard.LastReviewAt, diff)
	}

	sessionRepo := repository.NewSessionRepository(router.db, router.logger)
	var answeredAt time.Time
	err = router.db.QueryRow(`
		SELECT answered_at FROM review_events
		WHERE user_id = $1 AND client_attempt_id = $2
	`, userID, "offline-card-yesterday").Scan(&answeredAt)
	if err != nil {
		t.Fatalf("query review_events.answered_at: %v", err)
	}
	if exists, err := sessionRepo.HasReviewEventClientAttempt(userID, "offline-card-yesterday"); err != nil || !exists {
		t.Fatalf("expected review event, exists=%v err=%v", exists, err)
	}
	reviewDiff := answeredAt.UTC().Sub(yesterday.Add(2 * time.Second))
	if reviewDiff < -time.Second || reviewDiff > time.Second {
		t.Fatalf("expected review_events.answered_at near offline time, got %v diff %v", answeredAt, reviewDiff)
	}
}

func TestHandleTrainingOfflineSyncAttempts_SpellType(t *testing.T) {
	router, _, userID, userCardRepo, cleanup := setupTrainingOfflineTest(t)
	defer cleanup()
	trainingCardRepo := repository.NewTrainingCardRepository(router.db, router.logger)
	userCardID := seedUserCardForOfflineSync(t, router.db, userCardRepo, trainingCardRepo, userID)

	now := time.Now().UTC()
	payload := map[string]interface{}{
		"attempts": []map[string]interface{}{
			{
				"client_attempt_id": "offline-spell-1",
				"user_card_id":      userCardID,
				"direction":         "spell",
				"mode":              "spell",
				"shown_at":          now.Format(time.RFC3339),
				"options_shown_at":  now.Format(time.RFC3339),
				"answered_at":       now.Add(2 * time.Second).Format(time.RFC3339),
				"answer_time_ms":    2000,
				"answer_text":       "offline",
				"correct_answer":    "offline",
			},
		},
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/training/offline/sync-attempts", bytes.NewReader(raw))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w := httptest.NewRecorder()
	router.handleTrainingOfflineSyncAttempts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	sessionRepo := repository.NewSessionRepository(router.db, router.logger)
	exists, err := sessionRepo.HasReviewEventClientAttempt(userID, "offline-spell-1")
	if err != nil {
		t.Fatalf("HasReviewEventClientAttempt: %v", err)
	}
	if !exists {
		t.Fatal("expected review event after spell offline sync")
	}
}

func TestHandleTrainingOfflinePack_ReturnsV2Queue(t *testing.T) {
	router, db, userID, userCardRepo, cleanup := setupTrainingOfflineTest(t)
	defer cleanup()
	trainingCardRepo := repository.NewTrainingCardRepository(db, router.logger)
	_ = seedUserCardForOfflineSync(t, db, userCardRepo, trainingCardRepo, userID)

	req := httptest.NewRequest(http.MethodGet, "/api/training/offline/pack", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w := httptest.NewRecorder()
	router.handleTrainingOfflinePack(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["algo_version"] != "word_training_offline_v2_queue" {
		t.Fatalf("unexpected algo_version: %v", resp["algo_version"])
	}
	queue, ok := resp["queue"].([]interface{})
	if !ok || len(queue) == 0 {
		t.Fatalf("expected non-empty queue, got %#v", resp["queue"])
	}
}
