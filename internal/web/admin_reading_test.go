package web

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/readingbundle"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestHandleAdminReadingTextsListFiltersByCourse(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if _, err := db.Exec(`
		INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids)
		VALUES ('mixed', 'Mixed', 'A1', 1, '["en-text","es-text"]')
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
		VALUES
			('en-text', 'mixed', 'English text', 'A1', 'en', '{"segments":[]}'),
			('es-text', 'mixed', 'Spanish text', 'A1', 'es', '{"segments":[]}')
	`); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{
		TargetLang:    "en",
		NativeLang:    "ru",
		ContentSource: "db",
	}}
	router := NewRouter(zap.NewNop(), cfg, db, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/reading/texts?course_code=es_ru", nil)
	rr := httptest.NewRecorder()
	router.handleAdminReadingTextsList(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Texts []adminReadingTextItem `json:"texts"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Texts) != 1 || resp.Texts[0].TextID != "es-text" || resp.Texts[0].CourseCode != "es_ru" {
		t.Fatalf("texts = %+v, want only es_ru text", resp.Texts)
	}
}

func TestReadingWritableRootDir_GrammarBundleDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "reading"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reading", "index.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleDir: dir}}
	got, err := readingWritableRootDir(cfg)
	if err != nil {
		t.Fatal(err)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != absDir {
		t.Fatalf("got %q want %q", got, absDir)
	}
}

func TestReadingBundleFS_matchesWritableRootInRepo(t *testing.T) {
	t.Parallel()
	repoRoot := findRepoRoot(t)
	if repoRoot == "" {
		t.Skip("not inside english-ai-bot repo")
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "es", GrammarBundleDir: ""}}
	root, err := readingWritableRootDir(cfg)
	if err != nil {
		t.Fatal(err)
	}
	bundleFS, err := readingbundle.BundleFS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	data1, err := os.ReadFile(filepath.Join(root, "reading", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	data2, err := fs.ReadFile(bundleFS, "reading/index.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data1) != string(data2) {
		t.Fatal("readingBundleFS and readingWritableRootDir must use the same tree")
	}
}

func TestReadingWritableRootDir_DevRepoFallback(t *testing.T) {
	t.Parallel()
	repoRoot := findRepoRoot(t)
	if repoRoot == "" {
		t.Skip("not inside english-ai-bot repo")
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "es", GrammarBundleDir: ""}}
	got, err := readingWritableRootDir(cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(repoRoot, "internal", "grammarbundle", "es")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReadingWritableRootDir_NoDir(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "zz-unknown", GrammarBundleDir: ""}}
	_, err := readingWritableRootDir(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindRepoRootContainingCourses(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "courses"), 0o755); err != nil {
		t.Fatal(err)
	}
	bundle := filepath.Join(repo, "internal", "grammarbundle", "es")
	if err := os.MkdirAll(filepath.Join(bundle, "reading"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := findRepoRootContainingCourses(bundle)
	if got != repo {
		t.Fatalf("got %q want %q", got, repo)
	}
	if findRepoRootContainingCourses(t.TempDir()) != "" {
		t.Fatal("expected empty repo root for orphan dir")
	}
}

func TestSyncDeleteReadingTextInMatchingCourses_RemovesFromCourse(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	course := filepath.Join(repo, "courses", "spanish-grammar")
	if err := os.MkdirAll(filepath.Join(course, "reading", "texts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(course, "assets", "reading", "text_alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	dummyMP3 := filepath.Join(course, "assets", "reading", "text_alpha", "a.mp3")
	if err := os.WriteFile(dummyMP3, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	textJSON := filepath.Join(course, "reading", "texts", "doc.json")
	if err := os.WriteFile(textJSON, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	idx := map[string]interface{}{
		"version":      "1",
		"generated_at": "t",
		"categories": map[string]interface{}{
			"cat1": map[string]interface{}{
				"id":       "cat1",
				"text_ids": []string{"text_alpha"},
			},
		},
		"texts": map[string]string{"text_alpha": "texts/doc.json"},
	}
	idxBytes, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(course, "reading", "index.json"), append(idxBytes, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(course, "bundle.target"), []byte("es\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "courses"), 0o755); err != nil {
		t.Fatal(err)
	}

	bundleRoot := filepath.Join(repo, "internal", "grammarbundle", "es")
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "es"}}

	syncDeleteReadingTextInMatchingCourses(zap.NewNop(), cfg, bundleRoot, "text_alpha")

	if _, err := os.Stat(textJSON); !os.IsNotExist(err) {
		t.Fatalf("expected text json removed: %v", err)
	}
	if _, err := os.Stat(dummyMP3); !os.IsNotExist(err) {
		t.Fatalf("expected audio removed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(course, "reading", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var after map[string]interface{}
	if err := json.Unmarshal(data, &after); err != nil {
		t.Fatal(err)
	}
	texts, _ := after["texts"].(map[string]interface{})
	if len(texts) != 0 {
		t.Fatalf("texts map: %#v", texts)
	}
	cats, _ := after["categories"].(map[string]interface{})
	if len(cats) != 0 {
		t.Fatalf("expected empty categories, got %#v", cats)
	}
}

func TestSyncDeleteReadingTextInMatchingCourses_SkipsOtherBundle(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	course := filepath.Join(repo, "courses", "english")
	if err := os.MkdirAll(filepath.Join(course, "reading"), 0o755); err != nil {
		t.Fatal(err)
	}
	idx := `{"version":"1","generated_at":"t","categories":{},"texts":{"x":"texts/x.json"}}`
	if err := os.WriteFile(filepath.Join(course, "reading", "index.json"), []byte(idx), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(course, "bundle.target"), []byte("en\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "courses"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "es"}}
	syncDeleteReadingTextInMatchingCourses(zap.NewNop(), cfg, filepath.Join(repo, "internal", "grammarbundle", "es"), "x")

	data, err := os.ReadFile(filepath.Join(course, "reading", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != idx {
		t.Fatal("index should be unchanged when bundle.target does not match")
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for dir != "" && dir != filepath.Dir(dir) {
		if st, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !st.IsDir() {
			return dir
		}
		if st, err := os.Stat(filepath.Join(dir, "internal", "grammarbundle", "es", "reading", "index.json")); err == nil && !st.IsDir() {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}
