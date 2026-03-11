package web

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
)

func makeBadGrammarRouter(t *testing.T) (*Router, int64, func()) {
	t.Helper()
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	logger := router.logger
	contentRepo := repository.NewGrammarContentRepository(logger)
	badDB := badDBConn(t)
	publishRepo := repository.NewGrammarPublishRepository(badDB, logger)
	attemptRepo := repository.NewGrammarAttemptRepository(badDB, logger)
	badGrammarService := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, logger)
	router.SetGrammarService(badGrammarService)
	return router, adminUserID, cleanup
}

// makeBadContentGrammarRouter creates a router where the content repo uses an empty FS
// so GetSections() fails (sections.json not found).
func makeBadContentGrammarRouter(t *testing.T) (*Router, int64, func()) {
	t.Helper()
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	logger := router.logger
	emptyFS := fstest.MapFS{}
	contentRepo := repository.NewGrammarContentRepositoryWithFS(emptyFS, logger)
	goodDB := router.db
	publishRepo := repository.NewGrammarPublishRepository(goodDB, logger)
	attemptRepo := repository.NewGrammarAttemptRepository(goodDB, logger)
	badContentService := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, logger)
	router.SetGrammarService(badContentService)
	return router, adminUserID, cleanup
}

// Ensure fs and fstest are used
var _ fs.FS = fstest.MapFS{}

