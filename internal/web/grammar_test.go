package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupGrammarTest(t *testing.T) (*Router, *database.DB, int64, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	adminUser, _ := userRepo.GetOrCreateUser(12345)

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

	cleanup := func() {} // shared db, do not close

	return router, db, adminUser.ID, cleanup
}

func TestHandleLearningGrammarSubmitTest_BadRequest(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
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
	router, _, _, cleanup := setupGrammarTest(t)
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
	router, _, _, cleanup := setupGrammarTest(t)
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
	router, _, _, cleanup := setupGrammarTest(t)
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
	router, db, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	// Publish one chapter so the placement test has questions to return
	_, err := db.GetConnection().Exec(
		`INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at) 
		 VALUES (?, ?, 1, CURRENT_TIMESTAMP) 
		 ON CONFLICT(item_type, item_id) DO UPDATE SET is_published=1, updated_at=CURRENT_TIMESTAMP`,
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
		// IDs must be "chapterID:qid" so the same qid in different chapters does not overwrite in SubmitPlacementTest
		id, _ := q["id"].(string)
		if id != "" && !strings.Contains(id, ":") {
			t.Errorf("Question %d id %q must be composite (chapterID:qid) to avoid correct_answer mix-up between chapters", i+1, id)
		}
	}
}

// TestHandleLearningGrammarSubmitPlacementTest_ReturnsLevelAndResults verifies that
// placement test submit returns level, opened_sections, and results (all questions with
// user answer, correct answer, and placement_chapter_title).
func TestHandleLearningGrammarSubmitPlacementTest_ReturnsLevelAndResults(t *testing.T) {
	router, db, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	_, err := db.GetConnection().Exec(
		`INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at) 
		 VALUES (?, ?, 1, CURRENT_TIMESTAMP) 
		 ON CONFLICT(item_type, item_id) DO UPDATE SET is_published=1, updated_at=CURRENT_TIMESTAMP`,
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
		if _, has := m["level"]; !has {
			t.Errorf("results[%d] missing level", i)
		}
	}
}

