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
	"strings"
	"sync"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

const adminReadingGapTelegramID int64 = 900001

var adminReadingGapDB struct {
	once sync.Once
	db   *database.DB
	dsn  string
	err  error
	ref  interface{}
}

func adminReadingGapSharedDB(t *testing.T) *sql.DB {
	t.Helper()
	adminReadingGapDB.once.Do(func() {
		ctx := context.Background()
		ctr, err := postgres.Run(ctx, "postgres:16-alpine",
			postgres.WithDatabase("english_admin_reading_gap"),
			postgres.WithUsername("english"),
			postgres.WithPassword("english"),
		)
		if err != nil {
			adminReadingGapDB.err = err
			return
		}
		dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			adminReadingGapDB.err = err
			return
		}
		logger := zap.NewNop()
		var db *database.DB
		for attempt := 0; attempt < 10; attempt++ {
			db, err = database.NewWithConfig("postgres", "", dsn, logger)
			if err == nil {
				break
			}
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
		}
		if err != nil {
			adminReadingGapDB.err = err
			return
		}
		adminReadingGapDB.db = db
		adminReadingGapDB.dsn = dsn
		adminReadingGapDB.ref = ctr
	})
	if adminReadingGapDB.err != nil {
		if strings.Contains(adminReadingGapDB.err.Error(), "Cannot connect") ||
			strings.Contains(adminReadingGapDB.err.Error(), "docker") {
			t.Skipf("postgres unavailable: %v", adminReadingGapDB.err)
		}
		t.Fatalf("shared postgres: %v", adminReadingGapDB.err)
	}
	conn := adminReadingGapDB.db.GetConnection()
	for _, q := range []string{
		`TRUNCATE reading_text_progress, reading_texts, reading_categories, learning_items RESTART IDENTITY CASCADE`,
		`INSERT INTO circuit_breaker_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING`,
	} {
		if _, err := conn.Exec(q); err != nil {
			t.Fatalf("reset: %v", err)
		}
	}
	return conn
}

func adminReadingGapRouter(t *testing.T, learning config.LearningConfig) (*Router, *sql.DB) {
	t.Helper()
	db := adminReadingGapSharedDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: learning}
	return NewRouter(logger, cfg, db, nil, nil, nil, nil), db
}

func adminReadingGapSeedCatalog(t *testing.T, db *sql.DB, texts ...repository.ReadingTextUpsert) {
	t.Helper()
	repo := repository.NewReadingCatalogRepository(db)
	categories := []repository.ReadingCategoryUpsert{{
		CategoryID: "gap_cat",
		Title:      "Gap",
		Level:      "A1",
		SortOrder:  1,
		TextIDs:    nil,
	}}
	for _, tx := range texts {
		categories[0].TextIDs = append(categories[0].TextIDs, tx.TextID)
	}
	if len(texts) == 0 {
		categories[0].TextIDs = []string{"orphan-id"}
	}
	if err := repo.ReplaceCatalogForTargetLanguage("en", "1", "t", categories, texts); err != nil {
		t.Fatal(err)
	}
}

