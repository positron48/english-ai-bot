package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupGrammarTest(t *testing.T) (*Router, *database.DB, func()) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	// Initialize grammar repositories and service
	contentRepo := repository.NewGrammarContentRepository(logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	grammarService := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, logger)
	router.SetGrammarService(grammarService)

	cleanup := func() {
		db.Close()
	}

	return router, db, cleanup
}

func TestHandleLearningGrammarSubmitTest_BadRequest(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	// Invalid JSON body
	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/tests/submit", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarSubmitTest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestHandleLearningGrammarSubmitTest_MissingFields(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	// Missing required fields
	body := map[string]interface{}{}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/tests/submit", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarSubmitTest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestHandleLearningGrammarChapter_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/chapters/test-chapter", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarChapter(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleLearningGrammarChapterTest_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/chapters/test-chapter/test", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarChapterTest(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

// TestHandleLearningGrammarPlacementTest_QuestionsHaveChapterTitle verifies that
// placement test questions include placement_chapter_title (chapter name, not theory block).
func TestHandleLearningGrammarPlacementTest_QuestionsHaveChapterTitle(t *testing.T) {
	router, db, cleanup := setupGrammarTest(t)
	defer cleanup()

	// Publish one chapter so the placement test has questions to return
	_, err := db.GetConnection().Exec(
		`INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at) 
		 VALUES (?, ?, 1, datetime('now')) 
		 ON CONFLICT(item_type, item_id) DO UPDATE SET is_published=1, updated_at=datetime('now')`,
		"chapter", "en.grammar.first_sentences_be_as.personal_pronouns_am_is")
	if err != nil {
		t.Fatalf("Failed to publish chapter: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/placement-test", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarPlacementTest(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp struct {
		Questions []map[string]interface{} `json:"questions"`
		Total     int                     `json:"total"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Questions) == 0 {
		t.Skip("No questions in placement test (chapter may have no question_bank); skipping chapter title check")
		return
	}

	for i, q := range resp.Questions {
		if q["placement_chapter_id"] == nil {
			continue
		}
		title, _ := q["placement_chapter_title"].(string)
		if title == "" {
			t.Errorf("Question %d has placement_chapter_id but placement_chapter_title is missing or empty", i+1)
		}
	}
}

// TestHandleLearningGrammarSubmitPlacementTest_ReturnsLevelAndResults verifies that
// placement test submit returns level, opened_sections, and results (all questions with
// user answer, correct answer, and placement_chapter_title).
func TestHandleLearningGrammarSubmitPlacementTest_ReturnsLevelAndResults(t *testing.T) {
	router, db, cleanup := setupGrammarTest(t)
	defer cleanup()

	_, err := db.GetConnection().Exec(
		`INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at) 
		 VALUES (?, ?, 1, datetime('now')) 
		 ON CONFLICT(item_type, item_id) DO UPDATE SET is_published=1, updated_at=datetime('now')`,
		"chapter", "en.grammar.first_sentences_be_as.personal_pronouns_am_is")
	if err != nil {
		t.Fatalf("Failed to publish chapter: %v", err)
	}

	// Get questions to build answers
	getReq := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/placement-test", nil)
	getReq = setUserIDInContext(getReq, 1)
	getW := httptest.NewRecorder()
	router.handleLearningGrammarPlacementTest(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("GET placement-test: expected 200, got %d: %s", getW.Code, getW.Body.String())
	}

	var getResp struct {
		Questions []map[string]interface{} `json:"questions"`
	}
	if err := json.NewDecoder(getW.Body).Decode(&getResp); err != nil {
		t.Fatalf("Failed to decode GET response: %v", err)
	}

	answers := make(map[string]interface{})
	for _, q := range getResp.Questions {
		id, _ := q["id"].(string)
		if id == "" {
			continue
		}
		// Use a placeholder wrong answer; backend will still return correct_answer in results
		answers[id] = "__wrong__"
	}

	bodyJSON, _ := json.Marshal(answers)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/placement-test/submit", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarSubmitPlacementTest(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp struct {
		Level          string        `json:"level"`
		Score          int           `json:"score"`
		Correct        int           `json:"correct"`
		TotalQuestions int           `json:"total_questions"`
		OpenedSections []string      `json:"opened_sections"`
		Results        []interface{} `json:"results"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode submit response: %v", err)
	}

	// Required fields
	if resp.Level == "" {
		t.Error("submit response must include non-empty level")
	}
	if resp.OpenedSections == nil {
		t.Error("submit response must include opened_sections array")
	}
	if resp.Results == nil {
		t.Error("submit response must include results array")
	}
	if len(resp.Results) != len(answers) {
		t.Errorf("results length %d != answers length %d", len(resp.Results), len(answers))
	}

	for i, r := range resp.Results {
		m, ok := r.(map[string]interface{})
		if !ok {
			t.Errorf("results[%d] is not an object", i)
			continue
		}
		if m["question_id"] == nil {
			t.Errorf("results[%d] missing question_id", i)
		}
		if _, has := m["correct"]; !has {
			t.Errorf("results[%d] missing correct", i)
		}
		if _, has := m["user_answer"]; !has {
			t.Errorf("results[%d] missing user_answer", i)
		}
		if _, has := m["correct_answer"]; !has {
			t.Errorf("results[%d] missing correct_answer", i)
		}
		if m["placement_chapter_title"] == nil {
			t.Errorf("results[%d] missing placement_chapter_title", i)
		}
	}
}
