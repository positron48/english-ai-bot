package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestInternalContentReports_UnifiedListAndSummary(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	router.internalServiceTokens = map[string]string{"default": "tok"}

	_, err := repository.NewContentReportRepository(router.db, router.logger).Create(repository.CreateContentReportInput{
		UserID:         userID,
		SourceType:     "word_training",
		Word:           "hello",
		ReportCategory: "bad_audio",
		CommentText:    "bad",
	})
	if err != nil {
		t.Fatalf("create word report: %v", err)
	}
	_ = seedInternalGrammarReport(t, router, userID, "en.chapter.9", "b9", "en.chapter.9::b9::q1")

	req := httptest.NewRequest(http.MethodGet, "/api/internal/content-reports?limit=50", nil)
	req.Header.Set("X-Service-Token", "tok")
	w := httptest.NewRecorder()
	router.handleInternalContentReports(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unified list: %d %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/internal/content-reports/summary", nil)
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalContentReportsSummary(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", w.Code, w.Body.String())
	}
}

func TestInternalContentReport_ReadingCurrentDocument(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	router.internalServiceTokens = map[string]string{"default": "tok"}
	repo := repository.NewContentReportRepository(router.db, router.logger)
	id, err := repo.Create(repository.CreateContentReportInput{UserID: userID, SourceType: "reading_text", GrammarChapterID: "free_es_a1_rain", Payload: map[string]interface{}{"text_id": "free_es_a1_rain", "content_snapshot": map[string]interface{}{"title": "Old title"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := router.db.Exec(`INSERT INTO reading_texts (text_id,category_id,title,level,target_language,reading_passage) VALUES ('free_es_a1_rain','es_a1','Corrected title','A1','es','{"segments":[]}')`); err != nil {
		t.Fatal(err)
	}
	path := "/api/internal/content-reports/" + strconv.FormatInt(id, 10)
	get := func() map[string]interface{} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-Service-Token", "tok")
		w := httptest.NewRecorder()
		router.handleInternalContentReportByID(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%d: %s", w.Code, w.Body.String())
		}
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	result := get()
	if result["reading_text_found"] != true || result["reading_text"].(map[string]interface{})["title"] != "Corrected title" {
		t.Fatalf("wrong current document: %+v", result)
	}
	if _, err := router.db.Exec(`DELETE FROM reading_texts WHERE text_id='free_es_a1_rain'`); err != nil {
		t.Fatal(err)
	}
	result = get()
	if result["reading_text_found"] != false || result["reading_text"] != nil {
		t.Fatalf("deleted text must remain reportable: %+v", result)
	}
}