func adminReadingGapWriteBundle(t *testing.T, textID, targetLang, relPath string) string {
	t.Helper()
	dir := t.TempDir()
	textsDir := filepath.Join(dir, "reading", "texts")
	if err := os.MkdirAll(textsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if relPath == "" {
		relPath = "texts/" + textID + ".json"
	}
	doc := map[string]interface{}{
		"id":              textID,
		"category_id":     "gap_cat",
		"title":           "Gap title",
		"level":           "A1",
		"target_language": targetLang,
		"reading_passage": map[string]interface{}{
			"segments": []interface{}{
				map[string]interface{}{"id": "s1"},
			},
		},
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	full := filepath.Join(dir, "reading", filepath.Clean(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	idx := map[string]interface{}{
		"version":      "1",
		"generated_at": "t",
		"categories": map[string]interface{}{
			"gap_cat": map[string]interface{}{
				"id":       "gap_cat",
				"text_ids": []string{textID},
			},
		},
		"texts": map[string]string{textID: relPath},
	}
	idxRaw, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reading", "index.json"), append(idxRaw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAdminReadingGap_ResolveAdminReadingCourse(t *testing.T) {
	router, _ := adminReadingGapRouter(t, config.LearningConfig{TargetLang: "en", NativeLang: "ru"})
	req := httptest.NewRequest(http.MethodGet, "/?course_code=en_ru", nil)
	code, lang, err := router.resolveAdminReadingCourse(req)
	if err != nil || code != "en_ru" || lang != "en" {
		t.Fatalf("got %q %q err=%v", code, lang, err)
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	code, lang, err = router.resolveAdminReadingCourse(req)
	if err != nil || code != "en_ru" || lang != "en" {
		t.Fatalf("default course got %q %q err=%v", code, lang, err)
	}
}

func TestAdminReadingGap_HandleAdminReadingTextsList_TargetLanguageFilter(t *testing.T) {
	router, db := adminReadingGapRouter(t, config.LearningConfig{
		TargetLang: "es", NativeLang: "ru", ContentSource: "db",
	})
	repo := repository.NewReadingCatalogRepository(db)
	if err := repo.ReplaceCatalogForTargetLanguage("es", "1", "t",
		[]repository.ReadingCategoryUpsert{{
			CategoryID: "es_cat", Title: "ES", Level: "A1", SortOrder: 1,
			TextIDs: []string{"es-one"},
		}},
		[]repository.ReadingTextUpsert{{
			TextID: "es-one", CategoryID: "es_cat", Title: "Hola", Level: "A1",
			TargetLanguage: "es", ReadingPassageJSON: `{}`,
		}},
	); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/reading/texts?course_code=es_ru", nil)
	rr := httptest.NewRecorder()
	router.handleAdminReadingTextsList(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var resp struct {
		Texts []adminReadingTextItem `json:"texts"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Texts) != 1 || resp.Texts[0].TargetLanguage != "es" {
		t.Fatalf("texts = %+v", resp.Texts)
	}
}

func TestAdminReadingGap_HandleAdminReadingTexts_RoutesAnd405(t *testing.T) {
	bundleDir := adminReadingGapWriteBundle(t, "route-text", "en", "")
	router, _ := adminReadingGapRouter(t, config.LearningConfig{
		TargetLang:       "en",
		NativeLang:       "ru",
		GrammarBundleID:  "en",
		GrammarBundleDir: bundleDir,
		ContentSource:    "db",
	})
	adminReadingGapSeedCatalog(t, router.db, repository.ReadingTextUpsert{
		TextID:             "route-text",
		CategoryID:         "gap_cat",
		Title:              "Route",
		Level:              "A1",
		TargetLanguage:     "en",
		ReadingPassageJSON: `{"segments":[{"id":"s1"}]}`,
	})

	t.Run("GET list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/reading/texts?course_code=en_ru&search=route", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET status = %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/route-text?course_code=en_ru", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTexts(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("DELETE status = %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/admin/reading/texts", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTexts(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d", rr.Code)
		}
	})
}

func TestAdminReadingGap_HandleAdminReadingTextsList_FiltersAndEdges(t *testing.T) {
	bundleDir := adminReadingGapWriteBundle(t, "alpha-text", "en", "")
	router, db := adminReadingGapRouter(t, config.LearningConfig{
		TargetLang:       "en",
		NativeLang:       "ru",
		GrammarBundleID:  "en",
		GrammarBundleDir: bundleDir,
		ContentSource:    "db",
	})
	repo := repository.NewReadingCatalogRepository(db)
	categories := []repository.ReadingCategoryUpsert{
		{CategoryID: "c1", Title: "C1", Level: "A1", SortOrder: 1, TextIDs: []string{"alpha-text", "beta-text", "orphan-id"}},
	}
	texts := []repository.ReadingTextUpsert{
		{TextID: "alpha-text", CategoryID: "c1", Title: "Alpha morning", Level: "A1", TargetLanguage: "en", ReadingPassageJSON: `{"segments":[{}]}`},
		{TextID: "beta-text", CategoryID: "c1", Title: "", Level: "A2", TargetLanguage: "en", ReadingPassageJSON: `{"segments":[]}`},
	}
	if err := repo.ReplaceCatalogForTargetLanguage("en", "1", "t", categories, texts); err != nil {
		t.Fatal(err)
	}

	t.Run("search and level filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/reading/texts?course_code=en_ru&search=alpha&level=A1", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTextsList(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
		}
		var resp struct {
			Texts []adminReadingTextItem `json:"texts"`
			Total int                    `json:"total"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Texts) != 1 || resp.Texts[0].TextID != "alpha-text" || resp.Texts[0].SegmentsCount != 1 {
			t.Fatalf("texts = %+v", resp.Texts)
		}
	})

	t.Run("empty title falls back to text id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/reading/texts?course_code=en_ru&level=A2", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTextsList(rr, req)
		var resp struct {
			Texts []adminReadingTextItem `json:"texts"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Texts) != 1 || resp.Texts[0].Title != "beta-text" {
			t.Fatalf("texts = %+v", resp.Texts)
		}
	})

	t.Run("course not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/reading/texts?course_code=nope_xx", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTextsList(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("index load error", func(t *testing.T) {
		badBundle := t.TempDir()
		if err := os.MkdirAll(filepath.Join(badBundle, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(badBundle, "reading", "index.json"), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`TRUNCATE reading_categories, reading_texts RESTART IDENTITY CASCADE`); err != nil {
			t.Fatal(err)
		}
		badRouter := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{
			GrammarBundleID:  "en",
			GrammarBundleDir: badBundle,
			ContentSource:    "bundle",
		}}, db, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/admin/reading/texts?course_code=en_ru", nil)
		rr := httptest.NewRecorder()
		badRouter.handleAdminReadingTextsList(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
		}
		if err := repo.ReplaceCatalogForTargetLanguage("en", "1", "t", categories, texts); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("skips broken text doc", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/reading/texts?course_code=en_ru", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTextsList(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d", rr.Code)
		}
		var resp struct {
			Texts []adminReadingTextItem `json:"texts"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		for _, item := range resp.Texts {
			if item.TextID == "orphan-id" {
				t.Fatal("orphan-id should be skipped")
			}
		}
	})
}

func TestAdminReadingGap_ConfigForReadingCourse(t *testing.T) {
	base := &config.Config{Learning: config.LearningConfig{
		GrammarBundleID:  "en",
		GrammarBundleDir: "/tmp/en",
	}}
	if configForReadingCourse(nil, "en_ru") != nil {
		t.Fatal("nil cfg should return nil")
	}
	same := configForReadingCourse(base, "en_ru")
	if same.Learning.GrammarBundleID != "en" || same.Learning.GrammarBundleDir != "/tmp/en" {
		t.Fatalf("en config = %+v", same.Learning)
	}
	es := configForReadingCourse(base, "es_ru")
	if es.Learning.GrammarBundleID != "es" || es.Learning.GrammarBundleDir != "" {
		t.Fatalf("es config = %+v", es.Learning)
	}
}

func TestAdminReadingGap_ReadingWritableRootDirForCourse(t *testing.T) {
	dir := adminReadingGapWriteBundle(t, "wrt-text", "en", "")
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleID:  "en",
		GrammarBundleDir: dir,
	}}
	got, err := readingWritableRootDirForCourse(cfg, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(dir)
	if got != abs {
		t.Fatalf("got %q want %q", got, abs)
	}
	esDir := adminReadingGapWriteBundle(t, "es-wrt", "es", "texts/es-wrt.json")
	esCfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleID:  "es",
		GrammarBundleDir: esDir,
	}}
	gotES, err := readingWritableRootDirForCourse(esCfg, "es_ru")
	if err != nil {
		t.Fatal(err)
	}
	absES, _ := filepath.Abs(esDir)
	if gotES != absES {
		t.Fatalf("es got %q want %q", gotES, absES)
	}
}

func TestAdminReadingGap_HandleAdminReadingTextDelete_Validation(t *testing.T) {
	router, _ := adminReadingGapRouter(t, config.LearningConfig{ContentSource: "db"})

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/?course_code=en_ru", nil)
	rr := httptest.NewRecorder()
	router.handleAdminReadingTextDelete(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("empty text_id status = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/missing?course_code=bad_course", nil)
	rr = httptest.NewRecorder()
	router.handleAdminReadingTextDelete(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("bad course status = %d", rr.Code)
	}
}

func TestAdminReadingGap_HandleAdminReadingTextDelete_DBOnlySuccess(t *testing.T) {
	router, db := adminReadingGapRouter(t, config.LearningConfig{
		TargetLang:    "en",
		ContentSource: "db",
		GrammarBundleID: "zz-no-bundle",
	})
	adminReadingGapSeedCatalog(t, db, repository.ReadingTextUpsert{
		TextID:             "db-only",
		CategoryID:         "gap_cat",
		Title:              "DB only",
		Level:              "A1",
		TargetLanguage:     "en",
		ReadingPassageJSON: `{}`,
	})
	if _, err := db.Exec(`INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids)
		VALUES ('solo', 'Solo', 'A1', 1, '["db-only"]')`); err != nil {
		t.Fatal(err)
	}
	var courseID int64
	if err := db.QueryRow(`SELECT id FROM courses WHERE code = 'en_ru'`).Scan(&courseID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO learning_items (course_id, item_type, source_kind, source_id, title, cefr_level, status)
		VALUES (?, 'reading_text', 'reading_text', 'db-only', 'DB only', 'A1', 'published')`, courseID); err != nil {
		t.Fatal(err)
	}
	userRepo := repository.NewUserRepository(db, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(adminReadingGapTelegramID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO reading_text_progress (user_id, chapter_id) VALUES (?, 'db-only')`, user.ID); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/db-only?course_code=en_ru", nil)
	rr := httptest.NewRecorder()
	router.handleAdminReadingTextDelete(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_texts WHERE text_id = 'db-only'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("reading_texts row should be deleted")
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories WHERE category_id = 'solo'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("empty category should be removed")
	}
}

func TestAdminReadingGap_HandleAdminReadingTextDelete_FilesystemAndDB(t *testing.T) {
	bundleDir := adminReadingGapWriteBundle(t, "fs-del", "en", "texts/fs-del.json")
	router, db := adminReadingGapRouter(t, config.LearningConfig{
		TargetLang:       "en",
		GrammarBundleID:  "en",
		GrammarBundleDir: bundleDir,
		ContentSource:    "db",
	})
	adminReadingGapSeedCatalog(t, db, repository.ReadingTextUpsert{
		TextID:             "fs-del",
		CategoryID:         "gap_cat",
		Title:              "FS delete",
		Level:              "A1",
		TargetLanguage:     "en",
		ReadingPassageJSON: `{}`,
	})
	assetsDir := filepath.Join(bundleDir, "assets", "reading", "fs-del")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "a.mp3"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/fs-del?course_code=en_ru", nil)
	rr := httptest.NewRecorder()
	router.handleAdminReadingTextDelete(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
	}
	textFile := filepath.Join(bundleDir, "reading", "texts", "fs-del.json")
	if _, err := os.Stat(textFile); !os.IsNotExist(err) {
		t.Fatalf("text file should be removed: %v", err)
	}
	idxData, err := os.ReadFile(filepath.Join(bundleDir, "reading", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx map[string]interface{}
	if err := json.Unmarshal(idxData, &idx); err != nil {
		t.Fatal(err)
	}
	texts, _ := idx["texts"].(map[string]interface{})
	if len(texts) != 0 {
		t.Fatalf("index texts = %#v", texts)
	}
}

func TestAdminReadingGap_HandleAdminReadingTextDelete_FilesystemErrors(t *testing.T) {
	bundleDir := adminReadingGapWriteBundle(t, "err-text", "en", "texts/err-text.json")
	router, db := adminReadingGapRouter(t, config.LearningConfig{
		TargetLang:       "en",
		GrammarBundleID:  "en",
		GrammarBundleDir: bundleDir,
		ContentSource:    "db",
	})
	seed := repository.ReadingTextUpsert{
		TextID:             "err-text",
		CategoryID:         "gap_cat",
		Title:              "Err",
		Level:              "A1",
		TargetLanguage:     "en",
		ReadingPassageJSON: `{}`,
	}

	t.Run("wrong course language", func(t *testing.T) {
		adminReadingGapSeedCatalog(t, db, seed)
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/err-text?course_code=es_ru", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTextDelete(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("text not in catalog", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/absent?course_code=en_ru", nil)
		rr := httptest.NewRecorder()
		router.handleAdminReadingTextDelete(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d", rr.Code)
		}
	})

	t.Run("missing filesystem index", func(t *testing.T) {
		adminReadingGapSeedCatalog(t, db, seed)
		noIdxDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(noIdxDir, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		fsRouter := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{
			GrammarBundleID:  "en",
			GrammarBundleDir: noIdxDir,
			ContentSource:    "db",
		}}, db, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/err-text?course_code=en_ru", nil)
		rr := httptest.NewRecorder()
		fsRouter.handleAdminReadingTextDelete(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("invalid index json", func(t *testing.T) {
		adminReadingGapSeedCatalog(t, db, seed)
		badIdxDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(badIdxDir, "reading"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(badIdxDir, "reading", "index.json"), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		badRouter := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{
			GrammarBundleID:  "en",
			GrammarBundleDir: badIdxDir,
			ContentSource:    "db",
		}}, db, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/err-text?course_code=en_ru", nil)
		rr := httptest.NewRecorder()
		badRouter.handleAdminReadingTextDelete(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("invalid rel path", func(t *testing.T) {
		adminReadingGapSeedCatalog(t, db, seed)
		evilDir := adminReadingGapWriteBundle(t, "err-text", "en", "../evil.json")
		idx := map[string]interface{}{
			"version": "1", "generated_at": "t",
			"categories": map[string]interface{}{},
			"texts":      map[string]string{"err-text": "../evil.json"},
		}
		raw, _ := json.MarshalIndent(idx, "", "  ")
		_ = os.WriteFile(filepath.Join(evilDir, "reading", "index.json"), append(raw, '\n'), 0o644)
		evilRouter := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{
			GrammarBundleID:  "en",
			GrammarBundleDir: evilDir,
			ContentSource:    "db",
		}}, db, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/err-text?course_code=en_ru", nil)
		rr := httptest.NewRecorder()
		evilRouter.handleAdminReadingTextDelete(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("text missing from filesystem index", func(t *testing.T) {
		adminReadingGapSeedCatalog(t, db, seed)
		emptyDir := adminReadingGapWriteBundle(t, "other", "en", "texts/other.json")
		mismatchRouter := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{
			GrammarBundleID:  "en",
			GrammarBundleDir: emptyDir,
			ContentSource:    "db",
		}}, db, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/err-text?course_code=en_ru", nil)
		rr := httptest.NewRecorder()
		mismatchRouter.handleAdminReadingTextDelete(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
		}
	})
}

func TestAdminReadingGap_HandleAdminReadingTextDelete_NoWritableRootBundleMode(t *testing.T) {
	router, db := adminReadingGapRouter(t, config.LearningConfig{
		ContentSource:   "bundle",
		GrammarBundleID: "zz-missing",
	})
	adminReadingGapSeedCatalog(t, db, repository.ReadingTextUpsert{
		TextID:             "bundle-mode",
		CategoryID:         "gap_cat",
		Title:              "Bundle",
		Level:              "A1",
		TargetLanguage:     "en",
		ReadingPassageJSON: `{}`,
	})
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/bundle-mode?course_code=en_ru", nil)
	rr := httptest.NewRecorder()
	router.handleAdminReadingTextDelete(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAdminReadingGap_HandleAdminReadingTextDelete_CatalogCommitFailure(t *testing.T) {
	bundleDir := adminReadingGapWriteBundle(t, "db-fail", "en", "texts/db-fail.json")
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })

	mock.ExpectQuery(`SELECT target_lang FROM courses WHERE code = \? AND status = 'active'`).
		WithArgs("en_ru").
		WillReturnRows(sqlmock.NewRows([]string{"target_lang"}).AddRow("en"))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM reading_categories`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT category_id, title, COALESCE\(title_translations`).
		WillReturnRows(sqlmock.NewRows([]string{
			"category_id", "title", "title_translations", "level", "sort_order", "text_ids",
		}).AddRow("gap_cat", "Gap", "", "A1", 1, `["db-fail"]`))
	mock.ExpectQuery(`SELECT text_id, category_id, title, COALESCE\(title_translations`).
		WithArgs("db-fail").
		WillReturnRows(sqlmock.NewRows([]string{
			"text_id", "category_id", "title", "title_translations", "level", "target_language",
			"cover_thumb_rel_path", "cover_hero_rel_path", "cover_image_prompt", "reading_passage",
		}).AddRow("db-fail", "gap_cat", "Fail", "", "A1", "en", "", "", "", `{}`))
	mock.ExpectBegin().WillReturnError(fmt.Errorf("commit path failed"))
	mock.ExpectRollback()

	router := NewRouter(zap.NewNop(), &config.Config{Learning: config.LearningConfig{
		GrammarBundleID:  "en",
		GrammarBundleDir: bundleDir,
		ContentSource:    "db",
	}}, mockDB, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/reading/texts/db-fail?course_code=en_ru", nil)
	rr := httptest.NewRecorder()
	router.handleAdminReadingTextDelete(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAdminReadingGap_DeleteReadingTextFromCatalogDB(t *testing.T) {
	t.Run("nil db", func(t *testing.T) {
		router, _ := adminReadingGapRouter(t, config.LearningConfig{})
		router.db = nil
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		if err := router.deleteReadingTextFromCatalogDB(req, "x", "en_ru", "en"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("success updates and prunes categories", func(t *testing.T) {
		router, db := adminReadingGapRouter(t, config.LearningConfig{ContentSource: "db"})
		if _, err := db.Exec(`
			INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
			VALUES ('del-a', 'cat-a', 'A', 'A1', 'en', '{}'),
			       ('del-b', 'cat-b', 'B', 'A1', 'en', '{}')`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`
			INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids)
			VALUES ('cat-a', 'A', 'A1', 1, '["del-a","keep"]'),
			       ('cat-b', 'B', 'A1', 1, '["del-b"]'),
			       ('cat-bad', 'Bad', 'A1', 1, 'not-json')`); err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		if err := router.deleteReadingTextFromCatalogDB(req, "del-a", "en_ru", "en"); err != nil {
			t.Fatal(err)
		}
		var raw string
		if err := db.QueryRow(`SELECT text_ids FROM reading_categories WHERE category_id = 'cat-a'`).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var ids []string
		if err := json.Unmarshal([]byte(raw), &ids); err != nil {
			t.Fatal(err)
		}
		if len(ids) != 1 || ids[0] != "keep" {
			t.Fatalf("cat-a text_ids = %v", ids)
		}
		if err := router.deleteReadingTextFromCatalogDB(req, "del-b", "", ""); err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories WHERE category_id = 'cat-b'`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatal("cat-b should be deleted")
		}
	})

	t.Run("begin tx error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = mockDB.Close() })
		mock.ExpectBegin().WillReturnError(fmt.Errorf("begin failed"))
		router := NewRouter(zap.NewNop(), &config.Config{}, mockDB, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		if err := router.deleteReadingTextFromCatalogDB(req, "x", "en_ru", "en"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("delete reading_texts error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = mockDB.Close() })
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM reading_texts").WillReturnError(fmt.Errorf("delete failed"))
		mock.ExpectRollback()
		router := NewRouter(zap.NewNop(), &config.Config{}, mockDB, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		if err := router.deleteReadingTextFromCatalogDB(req, "x", "en_ru", "en"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("categories query error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = mockDB.Close() })
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM reading_texts").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("DELETE FROM reading_text_progress").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("DELETE FROM learning_items").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery("SELECT category_id, text_ids FROM reading_categories").WillReturnError(fmt.Errorf("query failed"))
		mock.ExpectRollback()
		router := NewRouter(zap.NewNop(), &config.Config{}, mockDB, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		if err := router.deleteReadingTextFromCatalogDB(req, "x", "en_ru", "en"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("progress delete error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = mockDB.Close() })
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM reading_texts").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("DELETE FROM reading_text_progress").WillReturnError(fmt.Errorf("progress failed"))
		mock.ExpectRollback()
		router := NewRouter(zap.NewNop(), &config.Config{}, mockDB, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		if err := router.deleteReadingTextFromCatalogDB(req, "x", "", ""); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("category update error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = mockDB.Close() })
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM reading_texts").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("DELETE FROM reading_text_progress").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery("SELECT category_id, text_ids FROM reading_categories").
			WillReturnRows(sqlmock.NewRows([]string{"category_id", "text_ids"}).
				AddRow("c1", `["x","y"]`))
		mock.ExpectExec("UPDATE reading_categories SET text_ids").
			WillReturnError(fmt.Errorf("update failed"))
		mock.ExpectRollback()
		router := NewRouter(zap.NewNop(), &config.Config{}, mockDB, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		if err := router.deleteReadingTextFromCatalogDB(req, "x", "", ""); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAdminReadingGap_ApplyReadingTextDeletion(t *testing.T) {
	root := t.TempDir()
	readingDir := filepath.Join(root, "reading", "texts")
	if err := os.MkdirAll(readingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	textPath := filepath.Join(readingDir, "doc.json")
	if err := os.WriteFile(textPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(root, "reading", "index.json")
	idx := &readingIndex{
		Version: "1", GeneratedAt: "t",
		Categories: map[string]readingCategory{"c1": {ID: "c1", TextIDs: []string{"tid"}}},
		Texts:      map[string]string{"tid": "texts/doc.json"},
	}
	if err := applyReadingTextDeletion(root, indexPath, idx, "tid", "../bad"); err == nil {
		t.Fatal("expected invalid path error")
	}
	if err := applyReadingTextDeletion(root, indexPath, idx, "tid", "texts/doc.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(textPath); !os.IsNotExist(err) {
		t.Fatal("text file should be removed")
	}
	if idx.Texts["tid"] != "" {
		t.Fatal("index entry should be removed")
	}
	if idx.GeneratedAt == "" || idx.GeneratedAt == "t" {
		_ = nowRFC3339UTC()
	}
}
