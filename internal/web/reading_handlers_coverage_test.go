package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

const readingCoverageTelegramID int64 = 900001

type readingCoverageWordService struct {
	isSingle    bool
	definition  string
	defErr      error
	onCourse    func(ctx context.Context, db *sql.DB, userID int64, word, courseCode string) error
}

func (m *readingCoverageWordService) IsSingleWord(text string) bool { return m.isSingle }

func (m *readingCoverageWordService) GetWordDefinition(ctx context.Context, userID int64, word string) (string, error) {
	if m.defErr != nil {
		return "", m.defErr
	}
	return m.definition, nil
}

func (m *readingCoverageWordService) GetWordDefinitionForCourse(ctx context.Context, userID int64, word, courseCode string) (string, error) {
	if m.onCourse != nil {
		if err := m.onCourse(ctx, nil, userID, word, courseCode); err != nil {
			return "", err
		}
	}
	if m.defErr != nil {
		return "", m.defErr
	}
	return m.definition, nil
}

func (m *readingCoverageWordService) ResolveNativeWordToTargetLemma(ctx context.Context, nativeWord, courseCode string) (string, error) {
	return "amar", nil
}

type readingCoverageWordServiceCloseDB struct {
	readingCoverageWordService
	db *sql.DB
}

func (m *readingCoverageWordServiceCloseDB) GetWordDefinitionForCourse(ctx context.Context, userID int64, word, courseCode string) (string, error) {
	_ = m.db.Close()
	return "", nil
}

