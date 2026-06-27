package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/repository"
)

func TestHandleLearningGrammarChapterOrTest_Routes(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 || len(sectionsData.Sections[0].ChapterIDs) == 0 {
		t.Fatalf("failed to get sections/chapters: %v", err)
	}
	section := sectionsData.Sections[0]
	chapterID := section.ChapterIDs[0]

	_ = router.grammarService.PublishRepo.SetPublished("section", section.SectionID, true, nil)
	_ = router.grammarService.PublishRepo.SetPublished("chapter", chapterID, true, nil)

	paths := []string{
		"/api/learning/grammar/chapters/" + chapterID + "/next",
		"/api/learning/grammar/chapters/" + chapterID + "/access",
		"/api/learning/grammar/chapters/" + chapterID + "/test",
		"/api/learning/grammar/chapters/" + chapterID,
	}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()

		router.handleLearningGrammarChapterOrTest(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("path %s: expected 200, got %d (%s)", path, w.Code, w.Body.String())
		}
	}
}

func TestHandleLearningGrammarNextChapter_SuccessAndErrors(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil {
		t.Fatalf("GetSections error: %v", err)
	}

	var sectionID string
	var firstChapter string
	var nextPublished string
	for _, section := range sectionsData.Sections {
		if len(section.ChapterIDs) >= 3 {
			sectionID = section.SectionID
			firstChapter = section.ChapterIDs[0]
			nextPublished = section.ChapterIDs[2]
			break
		}
	}
	if firstChapter == "" {
		t.Fatal("expected section with >= 3 chapters")
	}

	if err := router.grammarService.PublishRepo.SetPublished("chapter", nextPublished, true, nil); err != nil {
		t.Fatalf("SetPublished error: %v", err)
	}

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/chapters/"+firstChapter+"/next", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarNextChapter(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+firstChapter+"/next", nil)
		w := httptest.NewRecorder()
		router.handleLearningGrammarNextChapter(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("malformed path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters//next", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarNextChapter(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/missing.chapter/next", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarNextChapter(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+firstChapter+"/next", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarNextChapter(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if gotSection, _ := resp["section_id"].(string); gotSection != sectionID {
			t.Fatalf("expected section_id %s, got %v", sectionID, resp["section_id"])
		}
		if gotNext, _ := resp["next_chapter_id"].(string); gotNext != nextPublished {
			t.Fatalf("expected next_chapter_id %s, got %v", nextPublished, resp["next_chapter_id"])
		}
	})
}

func TestHandleLearningGrammarChapterAndChapterTest_SuccessAndBadRequest(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 || len(sectionsData.Sections[0].ChapterIDs) == 0 {
		t.Fatalf("failed to get sections/chapters: %v", err)
	}
	chapterID := sectionsData.Sections[0].ChapterIDs[0]
	if err := router.grammarService.PublishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("SetPublished error: %v", err)
	}

	t.Run("chapter bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters//", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarChapter(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("chapter success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+chapterID, nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarChapter(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if _, ok := resp["chapter"]; !ok {
			t.Fatal("expected chapter payload")
		}
	})

	t.Run("chapter test malformed path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters//test", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarChapterTest(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("chapter test success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/chapters/"+chapterID+"/test", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarChapterTest(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if _, ok := resp["questions"]; !ok {
			t.Fatal("expected questions payload")
		}
	})
}

func TestHandleLearningGrammarStatistics_SuccessAndGuards(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	settingsRepo := repository.NewAppSettingsRepository(router.db, router.logger)
	if err := settingsRepo.SetBoolSetting("hide_placement_test_button", true, 1); err != nil {
		t.Fatalf("SetBoolSetting error: %v", err)
	}

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/statistics", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarStatistics(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/statistics", nil)
		w := httptest.NewRecorder()
		router.handleLearningGrammarStatistics(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/statistics", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarStatistics(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if hidden, ok := resp["hide_placement_test_button"].(bool); !ok || !hidden {
			t.Fatalf("expected hide_placement_test_button=true, got %v", resp["hide_placement_test_button"])
		}
		if _, ok := resp["confirmed_level"]; !ok {
			t.Fatal("expected confirmed_level in response")
		}
	})
}

func TestHandleLearningGrammarContinueChapter_SuccessAndGuards(t *testing.T) {
	router, _, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/continue-chapter", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarContinueChapter(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/continue-chapter", nil)
		w := httptest.NewRecorder()
		router.handleLearningGrammarContinueChapter(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/continue-chapter", nil)
		req = setUserIDInContext(req, 1)
		w := httptest.NewRecorder()
		router.handleLearningGrammarContinueChapter(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if _, ok := resp["chapter"]; !ok {
			t.Fatal("expected chapter key in response")
		}
	})
}