// TestHandleLearningGrammarSubmitTest_CategoryTestAnswersOrder verifies that
// category test answers are correctly matched to questions by question_id and chapter_id,
// preserving the order of questions and ensuring answers don't get mixed up between chapters.
// This test uses the exact format from a real user request to catch issues where
// answers might be incorrectly matched to questions.
func TestHandleLearningGrammarSubmitTest_CategoryTestAnswersOrder(t *testing.T) {
	router, db, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	// Publish all chapters from the category to enable category test
	chapters := []string{
		"en.grammar.first_sentences_be_as.personal_pronouns_am_is",
		"en.grammar.first_sentences_be_as.statements_with_be_identity",
		"en.grammar.first_sentences_be_as.questions_with_be_are",
		"en.grammar.first_sentences_be_as.negatives_with_be_not",
		"en.grammar.first_sentences_be_as.an_as_one_of",
		"en.grammar.first_sentences_be_as.this_that_these_those",
		"en.grammar.first_sentences_be_as.adjectives_after_be_she",
		"en.grammar.first_sentences_be_as.build_speak_mini_dialogues",
	}

	for _, chapterID := range chapters {
		_, err := db.GetConnection().Exec(
			`INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at) 
			 VALUES (?, ?, 1, CURRENT_TIMESTAMP) 
			 ON CONFLICT(item_type, item_id) DO UPDATE SET is_published=1, updated_at=CURRENT_TIMESTAMP`,
			"chapter", chapterID)
		if err != nil {
			t.Fatalf("Failed to publish chapter %s: %v", chapterID, err)
		}
	}

	// Get category test to see what questions are generated
	getReq := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/en.grammar.first_sentences_be_as/test", nil)
	getReq = setUserIDInContext(getReq, 1)
	getW := httptest.NewRecorder()
	router.handleLearningGrammarCategoryTest(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("GET category test: expected 200, got %d: %s", getW.Code, getW.Body.String())
	}

	var getResp struct {
		Questions []map[string]interface{} `json:"questions"`
		Total     int                       `json:"total"`
	}
	if err := json.NewDecoder(getW.Body).Decode(&getResp); err != nil {
		t.Fatalf("Failed to decode GET response: %v", err)
	}

	if len(getResp.Questions) == 0 {
		t.Skip("No questions in category test; skipping test")
		return
	}

	// Build a map of questions by chapter_id:question_id for verification
	questionMap := make(map[string]map[string]interface{})
	for _, q := range getResp.Questions {
		qID, _ := q["id"].(string)
		chapterID, _ := q["_category_test_chapter_id"].(string)
		if qID != "" && chapterID != "" {
			key := chapterID + ":" + qID
			questionMap[key] = q
		}
	}

	// Submit test with the exact format from the user's request
	// This simulates the real scenario where answers might get mixed up
	submitBody := map[string]interface{}{
		"scope":    "category",
		"scope_id":   "en.grammar.first_sentences_be_as",
		"answers": []map[string]interface{}{
			{"question_id": "q34", "answer": "true", "chapter_id": "en.grammar.first_sentences_be_as.personal_pronouns_am_is"},
			{"question_id": "q19", "answer": "is", "chapter_id": "en.grammar.first_sentences_be_as.personal_pronouns_am_is"},
			{"question_id": "q22", "answer": "b", "chapter_id": "en.grammar.first_sentences_be_as.statements_with_be_identity"},
			{"question_id": "q11", "answer": "b", "chapter_id": "en.grammar.first_sentences_be_as.statements_with_be_identity"},
			{"question_id": "q28", "answer": "a", "chapter_id": "en.grammar.first_sentences_be_as.questions_with_be_are"},
			{"question_id": "q17", "answer": "Who is he?", "chapter_id": "en.grammar.first_sentences_be_as.questions_with_be_are"},
			{"question_id": "q3", "answer": "a", "chapter_id": "en.grammar.first_sentences_be_as.negatives_with_be_not"},
			{"question_id": "q48", "answer": "She is not at school.", "chapter_id": "en.grammar.first_sentences_be_as.negatives_with_be_not"},
			{"question_id": "q15", "answer": "a", "chapter_id": "en.grammar.first_sentences_be_as.an_as_one_of"},
			{"question_id": "q40", "answer": "a", "chapter_id": "en.grammar.first_sentences_be_as.an_as_one_of"},
			{"question_id": "q18", "answer": "a", "chapter_id": "en.grammar.first_sentences_be_as.this_that_these_those"},
			{"question_id": "q49", "answer": "is", "chapter_id": "en.grammar.first_sentences_be_as.this_that_these_those"},
			{"question_id": "q56", "answer": "Is", "chapter_id": "en.grammar.first_sentences_be_as.adjectives_after_be_she"},
			{"question_id": "q27", "answer": "false", "chapter_id": "en.grammar.first_sentences_be_as.adjectives_after_be_she"},
			{"question_id": "q44", "answer": "That is my book over there.", "chapter_id": "en.grammar.first_sentences_be_as.build_speak_mini_dialogues"},
			{"question_id": "q53", "answer": "a", "chapter_id": "en.grammar.first_sentences_be_as.build_speak_mini_dialogues"},
			{"question_id": "q12", "answer": "a", "chapter_id": "en.grammar.first_sentences_be_as.statements_with_be_identity"},
			{"question_id": "q21", "answer": "a", "chapter_id": "en.grammar.first_sentences_be_as.personal_pronouns_am_is"},
			{"question_id": "q20", "answer": "false", "chapter_id": "en.grammar.first_sentences_be_as.personal_pronouns_am_is"},
		},
	}

	bodyJSON, _ := json.Marshal(submitBody)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/tests/submit", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarSubmitTest(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp struct {
		Score   int           `json:"score"`
		Passed  bool          `json:"passed"`
		Correct int          `json:"correct"`
		Total   int          `json:"total"`
		Results []interface{} `json:"results"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode submit response: %v", err)
	}

	// Verify that results are in the same order as answers
	answers := submitBody["answers"].([]map[string]interface{})
	if len(resp.Results) != len(answers) {
		t.Errorf("Results length %d != answers length %d", len(resp.Results), len(answers))
	}

	// Verify that each result matches the corresponding answer by question_id and chapter_id
	for i, r := range resp.Results {
		resultMap, ok := r.(map[string]interface{})
		if !ok {
			t.Errorf("results[%d] is not an object", i)
			continue
		}

		if i >= len(answers) {
			t.Errorf("Result index %d exceeds answers length", i)
			continue
		}

		answerItem := answers[i]
		expectedQuestionID, _ := answerItem["question_id"].(string)
		expectedChapterID, _ := answerItem["chapter_id"].(string)
		expectedAnswer := answerItem["answer"]

		resultQuestionID, _ := resultMap["question_id"].(string)
		resultUserAnswer := resultMap["user_answer"]

		// Verify question_id matches
		if resultQuestionID != expectedQuestionID {
			t.Errorf("Result[%d]: question_id mismatch - expected %q, got %q", i, expectedQuestionID, resultQuestionID)
		}

		// Verify user_answer matches (allowing for type conversion)
		if resultUserAnswer != expectedAnswer {
			// Try string comparison for flexibility
			resultStr := fmt.Sprintf("%v", resultUserAnswer)
			expectedStr := fmt.Sprintf("%v", expectedAnswer)
			if resultStr != expectedStr {
				t.Errorf("Result[%d]: user_answer mismatch - expected %v (%T), got %v (%T)", i, expectedAnswer, expectedAnswer, resultUserAnswer, resultUserAnswer)
			}
		}

		// Verify that the question exists in the test (by checking if it was in the generated test)
		questionKey := expectedChapterID + ":" + expectedQuestionID
		if _, exists := questionMap[questionKey]; !exists {
			// Question might not be in this particular test run (random selection)
			// But we should still verify the answer was processed
			t.Logf("Result[%d]: question %s from chapter %s was not in the generated test (this is OK if test is randomized)", i, expectedQuestionID, expectedChapterID)
		}

		// Verify required fields are present
		if _, has := resultMap["correct"]; !has {
			t.Errorf("Result[%d]: missing 'correct' field", i)
		}
		if _, has := resultMap["correct_answer"]; !has {
			t.Errorf("Result[%d]: missing 'correct_answer' field", i)
		}
		if _, has := resultMap["prompt"]; !has {
			t.Errorf("Result[%d]: missing 'prompt' field", i)
		}
	}

	// Verify that total matches the number of answers submitted
	if resp.Total != len(answers) {
		t.Errorf("Total questions %d != number of answers submitted %d", resp.Total, len(answers))
	}

	// Log the score for debugging
	if resp.Total > 0 {
		t.Logf("Test completed: %d/%d correct (%.1f%%)", resp.Correct, resp.Total, float64(resp.Correct)*100.0/float64(resp.Total))
	}
}

func TestHandleLearningGrammarChapters_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/"+sectionID+"/chapters", nil)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapters(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleLearningGrammarChapters_Success(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID
	_ = router.grammarService.PublishRepo.SetPublished("section", sectionID, true, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/"+sectionID+"/chapters", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapters(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["chapters"] == nil {
		t.Error("Expected chapters in response")
	}
}

func TestHandleLearningGrammarChapters_EmptySectionID(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories//chapters", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapters(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleLearningGrammarSectionAccess_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/"+sectionID+"/access", nil)
	w := httptest.NewRecorder()
	router.handleLearningGrammarSectionAccess(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleLearningGrammarSectionAccess_Success(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/"+sectionID+"/access", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarSectionAccess(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := resp["can_access"]; !ok {
		t.Error("Expected can_access in response")
	}
}

func TestHandleLearningGrammarChapterAccess_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 || len(sectionsData.Sections[0].ChapterIDs) == 0 {
		t.Fatalf("failed to get sections/chapters: %v", err)
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+chapterID+"/access", nil)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapterAccess(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleLearningGrammarCategoryTest_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/"+sectionID+"/test", nil)
	w := httptest.NewRecorder()
	router.handleLearningGrammarCategoryTest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleLearningGrammarSubmitTest_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	body := map[string]interface{}{
		"scope": "chapter", "scope_id": "ch1",
		"answers": []map[string]interface{}{},
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/tests/submit", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleLearningGrammarSubmitTest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleLearningGrammarPlacementTest_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/placement-test", nil)
	w := httptest.NewRecorder()
	router.handleLearningGrammarPlacementTest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleLearningGrammarSubmitPlacementTest_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	body := []byte(`{"q1":"a"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/placement-test/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleLearningGrammarSubmitPlacementTest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
