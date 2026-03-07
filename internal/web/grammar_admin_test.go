package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAdminGrammarCategories(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/categories", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAdminGrammarCategoryPublish(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	sectionID := sectionsData.Sections[0].SectionID

	body := map[string]interface{}{"is_published": true, "cascade": false}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/categories/"+sectionID+"/publish", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategoryPublish(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAdminGrammarChapters(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	sectionID := sectionsData.Sections[0].SectionID

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/chapters?section_id="+sectionID, nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapters(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAdminGrammarChapterPublish(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]

	payload, _ := json.Marshal(map[string]interface{}{"is_published": true})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/chapters/"+chapterID+"/publish", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapterPublish(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAdminGrammarItemRename(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	sectionID := sectionsData.Sections[0].SectionID

	payload, _ := json.Marshal(map[string]interface{}{"name": "Custom"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/items/section/"+sectionID+"/rename", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarItemRename(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAdminGrammarCategoryPublish_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/categories/section/publish", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategoryPublish(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleAdminGrammarChapters_MissingSection(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/chapters", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapters(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAdminGrammarChapterPublish_InvalidBody(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/chapters/chapter/publish", bytes.NewBufferString("invalid"))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapterPublish(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAdminGrammarItemRename_InvalidType(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	payload, _ := json.Marshal(map[string]interface{}{"name": "bad"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/items/bad/id/rename", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarItemRename(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAdminGrammarCategories_MethodNotAllowed(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/categories", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategories(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminGrammarCategoryPublish_MethodNotAllowed(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/categories/section_id/publish", nil)
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategoryPublish(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminGrammarCategoryPublish_EmptySectionID(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	payload, _ := json.Marshal(map[string]interface{}{"is_published": true})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/categories//publish", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategoryPublish(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminGrammarCategoryPublish_InvalidBody(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	sectionID := sectionsData.Sections[0].SectionID

	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/categories/"+sectionID+"/publish", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategoryPublish(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAdminGrammarCategoryPublish_WithCascade(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	sectionID := sectionsData.Sections[0].SectionID

	body := map[string]interface{}{"is_published": true, "cascade": true}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/categories/"+sectionID+"/publish", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategoryPublish(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["success"] != true {
		t.Error("expected success true")
	}
}
