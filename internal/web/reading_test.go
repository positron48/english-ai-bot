package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupReadingHandlerTest(t *testing.T) (*Router, *sql.DB, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{
		TargetLang:      "en",
		NativeLang:      "ru",
		GrammarBundleID: "en",
		ContentSource:   "db",
	}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(88001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM reading_text_progress`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM reading_texts`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM reading_categories`); err != nil {
		t.Fatal(err)
	}

	catalogRepo := repository.NewReadingCatalogRepository(db)
	categories := []repository.ReadingCategoryUpsert{
		{
			CategoryID:        "en_daily",
			Title:             "Daily Life",
			TitleTranslations: map[string]string{"ru": "Быт"},
			Level:             "A1",
			SortOrder:         2,
			TextIDs:           []string{"en-text-1", "en-text-2"},
		},
		{
			CategoryID: "en_travel",
			Title:      "Travel",
			Level:      "A2",
			SortOrder:  1,
			TextIDs:    []string{"en-text-3"},
		},
	}
	texts := []repository.ReadingTextUpsert{
		{
			TextID:             "en-text-1",
			CategoryID:         "en_daily",
			Title:              "Morning coffee",
			TitleTranslations:  map[string]string{"ru": "Утренний кофе"},
			Level:              "A1",
			TargetLanguage:     "en",
			CoverThumbRelPath:  "covers/en-text-1.webp",
			ReadingPassageJSON: `{"title":"Morning coffee","segments":[{"tokens":[{"lemma":"coffee","clickable":true}]}]}`,
		},
		{
			TextID:             "en-text-2",
			CategoryID:         "en_daily",
			Title:              "At the shop",
			Level:              "A1",
			TargetLanguage:     "en",
			ReadingPassageJSON: `{"segments":[]}`,
		},
		{
			TextID:             "en-text-3",
			CategoryID:         "en_travel",
			Title:              "At the airport",
			Level:              "A2",
			TargetLanguage:     "en",
			ReadingPassageJSON: `{"segments":[]}`,
		},
	}
	if err := catalogRepo.ReplaceCatalogForTargetLanguage("en", "1.0", "t", categories, texts); err != nil {
		t.Fatal(err)
	}

	return router, db, user.ID
}

func TestReadingTextIDFromPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		path           string
		expectMarkRead bool
		wantID         string
		wantOK         bool
	}{
		{name: "plain text id", path: "/api/learning/reading/texts/en-text-1", expectMarkRead: false, wantID: "en-text-1", wantOK: true},
		{name: "mark read", path: "/api/learning/reading/texts/en-text-1/mark-read", expectMarkRead: true, wantID: "en-text-1", wantOK: true},
		{name: "missing suffix", path: "/api/learning/reading/texts/en-text-1/read", expectMarkRead: true, wantID: "", wantOK: false},
		{name: "nested path", path: "/api/learning/reading/texts/en-text-1/extra", expectMarkRead: false, wantID: "", wantOK: false},
		{name: "empty", path: "/api/learning/reading/texts/", expectMarkRead: false, wantID: "", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := readingTextIDFromPath(tt.path, tt.expectMarkRead)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Fatalf("readingTextIDFromPath() = (%q, %v), want (%q, %v)", gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

func TestSnapshotToReadingIndexAndDocFromRepo(t *testing.T) {
	t.Parallel()
	snap := &repository.ReadingCatalogSnapshot{
		Version:     "1.0",
		GeneratedAt: "now",
		Categories: map[string]*repository.ReadingCategorySnapshot{
			"en_daily": {
				CategoryID:        "en_daily",
				Title:             "Daily",
				TitleTranslations: map[string]string{"ru": "Быт"},
				Level:             "A1",
				Order:             1,
				TextIDs:           []string{"t1", "t2"},
			},
		},
	}
	idx := snapshotToReadingIndex(snap)
	if idx == nil || len(idx.Categories) != 1 || idx.Texts["t1"] != "" || idx.Texts["t2"] != "" {
		t.Fatalf("index = %#v", idx)
	}
	if idx.Categories["en_daily"].TitleTranslations["ru"] != "Быт" {
		t.Fatalf("category = %#v", idx.Categories["en_daily"])
	}

	if docFromRepo(nil) != nil {
		t.Fatal("docFromRepo(nil) should return nil")
	}
	repoDoc := &repository.ReadingTextDocument{
		ID:                "t1",
		CategoryID:        "en_daily",
		Title:             "Title",
		TitleTranslations: map[string]string{"ru": "Заголовок"},
		Level:             "A1",
		TargetLanguage:    "en",
		CoverThumbRelPath: "thumb.webp",
		ReadingPassage:    map[string]interface{}{"segments": []interface{}{}},
	}
	doc := docFromRepo(repoDoc)
	if doc == nil || doc.ID != "t1" || doc.CoverThumbRelPath != "thumb.webp" {
		t.Fatalf("doc = %#v", doc)
	}
}

func TestHandleLearningReadingCategories(t *testing.T) {
	router, db, userID := setupReadingHandlerTest(t)
	progressRepo := repository.NewReadingTextProgressRepository(db)
	if err := progressRepo.MarkRead(userID, "en-text-1"); err != nil {
		t.Fatal(err)
	}

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories", nil)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategories(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/learning/reading/categories", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategories(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("lists english categories with read counts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategories(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
		}
		var resp struct {
			Categories []struct {
				CategoryID string `json:"category_id"`
				Title      string `json:"title"`
				TextCount  int    `json:"text_count"`
				ReadCount  int    `json:"read_count"`
				Order      int    `json:"order"`
			} `json:"categories"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Categories) != 2 {
			t.Fatalf("categories = %+v", resp.Categories)
		}
		if resp.Categories[0].CategoryID != "en_travel" || resp.Categories[1].CategoryID != "en_daily" {
			t.Fatalf("sort order = %+v", resp.Categories)
		}
		if resp.Categories[1].ReadCount != 1 || resp.Categories[1].TextCount != 2 {
			t.Fatalf("daily category = %+v", resp.Categories[1])
		}
	})
}

