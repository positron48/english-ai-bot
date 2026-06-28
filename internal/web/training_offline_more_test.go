package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/i18n"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestHandleTrainingOfflinePack_GuardsAndEmptyQueue(t *testing.T) {
	router, _, userID, _, cleanup := setupTrainingOfflineTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/training/offline/pack", nil)
	w := httptest.NewRecorder()
	router.handleTrainingOfflinePack(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST expected 405, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/training/offline/pack", nil)
	w = httptest.NewRecorder()
	router.handleTrainingOfflinePack(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no user expected 401, got %d", w.Code)
	}

	// User with no due cards → empty queue payload.
	freshUserRepo := repository.NewUserRepository(router.db, router.logger)
	freshUser, err := freshUserRepo.GetOrCreateUser(991002)
	if err != nil {
		t.Fatalf("create fresh user: %v", err)
	}
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, freshUser.ID))
	w = httptest.NewRecorder()
	router.handleTrainingOfflinePack(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("empty queue expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if int(resp["total_cards"].(float64)) != 0 {
		t.Fatalf("expected empty queue, got %+v", resp)
	}

	// Service unavailable when options/training nil.
	bare := NewRouter(zap.NewNop(), &config.Config{Learning: config.DefaultLearningConfig()}, router.db, nil, nil, nil, nil)
	req = httptest.NewRequest(http.MethodGet, "/api/training/offline/pack", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w = httptest.NewRecorder()
	bare.handleTrainingOfflinePack(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("nil services expected 500, got %d", w.Code)
	}
}

func TestTrainingOfflineHelpers(t *testing.T) {
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "a", WordRU: "а"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordEN: "b", WordRU: "б"}},
		nil,
	}
	if got := indexOfCardInQueue(cards, 2); got != 1 {
		t.Fatalf("indexOfCardInQueue(2) = %d, want 1", got)
	}
	if got := indexOfCardInQueue(cards, 99); got != -1 {
		t.Fatalf("indexOfCardInQueue(missing) = %d, want -1", got)
	}

	ens := collectWordENs(cards, 1)
	if !ens["a"] || ens["b"] {
		t.Fatalf("collectWordENs exclude index 1 = %+v", ens)
	}
	rus := collectWordRUs(cards, 0)
	if rus["а"] || !rus["б"] {
		t.Fatalf("collectWordRUs exclude index 0 = %+v", rus)
	}

	now := time.Now()
	if fallbackTime(time.Time{}).Before(now.Add(-time.Minute)) {
		t.Fatal("fallbackTime zero should return recent time")
	}
	if !fallbackTime(now).Equal(now) {
		t.Fatal("fallbackTime should preserve non-zero")
	}
}

func TestTrainingSessionConfigForUser_UserSettings(t *testing.T) {
	router, _, userID, _, cleanup := setupTrainingOfflineTest(t)
	defer cleanup()

	spellOff := false
	spellTh := 80
	typeOff := false
	typeTh := 90
	settings := models.UserSettings{
		SpellModeEnabled:        &spellOff,
		SpellMasteringThreshold: &spellTh,
		TypeModeEnabled:         &typeOff,
		TypeMasteringThreshold:  &typeTh,
	}
	raw, _ := json.Marshal(settings)
	userRepo := repository.NewUserRepository(router.db, router.logger)
	if err := userRepo.UpdateUserSettings(userID, string(raw)); err != nil {
		t.Fatalf("UpdateUserSettings: %v", err)
	}

	cfg := router.trainingSessionConfigForUser(userID)
	if cfg.SpellEnabled || cfg.TypeEnabled {
		t.Fatalf("expected spell/type disabled from settings: %+v", cfg)
	}
	if cfg.SpellMasteringThreshold != 80 || cfg.TypeMasteringThreshold != 90 {
		t.Fatalf("thresholds = spell %d type %d", cfg.SpellMasteringThreshold, cfg.TypeMasteringThreshold)
	}

	// Clamp out-of-range thresholds.
	low, high := -5, 150
	settings.SpellMasteringThreshold = &low
	settings.TypeMasteringThreshold = &high
	raw, _ = json.Marshal(settings)
	if err := userRepo.UpdateUserSettings(userID, string(raw)); err != nil {
		t.Fatalf("UpdateUserSettings clamp: %v", err)
	}
	cfg = router.trainingSessionConfigForUser(userID)
	if cfg.SpellMasteringThreshold != 0 || cfg.TypeMasteringThreshold != 100 {
		t.Fatalf("clamped thresholds = spell %d type %d", cfg.SpellMasteringThreshold, cfg.TypeMasteringThreshold)
	}
}