func setupReadingCoverageDB(t *testing.T) (*Router, *sql.DB, int64) {
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
	user, err := userRepo.GetOrCreateUser(readingCoverageTelegramID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return router, db, user.ID
}

func seedReadingCoverageCatalog(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, q := range []string{
		`DELETE FROM reading_text_progress`,
		`DELETE FROM reading_texts`,
		`DELETE FROM reading_categories`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	catalogRepo := repository.NewReadingCatalogRepository(db)
	categories := []repository.ReadingCategoryUpsert{
		{
			CategoryID:        "en_daily",
			Title:             "",
			TitleTranslations: map[string]string{"ru": "Быт"},
			Level:             "A1",
			SortOrder:         1,
			TextIDs:           []string{"en-text-1", "en-text-2", "missing-text"},
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
			Level:              "A1",
			TargetLanguage:     "en",
			CoverThumbRelPath:  "covers/en-text-1.webp",
			CoverHeroRelPath:   "covers/en-text-1-hero.png",
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
}

func writeReadingCoverageBundle(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	readingDir := filepath.Join(dir, "reading", "texts")
	if err := os.MkdirAll(readingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	audioDir := filepath.Join(dir, "assets", "reading", "audio")
	imageDir := filepath.Join(dir, "assets", "reading", "images")
	for _, d := range []string{audioDir, imageDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(audioDir, "clip.mp3"), []byte("fake-mp3"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, ct := range map[string][]byte{
		"thumb.webp": []byte("webp"),
		"hero.png":   []byte("png"),
		"photo.jpg":  []byte("jpg"),
	} {
		if err := os.WriteFile(filepath.Join(imageDir, name), ct, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	textDoc := map[string]interface{}{
		"id":              "bundle-text-1",
		"category_id":     "en_bundle",
		"title":           "Bundle text",
		"level":           "A1",
		"target_language": "en",
		"reading_passage": map[string]interface{}{
			"segments": []interface{}{
				map[string]interface{}{
					"tokens": []interface{}{
						map[string]interface{}{"lemma": "alpha", "surface": "Alpha", "clickable": true},
						map[string]interface{}{"surface": "skip", "clickable": false},
						"not-a-map",
					},
				},
				"not-a-segment",
			},
		},
	}
	textBytes, err := json.Marshal(textDoc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(readingDir, "bundle-text-1.json"), textBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	idx := map[string]interface{}{
		"version":      "1.0",
		"generated_at": "now",
		"categories": map[string]interface{}{
			"en_bundle": map[string]interface{}{
				"id":       "en_bundle",
				"title":    "Bundle cat",
				"level":    "A1",
				"order":    1,
				"text_ids": []string{"bundle-text-1", "bad-text"},
			},
		},
		"texts": map[string]string{
			"bundle-text-1": "texts/bundle-text-1.json",
			"bad-text":      "../evil.json",
		},
	}
	idxBytes, err := json.Marshal(idx)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reading", "index.json"), idxBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func routerWithReadingBundle(t *testing.T, bundleDir string) (*Router, *sql.DB, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	for _, q := range []string{
		`DELETE FROM reading_text_progress`,
		`DELETE FROM reading_texts`,
		`DELETE FROM reading_categories`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{
		TargetLang:       "en",
		NativeLang:       "ru",
		GrammarBundleDir: bundleDir,
		ContentSource:    "bundle",
	}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(readingCoverageTelegramID + 1)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return router, db, user.ID
}

func insertReadingWordCard(t *testing.T, db *sql.DB, word, courseCode string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(
		`INSERT INTO word_cards (word, definition, course_code) VALUES (?, ?, ?) RETURNING id`,
		word, "definition", courseCode,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert word card: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, 0, ?, ?)`,
		id, word, "перевод", "meaning",
	)
	if err != nil {
		t.Fatalf("insert training card: %v", err)
	}
	return id
}

func TestReadingHandlersCoverage(t *testing.T) {
	t.Run("SyncReadingCatalogFromBundle", func(t *testing.T) {
		var nilRouter *Router
		if err := nilRouter.SyncReadingCatalogFromBundle(context.Background()); err != nil {
			t.Fatalf("nil router: %v", err)
		}
		router, _, _ := setupReadingCoverageDB(t)
		router.readingCatalogRepo = nil
		if err := router.SyncReadingCatalogFromBundle(context.Background()); err != nil {
			t.Fatalf("nil repo: %v", err)
		}
		bundleDir := writeReadingCoverageBundle(t)
		router, db, _ := routerWithReadingBundle(t, bundleDir)
		if err := router.SyncReadingCatalogFromBundle(context.Background()); err != nil {
			t.Fatalf("sync: %v", err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories`).Scan(&n); err != nil || n == 0 {
			t.Fatalf("categories after sync: n=%d err=%v", n, err)
		}
	})

	t.Run("getReadingWordService", func(t *testing.T) {
		var nilRouter *Router
		if nilRouter.getReadingWordService() != nil {
			t.Fatal("nil router should return nil service")
		}
		router, _, _ := setupReadingCoverageDB(t)
		if router.getReadingWordService() != nil {
			t.Fatal("nil wordService")
		}
		router.wordService = "not-a-service"
		if router.getReadingWordService() != nil {
			t.Fatal("wrong type")
		}
		mock := &readingCoverageWordService{isSingle: true}
		router.wordService = mock
		if router.getReadingWordService() == nil {
			t.Fatal("expected service")
		}
	})

	t.Run("findReadingWordCardByInput", func(t *testing.T) {
		router, db, _ := setupReadingCoverageDB(t)
		lemma, id, found, err := router.findReadingWordCardByInput("  ", "en_ru")
		if err != nil || found || lemma != "" || id != 0 {
			t.Fatalf("empty input: (%q,%d,%v,%v)", lemma, id, found, err)
		}
		cardID := insertReadingWordCard(t, db, "coffee", "en_ru")
		lemma, id, found, err = router.findReadingWordCardByInput("Coffee", "en_ru")
		if err != nil || !found || id != cardID {
			t.Fatalf("direct lemma: (%q,%d,%v,%v)", lemma, id, found, err)
		}
		_, err = db.Exec(`INSERT INTO word_forms (word_card_id, form) VALUES (?, ?)`, cardID, "coffees")
		if err != nil {
			t.Fatal(err)
		}
		lemma, id, found, err = router.findReadingWordCardByInput("coffees", "en_ru")
		if err != nil || !found || id != cardID {
			t.Fatalf("form mapping: (%q,%d,%v,%v)", lemma, id, found, err)
		}
		strayID := insertReadingWordCard(t, db, "coffees", "en_ru")
		lemma, id, found, err = router.findReadingWordCardByInput("coffees", "en_ru")
		if err != nil || !found || id != cardID || id == strayID {
			t.Fatalf("form mapping should beat stray lemma card: (%q,%d,%v,%v) stray=%d", lemma, id, found, err, strayID)
		}
		_, _, found, err = router.findReadingWordCardByInput("missingword", "en_ru")
		if err != nil || found {
			t.Fatalf("not found with course: found=%v err=%v", found, err)
		}
		enNevadaID := insertReadingWordCard(t, db, "nevada", "en_ru")
		esNevadaID := insertReadingWordCard(t, db, "nevada", "es_ru")
		lemma, id, found, err = router.findReadingWordCardByInput("nevada", "es_ru")
		if err != nil || !found || id != esNevadaID {
			t.Fatalf("es_ru homograph: (%q,%d,%v,%v) want es card %d", lemma, id, found, err, esNevadaID)
		}
		lemma, id, found, err = router.findReadingWordCardByInput("nevada", "en_ru")
		if err != nil || !found || id != enNevadaID {
			t.Fatalf("en_ru homograph: (%q,%d,%v,%v) want en card %d", lemma, id, found, err, enNevadaID)
		}
		globalID := insertReadingWordCard(t, db, "travel", "")
		_, err = db.Exec(`INSERT INTO word_forms (word_card_id, form) VALUES (?, ?)`, globalID, "travels")
		if err != nil {
			t.Fatal(err)
		}
		lemma, id, found, err = router.findReadingWordCardByInput("travels", "")
		if err != nil || !found || id != globalID {
			t.Fatalf("global form: (%q,%d,%v,%v)", lemma, id, found, err)
		}
		lemma, id, found, err = router.findReadingWordCardByInput("travel", "")
		if err != nil || !found || id != globalID {
			t.Fatalf("global lemma: (%q,%d,%v,%v)", lemma, id, found, err)
		}
		_, _, found, err = router.findReadingWordCardByInput("absentlemma", "")
		if err != nil || found {
			t.Fatalf("global absent: found=%v err=%v", found, err)
		}
		broken := NewRouter(zap.NewNop(), &config.Config{}, newBrokenDB(t), nil, nil, nil, nil)
		_, _, _, err = broken.findReadingWordCardByInput("coffee", "en_ru")
		if err == nil {
			t.Fatal("expected db error")
		}
	})

	t.Run("readingBootstrapTokens", func(t *testing.T) {
		if got := readingBootstrapTokens(nil); len(got) != 0 {
			t.Fatalf("nil doc: %v", got)
		}
		if got := readingBootstrapTokens(&readingTextDoc{ReadingPassage: nil}); len(got) != 0 {
			t.Fatalf("nil passage: %v", got)
		}
		if got := readingBootstrapTokens(&readingTextDoc{ReadingPassage: map[string]interface{}{}}); len(got) != 0 {
			t.Fatalf("no segments: %v", got)
		}
		doc := &readingTextDoc{
			ReadingPassage: map[string]interface{}{
				"segments": []interface{}{
					map[string]interface{}{
						"tokens": []interface{}{
							map[string]interface{}{"lemma": "  beta  ", "surface": "Beta", "clickable": true},
							map[string]interface{}{"lemma": "  ", "surface": "  ", "clickable": true},
							map[string]interface{}{"surface": "only", "clickable": true},
						},
					},
				},
			},
		}
		got := readingBootstrapTokens(doc)
		if len(got) != 3 {
			t.Fatalf("tokens = %v", got)
		}
	})

	t.Run("readReadingIndexFromBundleFS", func(t *testing.T) {
		emptyDir := t.TempDir()
		router, _, _ := routerWithReadingBundle(t, emptyDir)
		idx, err := router.readReadingIndexFromBundleFS()
		if err != nil || len(idx.Categories) != 0 {
			t.Fatalf("missing index: %#v err=%v", idx, err)
		}
		bundleDir := writeReadingCoverageBundle(t)
		router, _, _ = routerWithReadingBundle(t, bundleDir)
		idx, err = router.readReadingIndexFromBundleFS()
		if err != nil || len(idx.Categories) != 1 {
			t.Fatalf("valid index: %#v err=%v", idx, err)
		}
		badDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(badDir, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(badDir, "reading", "index.json"), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		router, _, _ = routerWithReadingBundle(t, badDir)
		if _, err := router.readReadingIndexFromBundleFS(); err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("readReadingText and readReadingIndex bundle paths", func(t *testing.T) {
		bundleDir := writeReadingCoverageBundle(t)
		router, _, _ := routerWithReadingBundle(t, bundleDir)
		router.readingCatalogRepo = nil
		idx, err := router.readReadingIndex()
		if err != nil || len(idx.Categories) == 0 {
			t.Fatalf("index: %#v err=%v", idx, err)
		}
		doc, err := router.readReadingText(idx, "bundle-text-1")
		if err != nil || doc == nil || doc.Title != "Bundle text" {
			t.Fatalf("bundle text: %#v err=%v", doc, err)
		}
		if _, err := router.readReadingText(idx, "missing"); err == nil {
			t.Fatal("expected missing text error")
		}
		if _, err := router.readReadingText(idx, "bad-text"); err == nil {
			t.Fatal("expected invalid path error")
		}
		if _, err := router.readReadingText(nil, "bundle-text-1"); err == nil {
			t.Fatal("expected nil index error")
		}
		badDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(badDir, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(badDir, "reading", "index.json"), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		broken := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{GrammarBundleDir: badDir}}, newBrokenDB(t), nil, nil, nil, nil)
		if _, err := broken.readReadingIndex(); err == nil {
			t.Fatal("expected index error when db and bundle fail")
		}
	})

	t.Run("BootstrapReadingWordCards", func(t *testing.T) {
		var nilRouter *Router
		nilRouter.BootstrapReadingWordCards(context.Background())

		router, db, _ := setupReadingCoverageDB(t)
		router.BootstrapReadingWordCards(context.Background())

		if _, err := db.Exec(`UPDATE courses SET status = 'archived' WHERE code = 'es_ru'`); err != nil {
			t.Fatal(err)
		}
		bundleDir := writeReadingCoverageBundle(t)
		router, db, _ = routerWithReadingBundle(t, bundleDir)
		router.readingCatalogRepo = nil
		mock := &readingCoverageWordService{
			isSingle:   true,
			definition: "def",
			onCourse: func(ctx context.Context, _ *sql.DB, userID int64, word, courseCode string) error {
				_, err := db.Exec(`INSERT INTO word_cards (word, definition) VALUES (?, ?)`, word, "def")
				return err
			},
		}
		router.wordService = mock
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		router.BootstrapReadingWordCards(ctx)

		router.BootstrapReadingWordCards(context.Background())
	})

	t.Run("handleLearningReadingCategories", func(t *testing.T) {
		router, db, userID := setupReadingCoverageDB(t)
		seedReadingCoverageCatalog(t, db)

		req := httptest.NewRequest(http.MethodPost, "/api/learning/reading/categories", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategories(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method: %d", rr.Code)
		}

		badDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(badDir, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(badDir, "reading", "index.json"), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		broken := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{GrammarBundleDir: badDir}}, newBrokenDB(t), nil, nil, nil, nil)
		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		broken.handleLearningReadingCategories(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("index error: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingCategories(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("ok: %d %s", rr.Code, rr.Body.String())
		}
		var resp struct {
			Categories []struct {
				CategoryID string `json:"category_id"`
				Title      string `json:"title"`
				Order      int    `json:"order"`
			} `json:"categories"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Categories) != 2 {
			t.Fatalf("categories = %+v", resp.Categories)
		}
		if resp.Categories[0].CategoryID != "en_daily" || resp.Categories[0].Title != "en_daily" {
			t.Fatalf("empty title fallback / sort: %+v", resp.Categories[0])
		}
	})

	t.Run("handleLearningReadingCategoryTexts", func(t *testing.T) {
		router, db, userID := setupReadingCoverageDB(t)
		seedReadingCoverageCatalog(t, db)

		req := httptest.NewRequest(http.MethodPost, "/api/learning/reading/categories/en_daily/texts", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories//texts", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("empty category: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories/en_daily/texts?page=abc&per_page=0", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("pagination defaults: %d", rr.Code)
		}

		progressRepo := repository.NewReadingTextProgressRepository(db)
		if err := progressRepo.MarkRead(userID, "en-text-1"); err != nil {
			t.Fatal(err)
		}
		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories/en_daily/texts?archive=true&page=5&per_page=1", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("archive page: %d", rr.Code)
		}
	})

	t.Run("handleLearningReadingTexts dispatch", func(t *testing.T) {
		router, db, userID := setupReadingCoverageDB(t)
		seedReadingCoverageCatalog(t, db)

		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/texts/en-text-1", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET: %d %s", rr.Code, rr.Body.String())
		}

		req = httptest.NewRequest(http.MethodPost, "/api/learning/reading/texts/en-text-2/mark-read", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("POST: %d %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("handleLearningReadingTextGet branches", func(t *testing.T) {
		router, db, userID := setupReadingCoverageDB(t)
		seedReadingCoverageCatalog(t, db)

		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/texts/", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTextGet(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("bad path: %d", rr.Code)
		}

		_, err := db.Exec(`UPDATE reading_texts SET reading_passage = 'null' WHERE text_id = 'en-text-2'`)
		if err != nil {
			t.Fatal(err)
		}
		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/texts/en-text-2", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingTextGet(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("nil passage: %d", rr.Code)
		}
	})

	t.Run("handleLearningReadingTextMarkRead branches", func(t *testing.T) {
		router, db, userID := setupReadingCoverageDB(t)
		seedReadingCoverageCatalog(t, db)

		req := httptest.NewRequest(http.MethodPost, "/api/learning/reading/texts/en-text-1/mark-read", nil)
		rr := httptest.NewRecorder()
		router.handleLearningReadingTextMarkRead(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("unauthorized: %d", rr.Code)
		}

		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingTextMarkRead(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("mark read warn linglow: %d %s", rr.Code, rr.Body.String())
		}

		courseRepo := repository.NewCourseRepository(db, zap.NewNop())
		if _, err := courseRepo.SelectCurrentCourse(context.Background(), userID, "en_ru"); err != nil {
			t.Fatal(err)
		}
		req = httptest.NewRequest(http.MethodPost, "/api/learning/reading/texts/en-text-3/mark-read", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingTextMarkRead(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("mark read with linglow: %d", rr.Code)
		}

		broken := NewRouter(zap.NewNop(), router.config, newBrokenDB(t), nil, nil, nil, nil)
		req = httptest.NewRequest(http.MethodPost, "/api/learning/reading/texts/en-text-1/mark-read", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		broken.handleLearningReadingTextMarkRead(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("mark read error: %d", rr.Code)
		}
	})

	t.Run("handleReadingWordLookup", func(t *testing.T) {
		router, db, userID := setupReadingCoverageDB(t)

		req := httptest.NewRequest(http.MethodPost, "/api/reading/word-lookup?lemma=hi", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup", nil)
		rr = httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("unauth: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=%20", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("lemma required: %d", rr.Code)
		}

		broken := NewRouter(zap.NewNop(), router.config, newBrokenDB(t), nil, nil, nil, nil)
		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=fail", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		broken.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("lookup db error: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=missing", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("not found: %d", rr.Code)
		}

		insertReadingWordCard(t, db, "apple", "en_ru")
		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=apple", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("found word: %d %s", rr.Code, rr.Body.String())
		}

		router.wordService = &readingCoverageWordService{
			isSingle:   true,
			definition: "created",
			onCourse: func(ctx context.Context, _ *sql.DB, userID int64, word, courseCode string) error {
				insertReadingWordCard(t, db, word, courseCode)
				return nil
			},
		}
		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=banana", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("fallback create: %d %s", rr.Code, rr.Body.String())
		}

		router2, db2, userID2 := setupReadingCoverageDB(t)
		dedicatedDB, err := sql.Open("postgres_compat", testutil.GetTestDSN(t))
		if err != nil {
			t.Fatalf("open dedicated db: %v", err)
		}
		t.Cleanup(func() { _ = dedicatedDB.Close() })
		router2.db = dedicatedDB
		router2.wordService = &readingCoverageWordServiceCloseDB{
			readingCoverageWordService: readingCoverageWordService{isSingle: true},
			db:                         dedicatedDB,
		}
		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=cherry", nil)
		req = setUserIDInContext(req, userID2)
		rr = httptest.NewRecorder()
		router2.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("fallback second lookup error: %d", rr.Code)
		}
		_ = db2

		router3, db3, userID3 := setupReadingCoverageDB(t)
		insertReadingWordCard(t, db3, "durian", "en_ru")
		if _, err := db3.Exec(`ALTER TABLE training_cards RENAME TO training_cards_cov_backup`); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_, _ = db3.Exec(`ALTER TABLE training_cards_cov_backup RENAME TO training_cards`)
		})
		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=durian", nil)
		req = setUserIDInContext(req, userID3)
		rr = httptest.NewRecorder()
		router3.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("ensure user cards error: %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("handleLearningReadingAudio", func(t *testing.T) {
		bundleDir := writeReadingCoverageBundle(t)
		router, _, _ := routerWithReadingBundle(t, bundleDir)

		req := httptest.NewRequest(http.MethodPost, "/api/learning/reading/audio?path=x", nil)
		rr := httptest.NewRecorder()
		router.handleLearningReadingAudio(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/audio", nil)
		rr = httptest.NewRecorder()
		router.handleLearningReadingAudio(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("path required: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/audio?path=../secret.mp3", nil)
		rr = httptest.NewRecorder()
		router.handleLearningReadingAudio(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("invalid path: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/audio?path=assets/reading/audio/missing.mp3", nil)
		rr = httptest.NewRecorder()
		router.handleLearningReadingAudio(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("not found: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/audio?path=assets/reading/free_en_a1_reading_1778500398/seg_04_narrator.mp3&course_code=en_ru", nil)
		rr = httptest.NewRecorder()
		router.handleLearningReadingAudio(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("embedded audio with course_code: %d", rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "audio/mpeg" {
			t.Fatalf("content-type = %q", ct)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/audio?path=assets/reading/audio/clip.mp3", nil)
		rr = httptest.NewRecorder()
		router.handleLearningReadingAudio(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("bundle audio: %d", rr.Code)
		}
	})

	t.Run("handleLearningReadingImage", func(t *testing.T) {
		bundleDir := writeReadingCoverageBundle(t)
		router, _, _ := routerWithReadingBundle(t, bundleDir)

		req := httptest.NewRequest(http.MethodPost, "/api/learning/reading/image?path=x", nil)
		rr := httptest.NewRecorder()
		router.handleLearningReadingImage(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/image?path=/abs.png", nil)
		rr = httptest.NewRecorder()
		router.handleLearningReadingImage(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("invalid path: %d", rr.Code)
		}

		tests := []struct {
			file string
			ct   string
		}{
			{"assets/reading/images/thumb.webp", "image/webp"},
			{"assets/reading/images/hero.png", "image/png"},
			{"assets/reading/images/photo.jpg", "image/jpeg"},
			{"assets/reading/images/raw.bin", "application/octet-stream"},
		}
		for _, tc := range tests {
			if tc.file == "assets/reading/images/raw.bin" {
				if err := os.WriteFile(filepath.Join(bundleDir, "assets", "reading", "images", "raw.bin"), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/image?path="+tc.file, nil)
			rr = httptest.NewRecorder()
			router.handleLearningReadingImage(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("%s: %d", tc.file, rr.Code)
			}
			if got := rr.Header().Get("Content-Type"); got != tc.ct {
				t.Fatalf("%s content-type = %q", tc.file, got)
			}
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/image?path=assets/reading/images/missing.gif&course_code=en_ru", nil)
		rr = httptest.NewRecorder()
		router.handleLearningReadingImage(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("not found: %d", rr.Code)
		}
	})

	t.Run("remaining branches", func(t *testing.T) {
		router, db, userID := setupReadingCoverageDB(t)
		seedReadingCoverageCatalog(t, db)
		if _, err := db.Exec(`
			INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids)
			VALUES ('es_filtered', 'Spanish filtered', 'A1', 0, '[]')
		`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`
			INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
			VALUES ('en-es-mismatch', 'en_daily', 'Mismatch', 'A1', 'es', '{"segments":[]}')
		`); err != nil {
			t.Fatal(err)
		}
		_, err := db.Exec(`UPDATE reading_categories SET text_ids = '["en-text-1","en-text-2","missing-text","en-es-mismatch"]' WHERE category_id = 'en_daily'`)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories", nil)
		req = setUserIDInContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleLearningReadingCategories(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("categories: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories/en_daily/texts", nil)
		rr = httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("texts unauthorized: %d", rr.Code)
		}

		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("texts ok: %d", rr.Code)
		}

		if _, err := db.Exec(`ALTER TABLE reading_text_progress RENAME TO reading_text_progress_cov_backup`); err != nil {
			t.Fatal(err)
		}
		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/categories/en_daily/texts", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleLearningReadingCategoryTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("texts with progress error: %d", rr.Code)
		}
		if _, err := db.Exec(`ALTER TABLE reading_text_progress_cov_backup RENAME TO reading_text_progress`); err != nil {
			t.Fatal(err)
		}

		badDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(badDir, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(badDir, "reading", "index.json"), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		broken := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{GrammarBundleDir: badDir}}, newBrokenDB(t), nil, nil, nil, nil)
		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/texts/en-text-1", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		broken.handleLearningReadingTextGet(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("text get index error: %d", rr.Code)
		}

		insertReadingWordCard(t, db, "knownword", "en_ru")
		var knownCardID int64
		if err := db.QueryRow(`SELECT id FROM word_cards WHERE word = 'knownword'`).Scan(&knownCardID); err != nil {
			t.Fatal(err)
		}
		knowledgeRepo := repository.NewUserWordKnowledgeRepository(db, zap.NewNop())
		if err := knowledgeRepo.MarkKnown(userID, knownCardID); err != nil {
			t.Fatal(err)
		}
		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=knownword", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("known word lookup: %d %s", rr.Code, rr.Body.String())
		}

		router.wordService = &readingCoverageWordService{isSingle: false}
		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=not-a-real-word", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("non-single fallback skip: %d", rr.Code)
		}

		router.wordService = &readingCoverageWordService{isSingle: true, defErr: fmt.Errorf("lookup failed")}
		req = httptest.NewRequest(http.MethodGet, "/api/reading/word-lookup?lemma=grape", nil)
		req = setUserIDInContext(req, userID)
		rr = httptest.NewRecorder()
		router.handleReadingWordLookup(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("fallback service error: %d", rr.Code)
		}

		if _, err := db.Exec(`UPDATE courses SET status = 'archived' WHERE code = 'es_ru'`); err != nil {
			t.Fatal(err)
		}
		bundleDir := writeReadingCoverageBundle(t)
		bootstrapRouter, db2, _ := routerWithReadingBundle(t, bundleDir)
		bootstrapRouter.readingCatalogRepo = nil
		insertReadingWordCard(t, db2, "alpha", "")
		bootstrapRouter.wordService = &readingCoverageWordService{isSingle: false}
		bootstrapRouter.BootstrapReadingWordCards(context.Background())

		badIndexDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(badIndexDir, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(badIndexDir, "reading", "index.json"), 0o755); err != nil {
			t.Fatal(err)
		}
		idxRouter, _, _ := routerWithReadingBundle(t, badIndexDir)
		if _, err := idxRouter.readReadingIndexFromBundleFS(); err == nil {
			t.Fatal("expected bundle index read error")
		}

		dirAudioBundle := writeReadingCoverageBundle(t)
		if err := os.MkdirAll(filepath.Join(dirAudioBundle, "assets/reading/audio/isdir"), 0o755); err != nil {
			t.Fatal(err)
		}
		mediaRouter, _, _ := routerWithReadingBundle(t, dirAudioBundle)
		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/audio?path=assets/reading/audio/isdir", nil)
		rr = httptest.NewRecorder()
		mediaRouter.handleLearningReadingAudio(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("audio read error: %d", rr.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/image?path=assets/reading/images/isdir", nil)
		if err := os.MkdirAll(filepath.Join(dirAudioBundle, "assets/reading/images/isdir"), 0o755); err != nil {
			t.Fatal(err)
		}
		rr = httptest.NewRecorder()
		mediaRouter.handleLearningReadingImage(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("image read error: %d", rr.Code)
		}

		corruptBundle := t.TempDir()
		if err := os.MkdirAll(filepath.Join(corruptBundle, "reading", "texts"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(corruptBundle, "reading", "texts", "bad.json"), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		corruptIdx := `{"version":"1","categories":{"c1":{"id":"c1","text_ids":["bad-json"]}},"texts":{"bad-json":"texts/bad.json"}}`
		if err := os.WriteFile(filepath.Join(corruptBundle, "reading", "index.json"), []byte(corruptIdx), 0o644); err != nil {
			t.Fatal(err)
		}
		corruptRouter, _, _ := routerWithReadingBundle(t, corruptBundle)
		corruptRouter.readingCatalogRepo = nil
		cidx, err := corruptRouter.readReadingIndex()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := corruptRouter.readReadingText(cidx, "bad-json"); err == nil {
			t.Fatal("expected json parse error")
		}
		emptyPathIdx := &readingIndex{Texts: map[string]string{"empty": "   "}}
		if _, err := corruptRouter.readReadingText(emptyPathIdx, "empty"); err == nil {
			t.Fatal("expected empty rel path error")
		}
		if err := os.WriteFile(filepath.Join(dirAudioBundle, "assets", "reading", "images", "cover.jpeg"), []byte("jpeg"), 0o644); err != nil {
			t.Fatal(err)
		}
		req = httptest.NewRequest(http.MethodGet, "/api/learning/reading/image?path=assets/reading/images/cover.jpeg", nil)
		rr = httptest.NewRecorder()
		mediaRouter.handleLearningReadingImage(rr, req)
		if rr.Code != http.StatusOK || rr.Header().Get("Content-Type") != "image/jpeg" {
			t.Fatalf("jpeg ext: %d ct=%s", rr.Code, rr.Header().Get("Content-Type"))
		}

		if _, err := db.Exec(`UPDATE courses SET status = 'archived' WHERE code = 'es_ru'`); err != nil {
			t.Fatal(err)
		}
		badBootstrapDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(badBootstrapDir, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(badBootstrapDir, "reading", "index.json"), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		badBootstrapRouter, _, _ := routerWithReadingBundle(t, badBootstrapDir)
		badBootstrapRouter.readingCatalogRepo = nil
		badBootstrapRouter.BootstrapReadingWordCards(context.Background())

		emptyBundle := t.TempDir()
		if err := os.MkdirAll(filepath.Join(emptyBundle, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(emptyBundle, "reading", "index.json"), []byte(`{"version":"1","categories":{},"texts":{}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		emptyRouter, _, _ := routerWithReadingBundle(t, emptyBundle)
		emptyRouter.readingCatalogRepo = nil
		emptyRouter.BootstrapReadingWordCards(context.Background())

		missingTextBundle := t.TempDir()
		if err := os.MkdirAll(filepath.Join(missingTextBundle, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(missingTextBundle, "reading", "index.json"), []byte(`{"version":"1","categories":{"c":{"id":"c","text_ids":["ghost"]}},"texts":{"ghost":"texts/missing.json"}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		missingRouter, _, _ := routerWithReadingBundle(t, missingTextBundle)
		missingRouter.readingCatalogRepo = nil
		missingRouter.wordService = &readingCoverageWordService{isSingle: true, definition: "x"}
		missingRouter.BootstrapReadingWordCards(context.Background())

		nullFieldsBundle := t.TempDir()
		if err := os.MkdirAll(filepath.Join(nullFieldsBundle, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(nullFieldsBundle, "reading", "index.json"), []byte(`{"version":"1","categories":null,"texts":null}`), 0o644); err != nil {
			t.Fatal(err)
		}
		nullRouter, _, _ := routerWithReadingBundle(t, nullFieldsBundle)
		if idx, err := nullRouter.readReadingIndexFromBundleFS(); err != nil || len(idx.Categories) != 0 {
			t.Fatalf("null fields index: %#v err=%v", idx, err)
		}

		brokenFind := NewRouter(zap.NewNop(), router.config, newBrokenDB(t), nil, nil, nil, nil)
		if _, _, _, err := brokenFind.findReadingWordCardByInput("any", "en_ru"); err == nil {
			t.Fatal("expected course lookup db error")
		}
		if _, _, _, err := brokenFind.findReadingWordCardByInput("any", ""); err == nil {
			t.Fatal("expected global lookup db error")
		}

		tokDoc := &readingTextDoc{ReadingPassage: map[string]interface{}{
			"segments": []interface{}{
				map[string]interface{}{"tokens": "not-a-slice"},
			},
		}}
		if got := readingBootstrapTokens(tokDoc); len(got) != 0 {
			t.Fatalf("bad tokens slice: %v", got)
		}

		catalogRouter, db3, _ := setupReadingCoverageDB(t)
		seedReadingCoverageCatalog(t, db3)
		if _, err := db3.Exec(`ALTER TABLE reading_texts RENAME TO reading_texts_cov_backup`); err != nil {
			t.Fatal(err)
		}
		idx, err := catalogRouter.readReadingIndex()
		if err != nil || len(idx.Categories) == 0 {
			t.Fatalf("snapshot fallback to bundle: %#v err=%v", idx, err)
		}
		if _, err := db3.Exec(`ALTER TABLE reading_texts_cov_backup RENAME TO reading_texts`); err != nil {
			t.Fatal(err)
		}

		brokenTextRouter, db4, _ := setupReadingCoverageDB(t)
		seedReadingCoverageCatalog(t, db4)
		_ = brokenTextRouter.readingCatalogRepo
		brokenTextDB := newBrokenDB(t)
		brokenTextRouter.readingCatalogRepo = repository.NewReadingCatalogRepository(brokenTextDB)
		if _, err := brokenTextRouter.readReadingText(&readingIndex{Texts: map[string]string{"x": "y"}}, "en-text-1"); err == nil {
			t.Fatal("expected repo read error")
		}
	})
}