func TestHandleLearningReadingCategoryTexts(t *testing.T) {
	router, db, userID := setupReadingHandlerTest(t)
	progressRepo := repository.NewReadingTextProgressRepository(db)
	if err := progressRepo.MarkRead(userID, "en-text-1"); err != nil {
		t.Fatal(err)
	}

	t.Run("invalid path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories/en_daily/details", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("category not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories/missing/texts", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("lists unread texts with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories/en_daily/texts?page=1&per_page=1", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
		}
		var resp struct {
			Category map[string]interface{} `json:"category"`
			Texts    []struct {
				TextID string `json:"text_id"`
				IsRead bool   `json:"is_read"`
			} `json:"texts"`
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp.Total != 1 || len(resp.Texts) != 1 || resp.Texts[0].TextID != "en-text-2" || resp.Texts[0].IsRead {
			t.Fatalf("resp = %+v", resp)
		}
		if resp.Category["category_id"] != "en_daily" {
			t.Fatalf("category = %#v", resp.Category)
		}
	})

	t.Run("archive lists read texts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories/en_daily/texts?archive=true", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
		}
		var resp struct {
			Texts []struct {
				TextID string `json:"text_id"`
				IsRead bool   `json:"is_read"`
			} `json:"texts"`
			Total int `json:"total"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp.Total != 1 || len(resp.Texts) != 1 || resp.Texts[0].TextID != "en-text-1" || !resp.Texts[0].IsRead {
			t.Fatalf("archive resp = %+v", resp)
		}
	})
}

func TestHandleLearningReadingTextGet(t *testing.T) {
	router, db, userID := setupReadingHandlerTest(t)

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/texts/en-text-1", nil)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTextGet(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/texts/missing-text", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTextGet(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("returns text and unread progress", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/texts/en-text-1", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTextGet(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
		}
		var resp struct {
			TextID          string                 `json:"text_id"`
			Title           string                 `json:"title"`
			Block           map[string]interface{} `json:"block"`
			ReadingProgress map[string]interface{} `json:"reading_progress"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp.TextID != "en-text-1" || resp.Title != "Morning coffee" {
			t.Fatalf("resp = %+v", resp)
		}
		if resp.Block["type"] != "reading_passage" {
			t.Fatalf("block = %#v", resp.Block)
		}
		if resp.ReadingProgress["is_read"] != false {
			t.Fatalf("progress = %#v", resp.ReadingProgress)
		}
	})

	t.Run("returns read progress after mark read", func(t *testing.T) {
		progressRepo := repository.NewReadingTextProgressRepository(db)
		if err := progressRepo.MarkRead(userID, "en-text-2"); err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/texts/en-text-2", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTextGet(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
		}
		var resp struct {
			ReadingProgress map[string]interface{} `json:"reading_progress"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp.ReadingProgress["is_read"] != true || resp.ReadingProgress["read_at"] == nil {
			t.Fatalf("progress = %#v", resp.ReadingProgress)
		}
	})
}

func TestHandleLearningReadingTextMarkRead(t *testing.T) {
	router, db, userID := setupReadingHandlerTest(t)

	t.Run("invalid path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/learning/reading/texts/en-text-1/read", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTextMarkRead(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("marks text as read", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/learning/reading/texts/en-text-1/mark-read", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTextMarkRead(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
		}
		var resp struct {
			Success bool   `json:"success"`
			TextID  string `json:"text_id"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if !resp.Success || resp.TextID != "en-text-1" {
			t.Fatalf("resp = %+v", resp)
		}

		progressRepo := repository.NewReadingTextProgressRepository(db)
		progress, err := progressRepo.Get(userID, "en-text-1")
		if err != nil || progress == nil {
			t.Fatalf("progress = %#v err=%v", progress, err)
		}
	})

	t.Run("method not allowed on texts handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/learning/reading/texts/en-text-1", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTexts(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d", rr.Code)
		}
	})
}
