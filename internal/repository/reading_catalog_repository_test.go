package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupReadingCatalogRepo(t *testing.T) (*ReadingCatalogRepository, *sql.DB) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	if _, err := db.Exec(`DELETE FROM reading_text_progress`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM reading_texts`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM reading_categories`); err != nil {
		t.Fatal(err)
	}
	return NewReadingCatalogRepository(db), db
}

func sampleReadingCatalogUpserts() ([]ReadingCategoryUpsert, []ReadingTextUpsert) {
	categories := []ReadingCategoryUpsert{
		{
			CategoryID:        "en_a1",
			Title:             "Daily Life",
			TitleTranslations: map[string]string{"ru": "Быт"},
			Level:             "A1",
			SortOrder:         2,
			TextIDs:           []string{"en-text-1", "en-text-2"},
		},
		{
			CategoryID: "en_a2",
			Title:      "Travel",
			Level:      "A2",
			SortOrder:  1,
			TextIDs:    []string{"en-text-3"},
		},
	}
	texts := []ReadingTextUpsert{
		{
			TextID:             "en-text-1",
			CategoryID:         "en_a1",
			Title:              "Morning coffee",
			TitleTranslations:  map[string]string{"ru": "Утренний кофе"},
			Level:              "A1",
			TargetLanguage:     "en",
			CoverThumbRelPath:  "covers/en-text-1-thumb.webp",
			ReadingPassageJSON: `{"segments":[{"tokens":[{"lemma":"coffee","clickable":true}]}]}`,
		},
		{
			TextID:             "en-text-2",
			CategoryID:         "en_a1",
			Title:              "At the shop",
			Level:              "A1",
			TargetLanguage:     "en",
			ReadingPassageJSON: `{"segments":[]}`,
		},
		{
			TextID:             "en-text-3",
			CategoryID:         "en_a2",
			Title:              "At the airport",
			Level:              "A2",
			TargetLanguage:     "en",
			ReadingPassageJSON: `{"segments":[]}`,
		},
	}
	return categories, texts
}

func TestNewReadingCatalogRepository_NilDB(t *testing.T) {
	if got := NewReadingCatalogRepository(nil); got != nil {
		t.Fatal("expected nil repo for nil db")
	}
}

func TestReadingCatalogRepository_CountCategories(t *testing.T) {
	repo, _ := setupReadingCatalogRepo(t)

	n, err := repo.CountCategories()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("count = %d, want 0", n)
	}

	categories, texts := sampleReadingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, texts); err != nil {
		t.Fatal(err)
	}

	n, err = repo.CountCategories()
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}
}

func TestReadingCatalogRepository_LoadSnapshot(t *testing.T) {
	repo, _ := setupReadingCatalogRepo(t)
	categories, texts := sampleReadingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, texts); err != nil {
		t.Fatal(err)
	}

	snap, err := repo.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snap == nil || len(snap.Categories) != 2 {
		t.Fatalf("snapshot categories = %#v", snap)
	}
	cat := snap.Categories["en_a1"]
	if cat == nil || cat.Title != "Daily Life" || cat.Order != 2 || len(cat.TextIDs) != 2 {
		t.Fatalf("en_a1 = %#v", cat)
	}
	if cat.TitleTranslations["ru"] != "Быт" {
		t.Fatalf("title translations = %#v", cat.TitleTranslations)
	}
}

func TestReadingCatalogRepository_GetTextDocument(t *testing.T) {
	repo, _ := setupReadingCatalogRepo(t)
	categories, texts := sampleReadingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, texts); err != nil {
		t.Fatal(err)
	}

	doc, ok, err := repo.GetTextDocument("en-text-1")
	if err != nil || !ok || doc == nil {
		t.Fatalf("GetTextDocument() = %#v ok=%v err=%v", doc, ok, err)
	}
	if doc.Title != "Morning coffee" || doc.TargetLanguage != "en" {
		t.Fatalf("doc = %#v", doc)
	}
	if doc.CoverThumbRelPath != "covers/en-text-1-thumb.webp" {
		t.Fatalf("cover = %q", doc.CoverThumbRelPath)
	}
	if doc.ReadingPassage["segments"] == nil {
		t.Fatal("expected reading passage segments")
	}

	missing, ok, err := repo.GetTextDocument("missing-text")
	if err != nil || ok || missing != nil {
		t.Fatalf("missing doc = %#v ok=%v err=%v", missing, ok, err)
	}
}

