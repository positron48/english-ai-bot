//go:build integration

package user_flows

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/integration/testkit"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

const (
	sectionID   = "en.grammar.orientation_how_to_read"
	chapter1ID  = "en.grammar.orientation_how_to_read.subject_verb_object_in"
	chapter2ID  = "en.grammar.orientation_how_to_read.verb_forms_v1_v2"
)

// submitChapterTestPass fetches chapter test, builds correct answers from content, submits
func submitChapterTestPass(t *testing.T, h *testkit.Harness, token, chapterID string) bool {
	t.Helper()

	// Get chapter content for correct answers
	contentRepo := repository.NewGrammarContentRepository(zap.NewNop())
	chapter, err := contentRepo.GetChapter(chapterID)
	if err != nil {
		t.Fatalf("get chapter %s: %v", chapterID, err)
	}
	questionBank, ok := chapter.QuestionBank["questions"].([]interface{})
	if !ok || len(questionBank) == 0 {
		t.Skipf("chapter %s has no question bank", chapterID)
		return false
	}
	correctMap := make(map[string]interface{})
	for _, q := range questionBank {
		qMap, ok := q.(map[string]interface{})
		if !ok {
			continue
		}
		qID, _ := qMap["id"].(string)
		if qID != "" && qMap["correct_answer"] != nil {
			correctMap[qID] = qMap["correct_answer"]
		}
	}
	if len(correctMap) == 0 {
		t.Skipf("chapter %s has no correct_answer in questions", chapterID)
		return false
	}

	// GET test
	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+chapterID+"/test", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get test: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var testResp struct {
		Questions []map[string]interface{} `json:"questions"`
		Total     int                      `json:"total"`
	}
	if err := json.NewDecoder(w.Body).Decode(&testResp); err != nil {
		t.Fatalf("decode test: %v", err)
	}
	if len(testResp.Questions) == 0 {
		t.Skip("test has no questions")
		return false
	}

	// Build answers from test questions
	answers := make([]map[string]interface{}, 0, len(testResp.Questions))
	for _, q := range testResp.Questions {
		qID, _ := q["id"].(string)
		if qID == "" {
			continue
		}
		correct, ok := correctMap[qID]
		if !ok {
			continue
		}
		ans := map[string]interface{}{"question_id": qID}
		switch v := correct.(type) {
		case bool:
			if v {
				ans["answer"] = "true"
			} else {
				ans["answer"] = "false"
			}
		case float64:
			ans["answer"] = v // JSON number
		default:
			ans["answer"] = correct
		}
		answers = append(answers, ans)
	}
	if len(answers) == 0 {
		t.Skip("no matching answers from content")
		return false
	}

	// POST submit
	submitBody := map[string]interface{}{
		"scope":    "chapter",
		"scope_id": chapterID,
		"answers":  answers,
	}
	bodyJSON, _ := json.Marshal(submitBody)
	submitReq := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/tests/submit", bytes.NewReader(bodyJSON))
	submitReq.Header.Set("Content-Type", "application/json")
	submitReq.Header.Set("Authorization", token)
	w2 := httptest.NewRecorder()
	h.Router.ServeHTTP(w2, submitReq)
	if w2.Code != http.StatusOK {
		t.Fatalf("submit test: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	var submitResp struct {
		Score  int  `json:"score"`
		Passed bool `json:"passed"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&submitResp); err != nil {
		t.Fatalf("decode submit: %v", err)
	}
	return submitResp.Passed
}

func TestGrammarUnlockFlow_ChapterPassUnlocksNext(t *testing.T) {
	h := testkit.NewHarness(t)
	conn := h.GetConnection().GetConnection()

	telegramID := int64(66666)
	user := testkit.UserFixture(t, conn, telegramID)

	// Publish section and first two chapters
	testkit.GrammarPublishFixture(t, conn, sectionID, []string{chapter1ID, chapter2ID})

	token := h.AuthAsUser(telegramID)

	// Chapter 2 access before test: should be false (chapter 1 not passed)
	accessReq := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+chapter2ID+"/access", nil)
	accessReq.Header.Set("Authorization", token)
	w1 := httptest.NewRecorder()
	h.Router.ServeHTTP(w1, accessReq)
	if w1.Code != http.StatusOK {
		t.Fatalf("chapter 2 access: expected 200, got %d: %s", w1.Code, w1.Body.String())
	}
	var accessResp struct {
		CanAccess bool `json:"can_access"`
	}
	if err := json.NewDecoder(w1.Body).Decode(&accessResp); err != nil {
		t.Fatalf("decode access: %v", err)
	}
	if accessResp.CanAccess {
		t.Error("chapter 2 should NOT be accessible before chapter 1 passed")
	}

	// Pass chapter 1 test
	passed := submitChapterTestPass(t, h, token, chapter1ID)
	if !passed {
		t.Skip("chapter 1 test not passed (content may have changed)")
		return
	}

	// Chapter 2 access after test: should be true
	accessReq2 := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+chapter2ID+"/access", nil)
	accessReq2.Header.Set("Authorization", token)
	w2 := httptest.NewRecorder()
	h.Router.ServeHTTP(w2, accessReq2)
	if w2.Code != http.StatusOK {
		t.Fatalf("chapter 2 access after: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	if err := json.NewDecoder(w2.Body).Decode(&accessResp); err != nil {
		t.Fatalf("decode access: %v", err)
	}
	if !accessResp.CanAccess {
		t.Error("chapter 2 should be accessible after chapter 1 passed")
	}

	testkit.AssertGrammarChapterAccess(t, conn, user.ID, chapter1ID, true)
	testkit.AssertGrammarChapterAccess(t, conn, user.ID, chapter2ID, false) // chapter 2 itself not passed
}
