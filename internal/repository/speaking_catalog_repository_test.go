package repository

import (
	"database/sql"
	"encoding/json"
	"testing"

	"tgbot-skeleton/internal/testutil"
)

func setupSpeakingCatalogRepo(t *testing.T) (*SpeakingCatalogRepository, *sql.DB) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	for _, q := range []string{
		`DELETE FROM speaking_tasks`,
		`DELETE FROM speaking_categories`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	return NewSpeakingCatalogRepository(db), db
}

func sampleSpeakingTaskJSON(taskID, categoryID, targetLang, title string) string {
	full := SpeakingTaskFull{
		SpeakingTaskDocument: SpeakingTaskDocument{
			ID:             taskID,
			CategoryID:     categoryID,
			Level:          "A1",
			Type:           "answer",
			TargetLanguage: targetLang,
			Title:          title,
			PromptRU:       "Скажите",
			MaxAttempts:    3,
			Order:          1,
		},
		ExpectedMeaningRU: "привет",
		AcceptableAnswers: []string{"hola"},
	}
	raw, _ := json.Marshal(full)
	return string(raw)
}

func sampleSpeakingCatalogUpserts() ([]SpeakingCategoryUpsert, []SpeakingTaskUpsert) {
	categories := []SpeakingCategoryUpsert{
		{
			CategoryID:        "en_a1",
			Title:             "Basics",
			TitleTranslations: map[string]string{"ru": "Основы"},
			Level:             "A1",
			SortOrder:         2,
			TaskIDs:           []string{"en-task-1", "en-task-2"},
		},
		{
			CategoryID: "en_a2",
			Title:      "Travel",
			Level:      "A2",
			SortOrder:  1,
			TaskIDs:    []string{"en-task-3"},
		},
	}
	tasks := []SpeakingTaskUpsert{
		{
			TaskID:         "en-task-1",
			CategoryID:     "en_a1",
			Title:          "Hello",
			Level:          "A1",
			TaskType:       "answer",
			TargetLanguage: "en",
			SortOrder:      1,
			TaskJSON:       sampleSpeakingTaskJSON("en-task-1", "en_a1", "en", "Hello"),
		},
		{
			TaskID:         "en-task-2",
			CategoryID:     "en_a1",
			Title:          "Goodbye",
			Level:          "A1",
			TaskType:       "answer",
			TargetLanguage: "en",
			SortOrder:      2,
			TaskJSON:       sampleSpeakingTaskJSON("en-task-2", "en_a1", "en", "Goodbye"),
		},
		{
			TaskID:         "en-task-3",
			CategoryID:     "en_a2",
			Title:          "At the airport",
			Level:          "A2",
			TaskType:       "answer",
			TargetLanguage: "en",
			SortOrder:      1,
			TaskJSON:       sampleSpeakingTaskJSON("en-task-3", "en_a2", "en", "At the airport"),
		},
	}
	return categories, tasks
}

func TestNewSpeakingCatalogRepository_NilDB(t *testing.T) {
	if got := NewSpeakingCatalogRepository(nil); got != nil {
		t.Fatal("expected nil repo for nil db")
	}
}