func TestReadingCatalogRepository_AllTextIDs(t *testing.T) {
	repo, _ := setupReadingCatalogRepo(t)
	categories, texts := sampleReadingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, texts); err != nil {
		t.Fatal(err)
	}

	ids, err := repo.AllTextIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 {
		t.Fatalf("ids = %#v", ids)
	}
}

func TestReadingCatalogRepository_ReplaceCatalog_ReplacesAll(t *testing.T) {
	repo, db := setupReadingCatalogRepo(t)
	categories, texts := sampleReadingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, texts); err != nil {
		t.Fatal(err)
	}

	esCategories := []ReadingCategoryUpsert{{
		CategoryID: "es_a1",
		Title:      "Spanish",
		Level:      "A1",
		SortOrder:  1,
		TextIDs:    []string{"es-text-1"},
	}}
	esTexts := []ReadingTextUpsert{{
		TextID:             "es-text-1",
		CategoryID:         "es_a1",
		Title:              "Hola",
		Level:              "A1",
		TargetLanguage:     "es",
		ReadingPassageJSON: `{"segments":[]}`,
	}}
	if err := repo.ReplaceCatalog("2.0", "t2", esCategories, esTexts); err != nil {
		t.Fatal(err)
	}

	var enCount, esCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories WHERE category_id LIKE 'en_%'`).Scan(&enCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories WHERE category_id LIKE 'es_%'`).Scan(&esCount); err != nil {
		t.Fatal(err)
	}
	if enCount != 0 || esCount != 1 {
		t.Fatalf("en=%d es=%d after full replace", enCount, esCount)
	}
}

func TestReadingCatalogRepository_ReplaceCatalogForTargetLanguage_Scoped(t *testing.T) {
	repo, db := setupReadingCatalogRepo(t)
	enCategories, enTexts := sampleReadingCatalogUpserts()
	if err := repo.ReplaceCatalogForTargetLanguage("en", "1.0", "t", enCategories, enTexts); err != nil {
		t.Fatal(err)
	}

	esCategories := []ReadingCategoryUpsert{{
		CategoryID: "es_a1",
		Title:      "Spanish",
		Level:      "A1",
		SortOrder:  1,
		TextIDs:    []string{"es-text-1"},
	}}
	esTexts := []ReadingTextUpsert{{
		TextID:             "es-text-1",
		CategoryID:         "es_a1",
		Title:              "Hola",
		Level:              "A1",
		TargetLanguage:     "es",
		ReadingPassageJSON: `{"segments":[]}`,
	}}
	if err := repo.ReplaceCatalogForTargetLanguage("es", "1.0", "t", esCategories, esTexts); err != nil {
		t.Fatal(err)
	}

	var enCount, esCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories WHERE category_id LIKE 'en_%'`).Scan(&enCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories WHERE category_id LIKE 'es_%'`).Scan(&esCount); err != nil {
		t.Fatal(err)
	}
	if enCount != 2 || esCount != 1 {
		t.Fatalf("en=%d es=%d after scoped imports", enCount, esCount)
	}

	updatedCategories := []ReadingCategoryUpsert{{
		CategoryID: "en_a1",
		Title:      "Updated Daily Life",
		Level:      "A1",
		SortOrder:  1,
		TextIDs:    []string{"en-text-1"},
	}}
	updatedTexts := []ReadingTextUpsert{{
		TextID:             "en-text-1",
		CategoryID:         "en_a1",
		Title:              "Updated coffee",
		Level:              "A1",
		TargetLanguage:     "en",
		ReadingPassageJSON: `{"segments":[]}`,
	}}
	if err := repo.ReplaceCatalogForTargetLanguage("en", "2.0", "t2", updatedCategories, updatedTexts); err != nil {
		t.Fatal(err)
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories WHERE category_id LIKE 'en_%'`).Scan(&enCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_categories WHERE category_id LIKE 'es_%'`).Scan(&esCount); err != nil {
		t.Fatal(err)
	}
	if enCount != 1 || esCount != 1 {
		t.Fatalf("after en replace: en=%d es=%d", enCount, esCount)
	}

	doc, ok, err := repo.GetTextDocument("en-text-1")
	if err != nil || !ok || doc.Title != "Updated coffee" {
		t.Fatalf("updated doc = %#v ok=%v err=%v", doc, ok, err)
	}
	if _, ok, err := repo.GetTextDocument("es-text-1"); err != nil || !ok {
		t.Fatalf("es text should remain: ok=%v err=%v", ok, err)
	}
}

