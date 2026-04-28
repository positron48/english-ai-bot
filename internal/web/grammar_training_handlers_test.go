package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupGrammarTrainingHandlersTest(t *testing.T) (*Router, int64, func()) {
	t.Helper()
	router, db, userID, cleanup := setupGrammarTest(t)
	logger := zap.NewNop()
	contentFS := fstest.MapFS{
		"sections.json": {Data: []byte(`{"version":"1","sections":[{"section_id":"s1","title":"S1","level":"A1","order":1,"chapter_ids":["ch1"]}]}`)},
		"index.json":    {Data: []byte(`{"version":"1","generated_at":"","chapters":{"ch1":"one.json"}}`)},
		"chapters/one.json": {Data: []byte(`{
			"schema_version":"1",
			"id":"ch1",
			"section_id":"s1",
			"title":"Chapter 1",
			"blocks":[{"id":"b1","type":"theory","theory":{"title":"T","content":"C"}}],
			"question_bank":{"questions":[{"id":"q1","type":"single_choice","question":"Q?","options":["A","B"],"correct_answer":"A","theory_block_id":"b1","chapter_id":"ch1","concept_id":"c1"}]},
			"chapter_test":{"selection_strategy":{"type":"random"},"pool_question_ids":["q1"],"num_questions":1}
		}`)},
	}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(contentFS, logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	lc := config.DefaultLearningConfig()
	lc.TargetLang = "es"
	lc.GrammarBundleID = "es"
	gs := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, lc, logger)
	packFS := fstest.MapFS{
		"index.json": {Data: []byte(`{"version":"1","language":"es","course_id":"es","generated_at":"","chapters":{"ch1":"one_questions.json"}}`)},
		"chapters/one_questions.json": {Data: []byte(`{"chapter_id":"ch1","questions":[{"id":"q1","chapter_id":"ch1","theory_block_id":"b1","concept_id":"c1","type":"single_choice","question":"Q?","options":["A","B"],"correct_answer":"A","explanation":"E"}]}`)},
	}
	gs.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(packFS, logger))
	gs.SetSRSRepository(repository.NewGrammarSRSRepository(db.GetConnection(), logger))
	router.SetGrammarService(gs)

	_ = publishRepo.SetPublished("section", "s1", true, nil)
	_ = publishRepo.SetPublished("chapter", "ch1", true, nil)
	_ = attemptRepo.SavePlacementTestResult(userID, 100, 100, []string{"s1"})

	return router, userID, cleanup
}

func TestGrammarTrainingHandlers_AvailabilityStartAnswer_Success(t *testing.T) {
	router, userID, cleanup := setupGrammarTrainingHandlersTest(t)
	defer cleanup()

	reqAvail := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/training/availability", nil)
	reqAvail = setUserIDInContext(reqAvail, userID)
	wAvail := httptest.NewRecorder()
	router.handleLearningGrammarTrainingAvailability(wAvail, reqAvail)
	if wAvail.Code != http.StatusOK {
		t.Fatalf("availability: expected 200, got %d", wAvail.Code)
	}

	reqStart := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/start", bytes.NewReader([]byte(`{"limit":1}`)))
	reqStart = setUserIDInContext(reqStart, userID)
	wStart := httptest.NewRecorder()
	router.handleLearningGrammarTrainingStart(wStart, reqStart)
	if wStart.Code != http.StatusOK {
		t.Fatalf("start: expected 200, got %d: %s", wStart.Code, wStart.Body.String())
	}
	var session struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.NewDecoder(wStart.Body).Decode(&session); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if len(session.Items) == 0 {
		t.Fatal("expected at least one session item")
	}
	q, _ := session.Items[0]["question"].(map[string]interface{})
	qID, _ := q["id"].(string)
	if qID == "" {
		t.Fatal("expected question id")
	}
	body, _ := json.Marshal(map[string]interface{}{"question_id": qID, "answer": "A"})
	reqAnswer := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/answer", bytes.NewReader(body))
	reqAnswer = setUserIDInContext(reqAnswer, userID)
	wAnswer := httptest.NewRecorder()
	router.handleLearningGrammarTrainingAnswer(wAnswer, reqAnswer)
	if wAnswer.Code != http.StatusOK {
		t.Fatalf("answer: expected 200, got %d: %s", wAnswer.Code, wAnswer.Body.String())
	}
}

