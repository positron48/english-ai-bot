package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupLumiFactsHandlerTest(t *testing.T) (*Router, *repository.LumiFactRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	if _, err := db.Exec(`DELETE FROM lumi_facts`); err != nil {
		t.Fatalf("clear lumi_facts: %v", err)
	}
	return &Router{db: db, logger: zap.NewNop()}, repository.NewLumiFactRepository(db)
}

func TestHandleAdminLumiFactsPostParagraphImportStillWorks(t *testing.T) {
	router, repo := setupLumiFactsHandlerTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/lumi-facts", strings.NewReader(`{
		"course_code": "es_ru",
		"context": "district",
		"locale": "ru",
		"text": "Первый факт.\n\nВторой факт."
	}`))
	w := httptest.NewRecorder()

	router.handleAdminLumiFacts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var res map[string]int
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["inserted"] != 2 {
		t.Fatalf("inserted = %d, want 2", res["inserted"])
	}
	facts, total, err := repo.List(req.Context(), repository.LumiFactFilter{CourseCode: "es_ru"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 || facts[0].Context != "general" || facts[1].Context != "general" {
		t.Fatalf("facts total=%d rows=%+v", total, facts)
	}
}

func TestHandleAdminLumiFactsPostArrayJSONImport(t *testing.T) {
	router, repo := setupLumiFactsHandlerTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/lumi-facts", strings.NewReader(`[
		{ "course_code": "es_ru", "context": "grammar", "locale": "ru", "body": "Факт грамматики." },
		{ "course_code": "", "context": "general", "body": "Общий факт.", "status": "archived" }
	]`))
	w := httptest.NewRecorder()

	router.handleAdminLumiFacts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	facts, total, err := repo.List(req.Context(), repository.LumiFactFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 {
		t.Fatalf("total = %d, want 2", total)
	}
	seenArchived := false
	for _, fact := range facts {
		if fact.Body == "Общий факт." {
			seenArchived = fact.Status == "archived" && fact.Locale == "ru"
		}
	}
	if !seenArchived {
		t.Fatalf("archived/default locale fact not found: %+v", facts)
	}
}

func TestHandleAdminLumiFactsPostWrappedJSONImport(t *testing.T) {
	router, repo := setupLumiFactsHandlerTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/lumi-facts", strings.NewReader(`{
		"facts": [
			{ "course_code": "en_ru", "context": "reading", "locale": "ru", "body": "Факт чтения." }
		]
	}`))
	w := httptest.NewRecorder()

	router.handleAdminLumiFacts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	_, total, err := repo.List(req.Context(), repository.LumiFactFilter{CourseCode: "en_ru", Context: "reading"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
}

func TestHandleAdminLumiFactsPostRejectsInvalidJSONImport(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: `{`},
		{name: "empty facts", body: `{"facts":[]}`},
		{name: "invalid context", body: `{"facts":[{"course_code":"es_ru","context":"district","locale":"ru","body":"bad"}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, repo := setupLumiFactsHandlerTest(t)
			req := httptest.NewRequest(http.MethodPost, "/api/admin/lumi-facts", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			router.handleAdminLumiFacts(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			_, total, err := repo.List(req.Context(), repository.LumiFactFilter{})
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if total != 0 {
				t.Fatalf("total = %d, want 0", total)
			}
		})
	}
}

func TestHandleAdminLumiFactsPromptTemplate(t *testing.T) {
	router, _ := setupLumiFactsHandlerTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/lumi-facts/prompt-template", nil)
	w := httptest.NewRecorder()

	router.handleAdminLumiFactsPromptTemplate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var res struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	for _, want := range []string{"\"facts\"", "course_code", "en_ru", "es_ru", "general", "grammar", "reading", "practice", "progress", "city"} {
		if !strings.Contains(res.Prompt, want) {
			t.Fatalf("prompt does not contain %q:\n%s", want, res.Prompt)
		}
	}
}