func TestReadingCatalogRepository_ReplaceCatalogForTargetLanguage_TargetMismatch(t *testing.T) {
	repo, _ := setupReadingCatalogRepo(t)
	categories := []ReadingCategoryUpsert{{
		CategoryID: "es_a1",
		Title:      "Spanish",
		Level:      "A1",
		SortOrder:  1,
		TextIDs:    []string{"es-text-1"},
	}}
	texts := []ReadingTextUpsert{{
		TextID:             "es-text-1",
		CategoryID:         "es_a1",
		Title:              "Hello",
		Level:              "A1",
		TargetLanguage:     "en",
		ReadingPassageJSON: `{"segments":[]}`,
	}}

	err := repo.ReplaceCatalogForTargetLanguage("es", "1.0", "t", categories, texts)
	if err == nil {
		t.Fatal("expected target language mismatch error")
	}
}

func TestReadingCatalogRepository_ReplaceCatalog_CleansOrphanProgress(t *testing.T) {
	repo, db := setupReadingCatalogRepo(t)
	categories, texts := sampleReadingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, texts); err != nil {
		t.Fatal(err)
	}

	userRepo := NewUserRepository(db, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(4242)
	if err != nil {
		t.Fatal(err)
	}
	progressRepo := NewReadingTextProgressRepository(db)
	if err := progressRepo.MarkRead(user.ID, "en-text-2"); err != nil {
		t.Fatal(err)
	}

	replacedCategories := []ReadingCategoryUpsert{{
		CategoryID: "en_a1",
		Title:      "Daily Life",
		Level:      "A1",
		SortOrder:  1,
		TextIDs:    []string{"en-text-1"},
	}}
	replacedTexts := []ReadingTextUpsert{{
		TextID:             "en-text-1",
		CategoryID:         "en_a1",
		Title:              "Morning coffee",
		Level:              "A1",
		TargetLanguage:     "en",
		ReadingPassageJSON: `{"segments":[]}`,
	}}
	if err := repo.ReplaceCatalog("2.0", "t2", replacedCategories, replacedTexts); err != nil {
		t.Fatal(err)
	}

	var progressCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reading_text_progress WHERE chapter_id = 'en-text-2'`).Scan(&progressCount); err != nil {
		t.Fatal(err)
	}
	if progressCount != 0 {
		t.Fatalf("orphan progress count = %d, want 0", progressCount)
	}
}

func TestReadingCatalogRepository_NilReceiver(t *testing.T) {
	var repo *ReadingCatalogRepository

	if n, err := repo.CountCategories(); err != nil || n != 0 {
		t.Fatalf("CountCategories() = %d err=%v", n, err)
	}
	if _, ok, err := repo.GetTextDocument("x"); err != nil || ok {
		t.Fatalf("GetTextDocument() ok=%v err=%v", ok, err)
	}
	if ids, err := repo.AllTextIDs(); err != nil || ids != nil {
		t.Fatalf("AllTextIDs() = %#v err=%v", ids, err)
	}
	if err := repo.ReplaceCatalog("1", "t", nil, nil); err == nil {
		t.Fatal("ReplaceCatalog expected error for nil db")
	}
	if err := repo.ReplaceCatalogForTargetLanguage("en", "1", "t", nil, nil); err == nil {
		t.Fatal("ReplaceCatalogForTargetLanguage expected error for nil db")
	}
	if _, err := repo.LoadSnapshot(); err == nil {
		t.Fatal("LoadSnapshot expected error for nil db")
	}
}