func TestGrammarTrainingHandlers_ValidationAndNotFound(t *testing.T) {
	router, userID, cleanup := setupGrammarTrainingHandlersTest(t)
	defer cleanup()

	reqBadJSON := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/answer", bytes.NewReader([]byte(`{`)))
	reqBadJSON = setUserIDInContext(reqBadJSON, userID)
	wBadJSON := httptest.NewRecorder()
	router.handleLearningGrammarTrainingAnswer(wBadJSON, reqBadJSON)
	if wBadJSON.Code != http.StatusBadRequest {
		t.Fatalf("bad json: expected 400, got %d", wBadJSON.Code)
	}

	reqMissingID := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/answer", bytes.NewReader([]byte(`{"answer":"A"}`)))
	reqMissingID = setUserIDInContext(reqMissingID, userID)
	wMissingID := httptest.NewRecorder()
	router.handleLearningGrammarTrainingAnswer(wMissingID, reqMissingID)
	if wMissingID.Code != http.StatusBadRequest {
		t.Fatalf("missing question_id: expected 400, got %d", wMissingID.Code)
	}

	reqNotFound := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/answer", bytes.NewReader([]byte(`{"question_id":"missing","answer":"A"}`)))
	reqNotFound = setUserIDInContext(reqNotFound, userID)
	wNotFound := httptest.NewRecorder()
	router.handleLearningGrammarTrainingAnswer(wNotFound, reqNotFound)
	if wNotFound.Code != http.StatusNotFound {
		t.Fatalf("not found: expected 404, got %d", wNotFound.Code)
	}
}

func TestGrammarTrainingHandlers_MethodAndAuthGuards(t *testing.T) {
	router, userID, cleanup := setupGrammarTrainingHandlersTest(t)
	defer cleanup()

	tests := []struct {
		name   string
		call   func(http.ResponseWriter, *http.Request)
		req    *http.Request
		status int
	}{
		{
			name: "availability method not allowed",
			call: router.handleLearningGrammarTrainingAvailability,
			req:  httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/availability", nil),
			status: http.StatusMethodNotAllowed,
		},
		{
			name: "start unauthorized",
			call: router.handleLearningGrammarTrainingStart,
			req:  httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/start", bytes.NewReader([]byte(`{}`))),
			status: http.StatusUnauthorized,
		},
		{
			name: "start method not allowed",
			call: router.handleLearningGrammarTrainingStart,
			req:  httptest.NewRequest(http.MethodGet, "/api/learning/grammar/training/session/start", nil),
			status: http.StatusMethodNotAllowed,
		},
		{
			name: "answer method not allowed",
			call: router.handleLearningGrammarTrainingAnswer,
			req:  setUserIDInContext(httptest.NewRequest(http.MethodGet, "/api/learning/grammar/training/session/answer", nil), userID),
			status: http.StatusMethodNotAllowed,
		},
		{
			name: "answer unauthorized",
			call: router.handleLearningGrammarTrainingAnswer,
			req:  httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/answer", bytes.NewReader([]byte(`{"question_id":"q1","answer":"A"}`))),
			status: http.StatusUnauthorized,
		},
		{
			name: "availability unauthorized",
			call: router.handleLearningGrammarTrainingAvailability,
			req:  httptest.NewRequest(http.MethodGet, "/api/learning/grammar/training/availability", nil),
			status: http.StatusUnauthorized,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.call(w, tc.req.WithContext(context.Background()))
			if w.Code != tc.status {
				t.Fatalf("expected %d, got %d", tc.status, w.Code)
			}
		})
	}
}

func TestGrammarTrainingHandlers_500Branches(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()
	logger := zap.NewNop()

	// Broken service dependencies to force internal errors.
	badContent := repository.NewGrammarContentRepositoryWithFS(fstest.MapFS{}, logger)
	gs := service.NewGrammarService(
		badContent,
		repository.NewGrammarPublishRepository(testutil.SetupTestDatabase(t).GetConnection(), logger),
		repository.NewGrammarAttemptRepository(testutil.SetupTestDatabase(t).GetConnection(), logger),
		config.DefaultLearningConfig(),
		logger,
	)
	gs.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryWithFS(fstest.MapFS{}, logger))
	router.SetGrammarService(gs)

	reqAvail := setUserIDInContext(httptest.NewRequest(http.MethodGet, "/api/learning/grammar/training/availability", nil), 1)
	wAvail := httptest.NewRecorder()
	router.handleLearningGrammarTrainingAvailability(wAvail, reqAvail)
	if wAvail.Code != http.StatusInternalServerError {
		t.Fatalf("availability 500: expected 500, got %d", wAvail.Code)
	}

	reqStart := setUserIDInContext(httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/start", bytes.NewReader([]byte(`{}`))), 1)
	wStart := httptest.NewRecorder()
	router.handleLearningGrammarTrainingStart(wStart, reqStart)
	if wStart.Code != http.StatusInternalServerError {
		t.Fatalf("start 500: expected 500, got %d", wStart.Code)
	}

	reqAnswer := setUserIDInContext(httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/session/answer", bytes.NewReader([]byte(`{"question_id":"q1","answer":"A"}`))), 1)
	wAnswer := httptest.NewRecorder()
	router.handleLearningGrammarTrainingAnswer(wAnswer, reqAnswer)
	if wAnswer.Code != http.StatusInternalServerError {
		t.Fatalf("answer 500: expected 500, got %d", wAnswer.Code)
	}
}

