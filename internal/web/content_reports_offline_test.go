package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

func withUserID(req *http.Request, userID int64) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
}

func TestHandleLearningGrammarChapterReport(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()

	body := map[string]interface{}{
		"chapter_id":      "es.chapter.1",
		"theory_block_id": "b1",
		"report_category": "wrong_theory",
		"comment":         "wrong_theory",
	}
	raw, _ := json.Marshal(body)
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/learning/grammar/chapter/report", bytes.NewReader(raw)), userID)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapterReport(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleLearningGrammarTestReport(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()

	body := map[string]interface{}{
		"question_id":     "es.chapter.1::b1::q1",
		"chapter_id":      "es.chapter.1",
		"scope":           "chapter",
		"scope_id":        "es.chapter.1",
		"report_category": "wrong_answer",
		"comment":         "wrong_answer",
	}
	raw, _ := json.Marshal(body)
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/learning/grammar/test/report", bytes.NewReader(raw)), userID)
	w := httptest.NewRecorder()
	router.handleLearningGrammarTestReport(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleContentReportsOfflineSync_Idempotent(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()

	clientID := "offline-report-1"
	payload := map[string]interface{}{
		"reports": []map[string]interface{}{
			{
				"client_report_id":   clientID,
				"source_type":        "grammar_chapter",
				"report_category":    "typo",
				"comment":            "typo",
				"grammar_chapter_id": "es.chapter.1",
				"theory_block_id":    "b1",
				"payload": map[string]interface{}{
					"content_snapshot": map[string]interface{}{"chapter_id": "es.chapter.1"},
				},
			},
		},
	}
	raw, _ := json.Marshal(payload)

	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/content-reports/offline/sync-reports", bytes.NewReader(raw)), userID)
	w := httptest.NewRecorder()
	router.handleContentReportsOfflineSync(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first sync expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req2 := withUserID(httptest.NewRequest(http.MethodPost, "/api/content-reports/offline/sync-reports", bytes.NewReader(raw)), userID)
	w2 := httptest.NewRecorder()
	router.handleContentReportsOfflineSync(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second sync expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	repo := repository.NewContentReportRepository(router.db, router.logger)
	exists, err := repo.HasClientReport(userID, clientID)
	if err != nil {
		t.Fatalf("HasClientReport: %v", err)
	}
	if !exists {
		t.Fatal("expected client report to exist after sync")
	}
}

func TestIsValidReportCategory_GrammarChapterAndTest(t *testing.T) {
	if !models.IsValidReportCategory("grammar_chapter", "wrong_theory") {
		t.Fatal("expected grammar_chapter wrong_theory valid")
	}
	if !models.IsValidReportCategory("grammar_test", "ambiguous") {
		t.Fatal("expected grammar_test ambiguous valid")
	}
}