// TestHandleAdminGrammarCategories_GetPublishedItemsError covers line 39-43
// (GetPublishedItemsByType error when publish repo uses bad DB).
func TestHandleAdminGrammarCategories_GetPublishedItemsError(t *testing.T) {
	router, _, cleanup := makeBadGrammarRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/categories", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategories(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarCategories_ChapterIsPublished covers line 69-71
// (chapterItem.IsPublished branch when a chapter is published).
func TestHandleAdminGrammarCategories_ChapterIsPublished(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	// Find a section with at least one existing chapter
	var chapterID string
	for _, section := range sectionsData.Sections {
		for _, cid := range section.ChapterIDs {
			if router.grammarService.ContentRepo.ChapterExists(cid) {
				chapterID = cid
				break
			}
		}
		if chapterID != "" {
			break
		}
	}
	if chapterID == "" {
		t.Skip("no existing chapter found")
	}

	// Publish the chapter so chapterItem.IsPublished == true
	if err := router.grammarService.PublishRepo.SetPublished("chapter", chapterID, true, &adminUserID); err != nil {
		t.Fatalf("SetPublished: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/categories", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarCategories_ItemNameNotNil covers lines 86-88
// (item.Name != nil branch when a section has a custom name).
func TestHandleAdminGrammarCategories_ItemNameNotNil(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID
	customName := "My Custom Section"

	if err := router.grammarService.PublishRepo.SetName("section", sectionID, &customName, &adminUserID); err != nil {
		t.Fatalf("SetName: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/categories", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Categories []struct {
			SectionID  string  `json:"section_id"`
			CustomName *string `json:"custom_name"`
		} `json:"categories"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, cat := range resp.Categories {
		if cat.SectionID == sectionID && cat.CustomName != nil && *cat.CustomName == customName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected custom_name %q for section %s in response", customName, sectionID)
	}
}

// TestHandleAdminGrammarCategoryPublish_SetPublishedError covers lines 147-151
// (SetPublished error when publish repo uses bad DB).
func TestHandleAdminGrammarCategoryPublish_SetPublishedError(t *testing.T) {
	router, adminUserID, cleanup := makeBadGrammarRouter(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID

	payload, _ := json.Marshal(map[string]interface{}{"is_published": true, "cascade": false})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/categories/"+sectionID+"/publish", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategoryPublish(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarChapters_MethodNotAllowed covers lines 195-198.
func TestHandleAdminGrammarChapters_MethodNotAllowed(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/chapters?section_id=some-id", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapters(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarChapters_SectionNotFound covers lines 221-224
// (section not found when section_id doesn't exist).
func TestHandleAdminGrammarChapters_SectionNotFound(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/chapters?section_id=nonexistent-section-id", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapters(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarChapters_GetPublishedItemsError covers lines 227-231
// (GetPublishedItemsByType error when publish repo uses bad DB).
func TestHandleAdminGrammarChapters_GetPublishedItemsError(t *testing.T) {
	router, _, cleanup := makeBadGrammarRouter(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/chapters?section_id="+sectionID, nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapters(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarChapters_ItemNameNotNil covers lines 262-264
// (item.Name != nil branch when a chapter has a custom name).
func TestHandleAdminGrammarChapters_ItemNameNotNil(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) == 0 {
		t.Skip("no chapters in first section")
	}
	chapterID := section.ChapterIDs[0]
	customName := "My Custom Chapter"

	if err := router.grammarService.PublishRepo.SetName("chapter", chapterID, &customName, &adminUserID); err != nil {
		t.Fatalf("SetName: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/chapters?section_id="+section.SectionID, nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapters(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Chapters []struct {
			ChapterID  string  `json:"chapter_id"`
			CustomName *string `json:"custom_name"`
		} `json:"chapters"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, ch := range resp.Chapters {
		if ch.ChapterID == chapterID && ch.CustomName != nil && *ch.CustomName == customName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected custom_name %q for chapter %s in response", customName, chapterID)
	}
}

// TestHandleAdminGrammarChapterPublish_SetPublishedError covers lines 321-325
// (SetPublished error when publish repo uses bad DB).
func TestHandleAdminGrammarChapterPublish_SetPublishedError(t *testing.T) {
	router, adminUserID, cleanup := makeBadGrammarRouter(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 || len(sectionsData.Sections[0].ChapterIDs) == 0 {
		t.Fatalf("failed to get sections/chapters: %v", err)
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]

	payload, _ := json.Marshal(map[string]interface{}{"is_published": true})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/chapters/"+chapterID+"/publish", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapterPublish(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarItemRename_SetNameError covers lines 387-391
// (SetName error when publish repo uses bad DB).
func TestHandleAdminGrammarItemRename_SetNameError(t *testing.T) {
	router, adminUserID, cleanup := makeBadGrammarRouter(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	sectionID := sectionsData.Sections[0].SectionID

	name := "Bad Name"
	payload, _ := json.Marshal(map[string]interface{}{"name": name})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/items/section/"+sectionID+"/rename", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarItemRename(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarCategories_GetSectionsError covers line 31-35
// (GetSections error when content repo uses empty FS).
func TestHandleAdminGrammarCategories_GetSectionsError(t *testing.T) {
	router, _, cleanup := makeBadContentGrammarRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/categories", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategories(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarChapters_GetSectionsError covers line 206-210
// (GetSections error when content repo uses empty FS).
func TestHandleAdminGrammarChapters_GetSectionsError(t *testing.T) {
	router, _, cleanup := makeBadContentGrammarRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/grammar/chapters?section_id=any-section", nil)
	w := httptest.NewRecorder()
	router.handleAdminGrammarChapters(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminGrammarCategoryPublish_CascadeBulkSetPublishedError covers line 164-167
// (BulkSetPublished error in cascade path).
// Uses sqlmock: SetPublished (Exec) succeeds, then BulkSetPublished (Begin) fails.
func TestHandleAdminGrammarCategoryPublish_CascadeBulkSetPublishedError(t *testing.T) {
	router, _, adminUserID, cleanup := setupGrammarTest(t)
	defer cleanup()

	// Get a real section with at least one existing chapter
	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections: %v", err)
	}
	var sectionID string
	for _, section := range sectionsData.Sections {
		for _, cid := range section.ChapterIDs {
			if router.grammarService.ContentRepo.ChapterExists(cid) {
				sectionID = section.SectionID
				break
			}
		}
		if sectionID != "" {
			break
		}
	}
	if sectionID == "" {
		t.Skip("no section with existing chapters found")
	}

	// Create sqlmock DB: SetPublished (Exec) succeeds, BulkSetPublished (Begin) fails
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// SetPublished uses Exec
	mock.ExpectExec("INSERT INTO grammar_published_items").WillReturnResult(sqlmock.NewResult(1, 1))
	// BulkSetPublished uses Begin → fail
	mock.ExpectBegin().WillReturnError(sqlmock.ErrCancelled)

	logger := router.logger
	contentRepo := repository.NewGrammarContentRepository(logger)
	publishRepo := repository.NewGrammarPublishRepository(db, logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db, logger)
	grammarService := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, logger)
	router.SetGrammarService(grammarService)

	payload, _ := json.Marshal(map[string]interface{}{"is_published": true, "cascade": true})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/grammar/categories/"+sectionID+"/publish", bytes.NewReader(payload))
	req = setUserIDInContext(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminGrammarCategoryPublish(w, req)

	// SetPublished succeeds → 200 (cascade error is just logged, not returned)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (cascade error is logged not returned), got %d: %s", w.Code, w.Body.String())
	}
}
