package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/testutil"
)

func setupVocabDeletePostTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.TrainingCardRepository, *repository.UserCardRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)

	return db, userRepo, trainingCardRepo, userCardRepo
}

func TestHandleVocabDelete_Delete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo := setupVocabDeletePostTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(414141)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word_card (lemma)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "deletepost", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	trainingCard := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "deletepost",
		SenseIndex: 0,
		WordRU:     "удалить пост",
		MeaningEN:  "delete post",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user card
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	_, err = userCardRepo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create POST request for delete using lemma
	req := httptest.NewRequest("POST", "/api/vocab/deletepost/delete", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["success"] == nil {
		t.Error("Response should contain success field")
	}
	if response["lemma"] == nil {
		t.Error("Response should contain lemma field")
	}
	if response["lemma"].(string) != "deletepost" {
		t.Errorf("Expected lemma 'deletepost', got %v", response["lemma"])
	}
}

func TestHandleVocabDelete_Delete_NoCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _ := setupVocabDeletePostTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(424242)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create POST request for delete (word doesn't exist)
	req := httptest.NewRequest("POST", "/api/vocab/nonexistent/delete", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	// Should return 404 (word not found)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleVocabDelete_DeleteByWordCardID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo := setupVocabDeletePostTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(434343)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word_card (lemma)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "testword", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create two training cards with same word_card_id
	trainingCard1 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "testword",
		SenseIndex: 0,
		WordRU:     "тестовое слово",
		MeaningEN:  "test word",
	}
	trainingCardID1, err := trainingCardRepo.CreateTrainingCard(trainingCard1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}

	trainingCard2 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "testword",
		SenseIndex: 1,
		WordRU:     "тестовое слово 2",
		MeaningEN:  "test word 2",
	}
	trainingCardID2, err := trainingCardRepo.CreateTrainingCard(trainingCard2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Create user cards for both training cards
	userCard1 := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	_, err = userCardRepo.CreateUserCard(userCard1)
	if err != nil {
		t.Fatalf("Failed to create user card 1: %v", err)
	}

	userCard2 := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID2,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateLearning,
		EF:             models.InitialEF,
	}
	_, err = userCardRepo.CreateUserCard(userCard2)
	if err != nil {
		t.Fatalf("Failed to create user card 2: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create POST request for delete using lemma
	req := httptest.NewRequest("POST", "/api/vocab/testword/delete", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["success"] == nil || !response["success"].(bool) {
		t.Error("Response should contain success=true")
	}
	if response["rows_affected"] == nil {
		t.Error("Response should contain rows_affected field")
	}
	rowsAffected := int64(response["rows_affected"].(float64))
	if rowsAffected != 2 {
		t.Errorf("Expected 2 rows affected (both user_cards deleted), got %d", rowsAffected)
	}

	// Verify both user_cards are deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user_cards WHERE user_id = ? AND (training_card_id = ? OR training_card_id = ?)",
		user.ID, trainingCardID1, trainingCardID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check user cards: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 user_cards remaining, got %d", count)
	}
}