func TestBuildOfflineWordTrainingCard_DirectionsAndLang(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{Learning: config.LearningConfig{TargetLang: "en", NativeLang: "ru"}}
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	router := NewRouter(logger, cfg, db, nil, nil, service.NewOptionsService(trainingCardRepo, logger, "en"), nil)

	pos := "noun"
	display := "to run"
	card := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: 10, Direction: models.DirectionENtoRU, State: models.StateReview, EF: 2.5},
		TrainingCard: models.TrainingCard{
			ID: 5, WordCardID: 1, WordEN: "run", WordRU: "бежать", Transcription: "[rʌn]",
			POS: &pos, DisplayWord: &display, ExampleEN: "Run fast", Hint: "verb",
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(i18n.WithLanguage(req.Context(), "ru"))
	got := router.buildOfflineWordTrainingCard(req, "ru", card, []string{"бежать", "идти"}, "бежать")
	if got.WordEN != "run" || got.WordRU != "" {
		t.Fatalf("ENtoRU card fields = %+v", got)
	}
	if got.Question == "" || got.Morph == nil {
		t.Fatalf("expected question and morph: %+v", got)
	}

	card.UserCard.Direction = models.DirectionRUtoEN
	req = req.WithContext(i18n.WithLanguage(req.Context(), "es"))
	got = router.buildOfflineWordTrainingCard(req, "es", card, []string{"run", "walk"}, "run")
	if got.WordRU != "бежать" || got.WordEN != "" {
		t.Fatalf("RUtoEN card fields = %+v", got)
	}
}

func TestHandleTrainingOfflineSyncAttempts_ErrorPaths(t *testing.T) {
	router, _, userID, userCardRepo, cleanup := setupTrainingOfflineTest(t)
	defer cleanup()
	trainingCardRepo := repository.NewTrainingCardRepository(router.db, router.logger)
	userCardID := seedUserCardForOfflineSync(t, router.db, userCardRepo, trainingCardRepo, userID)

	payload := map[string]interface{}{
		"attempts": []map[string]interface{}{
			{"client_attempt_id": "", "user_card_id": userCardID},
			{"client_attempt_id": "dup-1", "user_card_id": 999999},
		},
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/training/offline/sync-attempts", bytes.NewReader(raw))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w := httptest.NewRecorder()
	router.handleTrainingOfflineSyncAttempts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with per-item errors, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Synced  int                      `json:"synced"`
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Synced != 0 || len(resp.Results) != 2 {
		t.Fatalf("unexpected sync response: %+v", resp)
	}

	// Successful card attempt then duplicate.
	okPayload := map[string]interface{}{
		"attempts": []map[string]interface{}{
			{
				"client_attempt_id": "dup-card-1",
				"user_card_id":      userCardID,
				"mode":              "card",
				"options":           []string{"offline", "x", "y", "z"},
				"chosen_option":     "offline",
				"correct_answer":    "offline",
			},
		},
	}
	raw, _ = json.Marshal(okPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/training/offline/sync-attempts", bytes.NewReader(raw))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w = httptest.NewRecorder()
	router.handleTrainingOfflineSyncAttempts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first card sync: %d %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/training/offline/sync-attempts", bytes.NewReader(raw))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w = httptest.NewRecorder()
	router.handleTrainingOfflineSyncAttempts(w, req)
	var dupResp struct {
		Results []map[string]interface{} `json:"results"`
	}
	_ = json.NewDecoder(w.Body).Decode(&dupResp)
	if dup, _ := dupResp.Results[0]["duplicate"].(bool); !dup {
		t.Fatalf("expected duplicate=true on second sync: %+v", dupResp.Results[0])
	}

	// Training card mismatch.
	mismatchPayload := map[string]interface{}{
		"attempts": []map[string]interface{}{
			{
				"client_attempt_id":  "mismatch-1",
				"user_card_id":       userCardID,
				"training_card_id":   99999,
				"mode":               "card",
				"options":            []string{"offline"},
				"chosen_option":      "offline",
				"correct_answer":     "offline",
			},
		},
	}
	raw, _ = json.Marshal(mismatchPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/training/offline/sync-attempts", bytes.NewReader(raw))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w = httptest.NewRecorder()
	router.handleTrainingOfflineSyncAttempts(w, req)
	var mismatchResp struct {
		Results []map[string]interface{} `json:"results"`
	}
	_ = json.NewDecoder(w.Body).Decode(&mismatchResp)
	if errVal, _ := mismatchResp.Results[0]["error"].(string); errVal != "training_card_mismatch" {
		t.Fatalf("expected training_card_mismatch, got %+v", mismatchResp.Results[0])
	}

	// Type mode offline sync.
	typePayload := map[string]interface{}{
		"attempts": []map[string]interface{}{
			{
				"client_attempt_id": "type-1",
				"user_card_id":      userCardID,
				"mode":              "type",
				"answer_text":       "offline",
				"correct_answer":    "offline",
			},
		},
	}
	raw, _ = json.Marshal(typePayload)
	req = httptest.NewRequest(http.MethodPost, "/api/training/offline/sync-attempts", bytes.NewReader(raw))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w = httptest.NewRecorder()
	router.handleTrainingOfflineSyncAttempts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("type sync: %d %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/training/offline/sync-attempts", bytes.NewReader([]byte("{")))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w = httptest.NewRecorder()
	router.handleTrainingOfflineSyncAttempts(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json expected 400, got %d", w.Code)
	}
}
