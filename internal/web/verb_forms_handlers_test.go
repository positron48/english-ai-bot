package web

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupVerbFormsHandlerTest(t *testing.T) (*Router, *sql.DB, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{
		Learning: config.LearningConfig{
			TargetLang:      "es",
			NativeLang:      "ru",
			GrammarBundleID: "es",
		},
		Training: config.TrainingConfig{
			SpanishVerbFormsEnabled: true,
			VerbFormsMaxCards:       10,
			VerbFormsMaxNew:         5,
		},
		WebApp: config.WebAppConfig{JWTSecret: "verb-forms-test-secret"},
	}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(99001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	return router, db, user.ID
}

func verbFormsUserContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}

func seedVerbFormsHandlerVocab(t *testing.T, db *sql.DB, userID int64) (wordCardID int64) {
	t.Helper()
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ('hablar', 'говорить', 'verb') RETURNING id`).
		Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	var trainingCardID int64
	if err := db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES ($1, 'hablar', 0, 'говорить', 'to speak', 'verb') RETURNING id`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'es_ru', 'review')`,
		userID, trainingCardID); err != nil {
		t.Fatalf("insert user_card: %v", err)
	}
	lemmaID, err := repo.UpsertVerbLemma("hablar", "es", "test", "v1", "chk", `{"ru":{"gloss":"говорить"}}`)
	if err != nil {
		t.Fatalf("UpsertVerbLemma: %v", err)
	}
	if _, err := repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: lemmaID, Mood: "indicativo", Tense: "presente",
		Person: "1", Number: "singular", SurfaceForm: "hablo",
	}); err != nil {
		t.Fatalf("UpsertVerbForm: %v", err)
	}
	if err := repo.LinkWordCardToLemma(wordCardID, lemmaID, 1.0, "test"); err != nil {
		t.Fatalf("LinkWordCardToLemma: %v", err)
	}
	return wordCardID
}

func TestWriteVerbTrainingDisabled(t *testing.T) {
	router, _, _ := setupVerbFormsHandlerTest(t)
	rr := httptest.NewRecorder()
	router.writeVerbTrainingDisabled(rr)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "verb_training_disabled" {
		t.Fatalf("body=%v", body)
	}
}

func TestHandleVerbTrainingUpcoming_poolReady(t *testing.T) {
	router, db, userID := setupVerbFormsHandlerTest(t)
	seedVerbFormsHandlerVocab(t, db, userID)

	req := httptest.NewRequest(http.MethodGet, "/api/verb-training/upcoming", nil)
	req = verbFormsUserContext(req, userID)
	rr := httptest.NewRecorder()
	router.handleVerbTrainingUpcoming(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["enabled"] != true {
		t.Fatalf("expected enabled=true, got %v", resp["enabled"])
	}
	if poolReady, _ := resp["pool_ready"].(bool); !poolReady {
		t.Fatalf("expected pool_ready after vocab seed, got %v", resp)
	}
}

func TestHandleVerbTrainingUpcoming_disabledForEnglish(t *testing.T) {
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{
		Learning: config.LearningConfig{TargetLang: "en", GrammarBundleID: "en"},
		Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		WebApp:   config.WebAppConfig{JWTSecret: "verb-forms-test-secret"},
	}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(99002)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	req := httptest.NewRequest(http.MethodGet, "/api/verb-training/upcoming", nil)
	req = verbFormsUserContext(req, user.ID)
	rr := httptest.NewRecorder()
	router.handleVerbTrainingUpcoming(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleVerbTrainingStart_andAnswerFlow(t *testing.T) {
	router, db, userID := setupVerbFormsHandlerTest(t)
	seedVerbFormsHandlerVocab(t, db, userID)

	startReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/start", nil)
	startReq = verbFormsUserContext(startReq, userID)
	startRR := httptest.NewRecorder()
	router.handleVerbTrainingStart(startRR, startReq)
	if startRR.Code != http.StatusOK {
		t.Fatalf("start status=%d body=%s", startRR.Code, startRR.Body.String())
	}

	var card map[string]interface{}
	if err := json.Unmarshal(startRR.Body.Bytes(), &card); err != nil {
		t.Fatalf("decode start: %v", err)
	}
	uvcID, ok := card["user_verb_card_id"].(float64)
	if !ok || uvcID <= 0 {
		t.Fatalf("missing user_verb_card_id: %v", card)
	}
	sessionID, _ := card["session_id"].(float64)
	prompt, _ := card["prompt"].(map[string]interface{})
	if prompt == nil || strings.TrimSpace(promptString(prompt, "question")) == "" {
		t.Fatalf("expected hydrated prompt question: %v", card)
	}

	curReq := httptest.NewRequest(http.MethodGet, "/api/verb-training/current", nil)
	curReq = verbFormsUserContext(curReq, userID)
	curRR := httptest.NewRecorder()
	router.handleVerbTrainingCurrent(curRR, curReq)
	if curRR.Code != http.StatusOK {
		t.Fatalf("current status=%d", curRR.Code)
	}

	answerBody, _ := json.Marshal(map[string]interface{}{
		"user_verb_card_id": int64(uvcID),
		"answer":            "hablo",
	})
	ansReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader(answerBody))
	ansReq = verbFormsUserContext(ansReq, userID)
	ansRR := httptest.NewRecorder()
	router.handleVerbTrainingAnswer(ansRR, ansReq)
	if ansRR.Code != http.StatusOK {
		t.Fatalf("answer status=%d body=%s", ansRR.Code, ansRR.Body.String())
	}
	var feedback map[string]interface{}
	if err := json.Unmarshal(ansRR.Body.Bytes(), &feedback); err != nil {
		t.Fatal(err)
	}
	if feedback["is_correct"] != true {
		t.Fatalf("expected correct answer, got %v", feedback)
	}

	var eventCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM verb_review_events WHERE session_id=$1`, int64(sessionID)).Scan(&eventCount); err != nil {
		t.Fatal(err)
	}
	if eventCount != 1 {
		t.Fatalf("expected 1 review event, got %d", eventCount)
	}
}