func TestSpeakingCatalogRepository_CountCategories(t *testing.T) {
	repo, _ := setupSpeakingCatalogRepo(t)

	n, err := repo.CountCategories()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("count = %d, want 0", n)
	}

	categories, tasks := sampleSpeakingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, tasks); err != nil {
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

func TestSpeakingCatalogRepository_LoadSnapshot(t *testing.T) {
	repo, _ := setupSpeakingCatalogRepo(t)
	categories, tasks := sampleSpeakingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, tasks); err != nil {
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
	if cat == nil || cat.Title != "Basics" || cat.Order != 2 || len(cat.TaskIDs) != 2 {
		t.Fatalf("en_a1 = %#v", cat)
	}
	if cat.TitleTranslations["ru"] != "Основы" {
		t.Fatalf("title translations = %#v", cat.TitleTranslations)
	}
}

func TestSpeakingCatalogRepository_GetTaskFullAndPublic(t *testing.T) {
	repo, _ := setupSpeakingCatalogRepo(t)
	categories, tasks := sampleSpeakingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, tasks); err != nil {
		t.Fatal(err)
	}

	full, ok, err := repo.GetTaskFull("en-task-1")
	if err != nil || !ok || full == nil {
		t.Fatalf("GetTaskFull() = %#v ok=%v err=%v", full, ok, err)
	}
	if full.Title != "Hello" || full.ExpectedMeaningRU != "привет" {
		t.Fatalf("full = %#v", full)
	}

	pub, ok, err := repo.GetTaskPublic("en-task-1")
	if err != nil || !ok || pub == nil {
		t.Fatalf("GetTaskPublic() = %#v ok=%v err=%v", pub, ok, err)
	}
	if pub.Title != "Hello" || pub.TargetLanguage != "en" {
		t.Fatalf("pub = %#v", pub)
	}

	missing, ok, err := repo.GetTaskFull("missing-task")
	if err != nil || ok || missing != nil {
		t.Fatalf("missing full = %#v ok=%v err=%v", missing, ok, err)
	}
}

func TestSpeakingCatalogRepository_ListTasksByCategory(t *testing.T) {
	repo, _ := setupSpeakingCatalogRepo(t)
	categories, tasks := sampleSpeakingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, tasks); err != nil {
		t.Fatal(err)
	}

	docs, err := repo.ListTasksByCategory("en_a1")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("docs = %#v", docs)
	}
	if docs[0].Title != "Hello" || docs[1].Title != "Goodbye" {
		t.Fatalf("order = %#v", docs)
	}
}

func TestSpeakingCatalogRepository_ReplaceCatalog_ReplacesAll(t *testing.T) {
	repo, db := setupSpeakingCatalogRepo(t)
	categories, tasks := sampleSpeakingCatalogUpserts()
	if err := repo.ReplaceCatalog("1.0", "t", categories, tasks); err != nil {
		t.Fatal(err)
	}

	esCategories := []SpeakingCategoryUpsert{{
		CategoryID: "es_a1",
		Title:      "Spanish",
		Level:      "A1",
		SortOrder:  1,
		TaskIDs:    []string{"es-task-1"},
	}}
	esTasks := []SpeakingTaskUpsert{{
		TaskID:         "es-task-1",
		CategoryID:     "es_a1",
		Title:          "Hola",
		Level:          "A1",
		TaskType:       "answer",
		TargetLanguage: "es",
		SortOrder:      1,
		TaskJSON:       sampleSpeakingTaskJSON("es-task-1", "es_a1", "es", "Hola"),
	}}
	if err := repo.ReplaceCatalog("2.0", "t2", esCategories, esTasks); err != nil {
		t.Fatal(err)
	}

	var enCount, esCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM speaking_categories WHERE category_id LIKE 'en_%'`).Scan(&enCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM speaking_categories WHERE category_id LIKE 'es_%'`).Scan(&esCount); err != nil {
		t.Fatal(err)
	}
	if enCount != 0 || esCount != 1 {
		t.Fatalf("en=%d es=%d after full replace", enCount, esCount)
	}
}

