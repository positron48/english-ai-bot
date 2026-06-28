package wordsetimport

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func englishTestConfig() *config.Config {
	return &config.Config{
		Learning: config.LearningConfig{
			TargetLang: "en",
			AppCode:    "english",
		},
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for range 32 {
		if st, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !st.IsDir() {
			if st2, err2 := os.Stat(filepath.Join(dir, "courses")); err2 == nil && st2.IsDir() {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("repo root not found")
	return ""
}

func testEnglishCSVPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(findRepoRoot(t), "resources/wordsets/english_word_freq_pos_ud_top6000.filtered.csv")
}

func TestDetectEnglishCSVPathFallback(t *testing.T) {
	dir := findRepoRoot(t)
	t.Chdir(dir)
	path, err := detectEnglishCSVPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(path, "english_word_freq") {
		t.Fatalf("path=%q", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestDetectMustHaveYAMLPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.yaml")
	if _, err := detectMustHaveYAMLPath([]string{missing}); err == nil {
		t.Fatal("expected error")
	}
	path := filepath.Join(dir, "must-have.yaml")
	if err := os.WriteFile(path, []byte("must_have:\n  title: T\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := detectMustHaveYAMLPath([]string{missing, path})
	if err != nil || got != path {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func TestFileSHA256(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	sum, err := fileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sum) != 64 {
		t.Fatalf("sum=%q", sum)
	}
}

func TestDefaultDescription(t *testing.T) {
	t.Parallel()
	if !strings.Contains(defaultDescription("root", "Must Have"), "Must-have") {
		t.Fatal("root desc")
	}
	if !strings.Contains(defaultDescription("subcategory", "Travel"), "Travel") {
		t.Fatal("sub desc")
	}
	if !strings.Contains(defaultDescription("set", "Greetings"), "Greetings") {
		t.Fatal("set desc")
	}
}

func TestEnsureEnglishWordSetBlueprint(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	if err := EnsureEnglishWordSetBlueprint(conn); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM word_set_categories WHERE parent_id IS NULL`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count < 2 {
		t.Fatalf("root categories=%d", count)
	}
	if err := EnsureEnglishWordSetBlueprint(conn); err != nil {
		t.Fatal(err)
	}
}

func TestShouldForceEnglishImport(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	if err := EnsureEnglishWordSetBlueprint(conn); err != nil {
		t.Fatal(err)
	}
	force, reason, err := shouldForceEnglishImport(conn, "old", "new")
	if err != nil || !force || reason != "csv_checksum_changed" {
		t.Fatalf("force=%v reason=%q err=%v", force, reason, err)
	}
	force, reason, err = shouldForceEnglishImport(conn, "same", "same")
	if err != nil || !force || reason != "word_set_items_empty" {
		t.Fatalf("empty items force=%v reason=%q err=%v", force, reason, err)
	}
}

func TestEnsureEnglishLabels(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	_, err := conn.Exec(`
		INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
		VALUES (NULL, 'Frecuencia esencial', 'es', 1, 0)`)
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureEnglishLabels(conn); err != nil {
		t.Fatal(err)
	}
	var name string
	if err := conn.QueryRow(`
		SELECT name FROM word_set_categories WHERE parent_id IS NULL ORDER BY id DESC LIMIT 1`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Core Frequency" {
		t.Fatalf("name=%q", name)
	}
}

func TestAutoSyncEnglishWordSetsSkipsWhenUpToDate(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	log := zap.NewNop()
	cfg := englishTestConfig()
	t.Chdir(findRepoRoot(t))
	csvPath := testEnglishCSVPath(t)
	checksum, err := fileSHA256(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureEnglishWordSetBlueprint(conn); err != nil {
		t.Fatal(err)
	}
	res, err := Import(context.Background(), cfg, conn, log, ImportOptions{
		CSVPath:   csvPath,
		Lang:      "en",
		Commit:    true,
		LimitSets: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Processed == 0 {
		t.Fatal("expected at least one processed set")
	}
	settings := repository.NewAppSettingsRepository(conn, log)
	if err := settings.SetSetting(englishCSVChecksumSetting, checksum, systemUserID); err != nil {
		t.Fatal(err)
	}
	if err := AutoSyncEnglishWordSets(context.Background(), cfg, conn, log); err != nil {
		t.Fatal(err)
	}
}

func TestAutoSyncEnglishWordSetsNonEnglishNoop(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	cfg := &config.Config{
		Learning: config.LearningConfig{TargetLang: "es", AppCode: "spanish"},
	}
	if err := AutoSyncEnglishWordSets(context.Background(), cfg, conn, zap.NewNop()); err != nil {
		t.Fatal(err)
	}
}

func TestAutoSyncEnglishMustHaveWordSets(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	log := zap.NewNop()
	cfg := englishTestConfig()
	dir := findRepoRoot(t)
	t.Chdir(dir)
	if err := AutoSyncEnglishMustHaveWordSets(context.Background(), cfg, conn, log); err != nil {
		t.Fatal(err)
	}
	if err := AutoSyncEnglishMustHaveWordSets(context.Background(), cfg, conn, log); err != nil {
		t.Fatal(err)
	}
	var sets int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM word_sets WHERE title <> ''`).Scan(&sets); err != nil {
		t.Fatal(err)
	}
	if sets == 0 {
		t.Fatal("expected must-have sets")
	}
}

func TestUpsertCategoryAndWordSet(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	rootID, err := upsertCategory(conn, nil, "Test Root", "desc", 1)
	if err != nil || rootID == 0 {
		t.Fatalf("rootID=%d err=%v", rootID, err)
	}
	if err := ensureCategoryMetadata(conn, rootID, "desc2", 2); err != nil {
		t.Fatal(err)
	}
	childID, err := upsertCategory(conn, &rootID, "Child", "child desc", 0)
	if err != nil || childID == 0 {
		t.Fatalf("childID=%d err=%v", childID, err)
	}
	if err := upsertWordSet(conn, childID, "Set A", "set desc", "noun", 0); err != nil {
		t.Fatal(err)
	}
	pos := "verb"
	setID, err := ensureWordSet(conn, childID, "Set B", "set b", 1, &pos)
	if err != nil || setID == 0 {
		t.Fatalf("setID=%d err=%v", setID, err)
	}
}

func TestEnsureRangeSets(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	catID, err := upsertCategory(conn, nil, "Range cat", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureRangeSets(conn, catID, "Core Verbs", "Verbs", "verb", 1, 50, 50); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM word_sets WHERE category_id = $1`, catID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("sets=%d", count)
	}
}

func TestEnsureMustHaveBlueprintDryRun(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	root := &mustHaveRoot{
		Title: "Must Have Test",
		Subcategories: []mustHaveSubcategory{{
			Title: "Basics",
			Sets: []mustHaveSet{{
				Title: "Greetings",
				Words: []mustHaveWord{{EN: "hello", RU: "привет"}},
			}},
		}},
	}
	sets, items, err := ensureMustHaveBlueprint(context.Background(), conn, englishTestConfig(), root, "en", false, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if sets != 1 || items != 1 {
		t.Fatalf("sets=%d items=%d", sets, items)
	}
}
