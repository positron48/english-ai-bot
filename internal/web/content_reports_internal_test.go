package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/repository"
)

func seedInternalGrammarReport(t *testing.T, router *Router, userID int64, chapterID, theoryBlock, questionID string) int64 {
	t.Helper()
	repo := repository.NewContentReportRepository(router.db, router.logger)
	id, err := repo.Create(repository.CreateContentReportInput{
		UserID:            userID,
		SourceType:        "grammar_training",
		GrammarChapterID:  chapterID,
		TheoryBlockID:     theoryBlock,
		GrammarQuestionID: questionID,
	})
	if err != nil {
		t.Fatalf("seed report: %v", err)
	}
	return id
}

func TestInternalGrammarContentReports_UnauthorizedAndMethod(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	_ = userID
	router.internalServiceTokens = map[string]string{"default": "tok"}

	req := httptest.NewRequest(http.MethodPost, "/api/internal/content-reports/grammar", nil)
	w := httptest.NewRecorder()
	router.handleInternalGrammarContentReports(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/internal/content-reports/grammar", nil)
	w = httptest.NewRecorder()
	router.handleInternalGrammarContentReports(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestInternalGrammarContentReports_ListAndResolveBulk(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	router.internalServiceTokens = map[string]string{"default": "tok"}

	id1 := seedInternalGrammarReport(t, router, userID, "es.chapter.1", "b1", "es.chapter.1::b1::q1")
	id2 := seedInternalGrammarReport(t, router, userID, "en.chapter.1", "b2", "en.chapter.1::b2::q1")
	_ = seedInternalGrammarReport(t, router, userID, "es.chapter.2", "b3", "es.chapter.2::b3::q1")

	req := httptest.NewRequest(http.MethodGet, "/api/internal/content-reports/grammar?course=es&limit=2", nil)
	req.Header.Set("X-Service-Token", "tok")
	w := httptest.NewRecorder()
	router.handleInternalGrammarContentReports(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var listResp struct {
		Reports    []map[string]interface{} `json:"reports"`
		NextCursor int64                    `json:"next_cursor"`
	}
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Reports) != 2 {
		t.Fatalf("expected 2 reports for es with limit=2, got %d", len(listResp.Reports))
	}
	if listResp.NextCursor == 0 {
		t.Fatal("expected next_cursor")
	}

	// invalid cursor
	req = httptest.NewRequest(http.MethodGet, "/api/internal/content-reports/grammar?cursor=bad", nil)
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalGrammarContentReports(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad cursor, got %d", w.Code)
	}

	// resolve bulk
	body, _ := json.Marshal(map[string]interface{}{
		"report_ids": []int64{id1, id2},
		"reason":     "auto_cleanup_by_worker",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/internal/content-reports/grammar/resolve-bulk", bytes.NewReader(body))
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalGrammarContentReportsResolveBulk(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 resolve bulk, got %d: %s", w.Code, w.Body.String())
	}

	// verify those reports are no longer active by cursor/list endpoint
	req = httptest.NewRequest(http.MethodGet, "/api/internal/content-reports/grammar?chapter_id=en.chapter.1", nil)
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalGrammarContentReports(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 list after resolve, got %d", w.Code)
	}
	var after struct {
		Reports []map[string]interface{} `json:"reports"`
	}
	if err := json.NewDecoder(w.Body).Decode(&after); err != nil {
		t.Fatalf("decode after response: %v", err)
	}
	if len(after.Reports) != 0 {
		t.Fatalf("expected 0 active reports for en.chapter.1 after resolve, got %d", len(after.Reports))
	}
}