func TestSpeakingCatalogRepository_ReplaceCatalogForTargetLanguage_Scoped(t *testing.T) {
	repo, db := setupSpeakingCatalogRepo(t)
	enCategories, enTasks := sampleSpeakingCatalogUpserts()
	if err := repo.ReplaceCatalogForTargetLanguage("en", "1.0", "t", enCategories, enTasks); err != nil {
		t.Fatal(err)
	}

	esCategories := []SpeakingCategoryUpsert{{
		CategoryID: "es_a1",
		Title:      "Spanish",
		Level:      "A1",
		SortOrder:  1,
		TaskIDs:    []string{"es-task-1"},
	}}
	esTasks := []SpeakingTaskUpsert{{
		TaskID:         "es-task-1",
		CategoryID:     "es_a1",
		Title:          "Hola",
		Level:          "A1",
		TaskType:       "answer",
		TargetLanguage: "es",
		SortOrder:      1,
		TaskJSON:       sampleSpeakingTaskJSON("es-task-1", "es_a1", "es", "Hola"),
	}}
	if err := repo.ReplaceCatalogForTargetLanguage("es", "1.0", "t", esCategories, esTasks); err != nil {
		t.Fatal(err)
	}

	var enCount, esCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM speaking_categories WHERE category_id LIKE 'en_%'`).Scan(&enCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM speaking_categories WHERE category_id LIKE 'es_%'`).Scan(&esCount); err != nil {
		t.Fatal(err)
	}
	if enCount != 2 || esCount != 1 {
		t.Fatalf("en=%d es=%d after scoped imports", enCount, esCount)
	}

	updatedCategories := []SpeakingCategoryUpsert{{
		CategoryID: "en_a1",
		Title:      "Updated Basics",
		Level:      "A1",
		SortOrder:  1,
		TaskIDs:    []string{"en-task-1"},
	}}
	updatedTasks := []SpeakingTaskUpsert{{
		TaskID:         "en-task-1",
		CategoryID:     "en_a1",
		Title:          "Updated Hello",
		Level:          "A1",
		TaskType:       "answer",
		TargetLanguage: "en",
		SortOrder:      1,
		TaskJSON:       sampleSpeakingTaskJSON("en-task-1", "en_a1", "en", "Updated Hello"),
	}}
	if err := repo.ReplaceCatalogForTargetLanguage("en", "2.0", "t2", updatedCategories, updatedTasks); err != nil {
		t.Fatal(err)
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM speaking_categories WHERE category_id LIKE 'en_%'`).Scan(&enCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM speaking_categories WHERE category_id LIKE 'es_%'`).Scan(&esCount); err != nil {
		t.Fatal(err)
	}
	if enCount != 1 || esCount != 1 {
		t.Fatalf("after en replace: en=%d es=%d", enCount, esCount)
	}

	full, ok, err := repo.GetTaskFull("en-task-1")
	if err != nil || !ok || full.Title != "Updated Hello" {
		t.Fatalf("updated task = %#v ok=%v err=%v", full, ok, err)
	}
	if _, ok, err := repo.GetTaskFull("es-task-1"); err != nil || !ok {
		t.Fatalf("es task should remain: ok=%v err=%v", ok, err)
	}
}

func TestSpeakingCatalogRepository_ReplaceCatalogForTargetLanguage_TargetMismatch(t *testing.T) {
	repo, _ := setupSpeakingCatalogRepo(t)
	categories := []SpeakingCategoryUpsert{{
		CategoryID: "es_a1",
		Title:      "Spanish",
		Level:      "A1",
		SortOrder:  1,
		TaskIDs:    []string{"es-task-1"},
	}}
	tasks := []SpeakingTaskUpsert{{
		TaskID:         "es-task-1",
		CategoryID:     "es_a1",
		Title:          "Hello",
		Level:          "A1",
		TaskType:       "answer",
		TargetLanguage: "en",
		SortOrder:      1,
		TaskJSON:       sampleSpeakingTaskJSON("es-task-1", "es_a1", "en", "Hello"),
	}}

	err := repo.ReplaceCatalogForTargetLanguage("es", "1.0", "t", categories, tasks)
	if err == nil {
		t.Fatal("expected target language mismatch error")
	}
}

func TestSpeakingCatalogRepository_NilReceiver(t *testing.T) {
	var repo *SpeakingCatalogRepository

	if n, err := repo.CountCategories(); err != nil || n != 0 {
		t.Fatalf("CountCategories() = %d err=%v", n, err)
	}
	if _, ok, err := repo.GetTaskFull("x"); err != nil || ok {
		t.Fatalf("GetTaskFull() ok=%v err=%v", ok, err)
	}
	if _, ok, err := repo.GetTaskPublic("x"); err != nil || ok {
		t.Fatalf("GetTaskPublic() ok=%v err=%v", ok, err)
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
	if _, err := repo.ListTasksByCategory("x"); err == nil {
		t.Fatal("ListTasksByCategory expected error for nil db")
	}
}
