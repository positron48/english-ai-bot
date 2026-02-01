package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleLearningGrammarCategories_OK(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	_ = router.grammarService.PublishRepo.SetPublished("section", sectionsData.Sections[0].SectionID, true, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleLearningGrammarCategories_Unauthorized(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories", nil)
	w := httptest.NewRecorder()
	router.handleLearningGrammarCategories(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleLearningGrammarChapters_OK(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	section := sectionsData.Sections[0]
	_ = router.grammarService.PublishRepo.SetPublished("section", section.SectionID, true, nil)
	_ = router.grammarService.PublishRepo.SetPublished("chapter", section.ChapterIDs[0], true, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/"+section.SectionID+"/chapters", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapters(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleLearningGrammarChapters_UnpublishedSection_404(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	section := sectionsData.Sections[0]
	// Do not publish section

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/"+section.SectionID+"/chapters", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapters(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unpublished section, got %d", w.Code)
	}
}

func TestHandleLearningGrammarSectionAccess(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	section := sectionsData.Sections[0]

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories/"+section.SectionID+"/access", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarSectionAccess(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleLearningGrammarChapterAccess(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("failed to get sections")
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]
	_ = router.grammarService.PublishRepo.SetPublished("chapter", chapterID, true, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+chapterID+"/access", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLearningGrammarChapterAccess(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