func TestHandleVerbTrainingLemmaForms(t *testing.T) {
	router, db, userID := setupVerbFormsHandlerTest(t)
	seedVerbFormsHandlerVocab(t, db, userID)

	req := httptest.NewRequest(http.MethodGet, "/api/verb-training/forms-by-lemma?lemma=hablar", nil)
	req = verbFormsUserContext(req, userID)
	rr := httptest.NewRecorder()
	router.handleVerbTrainingLemmaForms(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	forms, ok := resp["forms"].([]interface{})
	if !ok || len(forms) == 0 {
		t.Fatalf("expected forms, got %v", resp)
	}
}

func TestHandleVocabVerbForms(t *testing.T) {
	router, db, userID := setupVerbFormsHandlerTest(t)
	wordCardID := seedVerbFormsHandlerVocab(t, db, userID)

	path := "/api/vocab/" + strconv.FormatInt(wordCardID, 10) + "/verb-forms"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req = verbFormsUserContext(req, userID)
	rr := httptest.NewRecorder()
	router.handleVocabVerbForms(rr, req, userID, wordCardID)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestGetUserVerbScopes_unlockLadderAlwaysUnlocked(t *testing.T) {
	router, _, userID := setupVerbFormsHandlerTest(t)
	scopes := router.getUserVerbScopes(context.Background(), userID)
	found := false
	for _, s := range scopes {
		if s == "es.presente.indicativo" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("unlock ladder should include always-unlocked presente, got %v", scopes)
	}
}

func TestParseWordCardIDForVerbForms(t *testing.T) {
	id, ok := parseWordCardIDForVerbForms("/api/vocab/42/verb-forms")
	if !ok || id != 42 {
		t.Fatalf("parse valid: id=%d ok=%v", id, ok)
	}
	if _, ok := parseWordCardIDForVerbForms("/api/vocab/0/verb-forms"); ok {
		t.Fatal("zero id should fail")
	}
	if _, ok := parseWordCardIDForVerbForms("/api/vocab/x/verb-forms"); ok {
		t.Fatal("invalid id should fail")
	}
}

func TestHandleVerbTrainingStart_methodNotAllowed(t *testing.T) {
	router, _, userID := setupVerbFormsHandlerTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/verb-training/start", nil)
	req = verbFormsUserContext(req, userID)
	rr := httptest.NewRecorder()
	router.handleVerbTrainingStart(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestHandleVerbTrainingCurrent_noSession(t *testing.T) {
	router, _, userID := setupVerbFormsHandlerTest(t)
	webVerbSessionsMu.Lock()
	delete(webVerbSessions, userID)
	webVerbSessionsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/verb-training/current", nil)
	req = verbFormsUserContext(req, userID)
	rr := httptest.NewRecorder()
	router.handleVerbTrainingCurrent(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
}
