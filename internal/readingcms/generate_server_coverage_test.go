package readingcms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockGenerateReadingScript is a stand-in for courses/*/scripts/generate-reading-text.py.
// Special --title values drive error branches; otherwise it writes a minimal draft bundle.
const mockGenerateReadingScript = `#!/usr/bin/env python3
import argparse, json, os, sys

p = argparse.ArgumentParser()
p.add_argument("--course-root")
p.add_argument("--draft-dir")
p.add_argument("--target-lang", default="en")
p.add_argument("--level", default="A2")
p.add_argument("--format", default="dialogue")
p.add_argument("--title", default="")
p.add_argument("--input-text")
p.add_argument("--input-json")
p.add_argument("--print-prompt", action="store_true")
p.add_argument("--prompt-kind", default="generate")
args, _ = p.parse_known_args()

if args.print_prompt:
    print("mock prompt kind=%s level=%s" % (args.prompt_kind, args.level))
    sys.exit(0)

if not args.title.strip():
    counter = os.path.join(args.course_root, ".mock_gen_counter_empty_title")
    n = 0
    if os.path.exists(counter):
        with open(counter) as f:
            n = int(f.read().strip() or "0")
    with open(counter, "w") as f:
        f.write(str(n + 1))
    if n >= 1:
        sys.stderr.write("empty-title second run fails\n")
        sys.exit(1)

title = args.title or "Generated title"
if title == "__MOCK_FAIL__":
    sys.stderr.write("mock script failed\n")
    sys.exit(1)
if title == "__MOCK_FAIL_SILENT__":
    sys.exit(1)
if title == "__MOCK_STDERR_WARN__":
    sys.stderr.write("non-fatal warning\n")

if title == "__MOCK_EMPTY_INDEX__":
    os.makedirs(args.draft_dir, exist_ok=True)
    with open(os.path.join(args.draft_dir, "index.json"), "w") as f:
        json.dump({"texts": {}}, f)
    sys.exit(0)

if title == "__MOCK_BAD_INDEX__":
    os.makedirs(args.draft_dir, exist_ok=True)
    with open(os.path.join(args.draft_dir, "index.json"), "w") as f:
        f.write("{not-json")
    sys.exit(0)

if title == "__MOCK_MISSING_TEXT__":
    os.makedirs(args.draft_dir, exist_ok=True)
    tid = "missing_text_doc"
    with open(os.path.join(args.draft_dir, "index.json"), "w") as f:
        json.dump({"texts": {tid: "texts/nope.json"}}, f)
    sys.exit(0)

if title == "__MOCK_BAD_DOC__":
    os.makedirs(os.path.join(args.draft_dir, "texts"), exist_ok=True)
    tid = "bad_doc"
    rel = "texts/" + tid + ".json"
    with open(os.path.join(args.draft_dir, rel), "w") as f:
        f.write("{bad")
    with open(os.path.join(args.draft_dir, "index.json"), "w") as f:
        json.dump({"texts": {tid: rel}}, f)
    sys.exit(0)

if title == "__MOCK_COUNTER_FAIL__":
    counter = os.path.join(args.course_root, ".mock_gen_counter")
    n = 0
    if os.path.exists(counter):
        with open(counter) as f:
            n = int(f.read().strip() or "0")
    with open(counter, "w") as f:
        f.write(str(n + 1))
    if n >= 1:
        sys.stderr.write("second run fails\n")
        sys.exit(1)

if not args.draft_dir:
    sys.stderr.write("draft-dir required\n")
    sys.exit(1)

os.makedirs(os.path.join(args.draft_dir, "texts"), exist_ok=True)
text_id = "gen_mock_%s" % args.target_lang
if args.input_json:
    text_id = "gen_json_%s" % args.target_lang
if args.input_text:
    text_id = "gen_text_%s" % args.target_lang
rel = "texts/" + text_id + ".json"
doc = {
    "id": text_id,
    "title": title,
    "level": args.level,
    "reading_passage": {"segments": [{"segment_id": "s1", "text": "Hi"}]},
}
if title == "__MOCK_NO_TARGET_LANG__":
    doc.pop("target_language", None)
else:
    doc["target_language"] = args.target_lang

with open(os.path.join(args.draft_dir, rel), "w") as f:
    json.dump(doc, f)
with open(os.path.join(args.draft_dir, "index.json"), "w") as f:
    json.dump({"texts": {text_id: rel}}, f)

if title == "__MOCK_WITH_AUDIO__":
    audio_dir = os.path.join(args.course_root, "assets", "reading", text_id)
    os.makedirs(audio_dir, exist_ok=True)
    with open(os.path.join(audio_dir, "seg_01.mp3"), "wb") as f:
        f.write(b"mp3")

sys.exit(0)
`

type genFixture struct {
	svc  *Service
	root string
}

func installMockReadingScript(t *testing.T, root, courseRel string) {
	t.Helper()
	courseDir := filepath.Join(root, courseRel)
	scriptsDir := filepath.Join(courseDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(scriptsDir, "generate-reading-text.py")
	if err := os.WriteFile(scriptPath, []byte(mockGenerateReadingScript), 0o755); err != nil {
		t.Fatal(err)
	}
	textsDir := filepath.Join(courseDir, "reading", "texts")
	if err := os.MkdirAll(textsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	idx := `{"version":"1.0.0","categories":{},"texts":{}}`
	if err := os.WriteFile(filepath.Join(courseDir, "reading", "index.json"), []byte(idx), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setupGenFixture(t *testing.T) *genFixture {
	t.Helper()
	root := t.TempDir()
	installMockReadingScript(t, root, "courses/english-grammar")
	installMockReadingScript(t, root, "courses/spanish-grammar")
	svc, err := NewService(root)
	if err != nil {
		t.Fatal(err)
	}
	return &genFixture{svc: svc, root: root}
}

func TestLoadGeneratedDraft(t *testing.T) {
	f := setupGenFixture(t)
	course, _ := f.svc.paths.Course("en_ru")
	readingDir := filepath.Join(t.TempDir(), "reading")

	tests := []struct {
		name    string
		setup   func() string
		wantErr string
		check   func(t *testing.T, id string, doc *TextDocument)
	}{
		{
			name: "missing index",
			setup: func() string {
				_ = os.MkdirAll(readingDir, 0o755)
				return readingDir
			},
			wantErr: "reading index missing",
		},
		{
			name: "bad index json",
			setup: func() string {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "index.json"), []byte("{"), 0o644)
				return dir
			},
			wantErr: "unexpected end",
		},
		{
			name: "empty texts map",
			setup: func() string {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "index.json"), []byte(`{"texts":{}}`), 0o644)
				return dir
			},
			wantErr: "no text generated",
		},
		{
			name: "missing text file",
			setup: func() string {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "index.json"), []byte(`{"texts":{"x":"texts/x.json"}}`), 0o644)
				return dir
			},
			wantErr: "no such file",
		},
		{
			name: "bad document json",
			setup: func() string {
				dir := t.TempDir()
				_ = os.MkdirAll(filepath.Join(dir, "texts"), 0o755)
				_ = os.WriteFile(filepath.Join(dir, "texts", "bad.json"), []byte("{"), 0o644)
				_ = os.WriteFile(filepath.Join(dir, "index.json"), []byte(`{"texts":{"bad":"texts/bad.json"}}`), 0o644)
				return dir
			},
			wantErr: "unexpected end",
		},
		{
			name: "fills target language from course",
			setup: func() string {
				dir := t.TempDir()
				_ = os.MkdirAll(filepath.Join(dir, "texts"), 0o755)
				doc := map[string]interface{}{
					"id": "fill_lang", "title": "T", "level": "A2",
					"reading_passage": map[string]interface{}{"segments": []interface{}{}},
				}
				raw, _ := json.Marshal(doc)
				_ = os.WriteFile(filepath.Join(dir, "texts", "fill.json"), raw, 0o644)
				_ = os.WriteFile(filepath.Join(dir, "index.json"), []byte(`{"texts":{"fill_lang":"texts/fill.json"}}`), 0o644)
				return dir
			},
			check: func(t *testing.T, id string, doc *TextDocument) {
				if doc.TargetLanguage != "en" {
					t.Fatalf("target_language=%q", doc.TargetLanguage)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := tc.setup()
			id, doc, err := f.svc.loadGeneratedDraft(dir, course)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err=%v want contains %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if id == "" || doc == nil {
				t.Fatal("expected id and doc")
			}
			if tc.check != nil {
				tc.check(t, id, doc)
			}
		})
	}
}

func TestCopyDirIfExists(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")

	if err := copyDirIfExists(filepath.Join(root, "missing"), dst); err != nil {
		t.Fatalf("missing src: %v", err)
	}
	if err := os.WriteFile(src, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyDirIfExists(src, dst); err != nil {
		t.Fatalf("file src: %v", err)
	}
	if err := os.Remove(src); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "nested", "a.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyDirIfExists(src, dst); err != nil {
		t.Fatalf("dir copy: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "nested", "a.txt"))
	if err != nil || string(got) != "data" {
		t.Fatalf("copied file: err=%v data=%q", err, got)
	}

	unreadable := filepath.Join(root, "unreadable")
	if err := os.MkdirAll(unreadable, 0o755); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(unreadable, "secret.bin")
	if err := os.WriteFile(secret, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(secret, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(secret, 0o644) })
	if err := copyDirIfExists(unreadable, filepath.Join(root, "dst2")); err == nil {
		t.Fatal("expected copy error for unreadable file")
	}
}

func TestRunReadingScript(t *testing.T) {
	f := setupGenFixture(t)
	course, _ := f.svc.paths.Course("en_ru")

	tests := []struct {
		name    string
		opts    scriptRunOptions
		pre     func()
		wantErr string
	}{
		{
			name: "success with title and audio",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "__MOCK_WITH_AUDIO__", withAudio: true,
				origin: OriginLLM, jobLog: "test",
			},
		},
		{
			name: "stderr warning on success",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "narrative",
				title: "__MOCK_STDERR_WARN__", origin: OriginLLM, jobLog: "warn",
			},
		},
		{
			name: "with input text",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "From text", inputText: "Hello source", origin: OriginManualText, jobLog: "text",
			},
		},
		{
			name: "with input json",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "From json",
				inputJSON: []byte(`{"id":"j1","title":"J","level":"A2","target_language":"en","reading_passage":{"segments":[]}}`),
				origin: OriginInputJSON, jobLog: "json",
			},
		},
		{
			name: "script failure",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "__MOCK_FAIL__", origin: OriginLLM, jobLog: "fail",
			},
			wantErr: "reading script failed",
		},
		{
			name: "empty index after script",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "__MOCK_EMPTY_INDEX__", origin: OriginLLM, jobLog: "empty",
			},
			wantErr: "no text generated",
		},
		{
			name: "bad index after script",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "__MOCK_BAD_INDEX__", origin: OriginLLM, jobLog: "bad idx",
			},
			wantErr: "invalid character",
		},
		{
			name: "missing generated text file",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "__MOCK_MISSING_TEXT__", origin: OriginLLM, jobLog: "missing",
			},
			wantErr: "no such file",
		},
		{
			name: "bad generated document",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "__MOCK_BAD_DOC__", origin: OriginLLM, jobLog: "bad doc",
			},
			wantErr: "invalid character",
		},
		{
			name: "no target language filled",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "__MOCK_NO_TARGET_LANG__", origin: OriginLLM, jobLog: "lang",
			},
		},
		{
			name: "script failure silent stderr",
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "__MOCK_FAIL_SILENT__", origin: OriginLLM, jobLog: "silent",
			},
			wantErr: "reading script failed",
		},
		{
			name: "staging dir blocked",
			pre: func() {
				stagingParent := filepath.Join(f.svc.paths.DataDir, "staging")
				_ = os.RemoveAll(stagingParent)
				_ = os.WriteFile(stagingParent, []byte("x"), 0o644)
			},
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "Staging blocked", origin: OriginLLM, jobLog: "staging",
			},
			wantErr: "not a directory",
		},
		{
			name: "gen work dir blocked",
			pre: func() {
				genWork := f.svc.paths.GenWorkDir()
				_ = os.RemoveAll(genWork)
				_ = os.WriteFile(genWork, []byte("x"), 0o644)
			},
			opts: scriptRunOptions{
				course: course, level: "A2", format: "dialogue",
				title: "Blocked", origin: OriginLLM, jobLog: "blocked",
			},
			wantErr: "not a directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pre != nil {
				tc.pre()
				t.Cleanup(func() {
					_ = os.Chmod(f.svc.paths.DraftsDir(), 0o755)
					_ = os.RemoveAll(f.svc.paths.GenWorkDir())
					_ = os.RemoveAll(filepath.Join(f.svc.paths.DataDir, "staging"))
					_ = f.svc.store.EnsureDirs()
				})
			}
			meta, err := f.svc.runReadingScript(context.Background(), tc.opts)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.wantErr)) {
					t.Fatalf("err=%v want contains %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if meta == nil || meta.TextID == "" {
				t.Fatal("expected draft meta")
			}
		})
	}
}

func TestGenerateBatch(t *testing.T) {
	f := setupGenFixture(t)

	t.Run("defaults count to one", func(t *testing.T) {
		created, err := f.svc.GenerateBatch(context.Background(), GenerateRequest{
			CourseCode: "en_ru", Level: "A2", Count: 0,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(created) != 1 {
			t.Fatalf("count=%d", len(created))
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		_, err := f.svc.GenerateBatch(context.Background(), GenerateRequest{
			CourseCode: "en_ru", Level: "Z9",
		})
		if err == nil || !strings.Contains(err.Error(), "level must be") {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("unknown course", func(t *testing.T) {
		_, err := f.svc.GenerateBatch(context.Background(), GenerateRequest{
			CourseCode: "xx_ru", Level: "A2",
		})
		if err == nil {
			t.Fatal("expected course error")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := f.svc.GenerateBatch(ctx, GenerateRequest{
			CourseCode: "en_ru", Level: "A2", Count: 2,
		})
		if err != context.Canceled {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("all runs fail", func(t *testing.T) {
		_, err := f.svc.GenerateBatch(context.Background(), GenerateRequest{
			CourseCode: "en_ru", Level: "A2", Count: 1, Title: "__MOCK_FAIL__",
		})
		if err == nil || !strings.Contains(err.Error(), "reading script failed") {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("batch partial success on multi count", func(t *testing.T) {
		courseDir := filepath.Join(f.root, "courses", "english-grammar")
		_ = os.Remove(filepath.Join(courseDir, ".mock_gen_counter_empty_title"))
		created, err := f.svc.GenerateBatch(context.Background(), GenerateRequest{
			CourseCode: "en_ru", Level: "A2", Count: 2, Title: "ignored for batch",
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(created) != 1 {
			t.Fatalf("partial created=%d", len(created))
		}
	})

	t.Run("success with audio and custom title", func(t *testing.T) {
		created, err := f.svc.GenerateBatch(context.Background(), GenerateRequest{
			CourseCode: "en_ru", Level: "A2", Count: 1,
			Format: "narrative", Title: "My story", WithAudio: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(created) != 1 || created[0].Title != "My story" {
			t.Fatalf("created=%+v", created)
		}
	})
}

func TestServerGenerateImportPromptPublish(t *testing.T) {
	f := setupGenFixture(t)
	handler := NewServer(f.svc, "").Handler()

	t.Run("generate success", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/drafts/generate", map[string]any{
			"course_code": "en_ru",
			"level":       "A2",
			"count":       1,
			"format":      "dialogue",
			"title":       "HTTP generate",
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
		var out map[string]interface{}
		decodeJSONBody(t, rec, &out)
		if int(out["total"].(float64)) != 1 {
			t.Fatalf("total=%v", out["total"])
		}
	})

	t.Run("generate service error", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/drafts/generate", map[string]any{
			"course_code": "en_ru",
			"level":       "A2",
			"title":       "__MOCK_FAIL__",
		})
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("generate method not allowed", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodGet, "/api/drafts/generate", nil)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("import text success", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/drafts/import-text", map[string]any{
			"course_code": "en_ru",
			"level":       "A2",
			"format":      "dialogue",
			"title":       "Imported",
			"text":        "Once upon a time.",
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("import text missing body", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/drafts/import-text", map[string]any{
			"course_code": "en_ru",
			"level":       "A2",
		})
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("import text method not allowed", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodGet, "/api/drafts/import-text", nil)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("import json success", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/drafts/import-json", map[string]any{
			"course_code": "en_ru",
			"level":       "A2",
			"document": map[string]any{
				"id":              "json_http_1",
				"title":           "JSON import",
				"level":           "A2",
				"target_language": "en",
				"reading_passage": map[string]any{"segments": []any{}},
			},
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("import json method not allowed", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodGet, "/api/drafts/import-json", nil)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("reading prompt generate", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/prompts/reading", map[string]any{
			"course_code": "en_ru",
			"level":       "A2",
			"format":      "dialogue",
			"kind":        "generate",
			"title":       "Prompt title",
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("reading prompt transform", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/prompts/reading", map[string]any{
			"course_code": "es_ru",
			"level":       "B1",
			"kind":        "transform",
			"source_text": "Hola mundo",
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("reading prompt bad kind", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/prompts/reading", map[string]any{
			"course_code": "en_ru",
			"level":       "A2",
			"kind":        "other",
		})
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("reading prompt transform missing source", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodPost, "/api/prompts/reading", map[string]any{
			"course_code": "en_ru",
			"level":       "A2",
			"kind":        "transform",
		})
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("reading prompt method not allowed", func(t *testing.T) {
		rec := doRequest(t, handler, http.MethodGet, "/api/prompts/reading", nil)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d", rec.Code)
		}
	})
}

func TestServerAdditionalHandlers(t *testing.T) {
	f := setupGenFixture(t)
	handler := NewServer(f.svc, "").Handler()

	meta := DraftMeta{
		TextID:         "pub_publish_flow",
		CourseCode:     "en_ru",
		Title:          "Publish flow",
		Level:          "A2",
		Format:         "dialogue",
		TargetLanguage: "en",
		Status:         StatusApproved,
		Origin:         OriginManualText,
	}
	doc := saveTestDraft(t, f.svc, meta, 1)
	textID := doc.ID
	staging := f.svc.paths.StagingDir(textID)
	segs := passageSegments(doc.ReadingPassage)
	for _, seg := range segs {
		rel := seg["audio_rel_path"].(string)
		abs := filepath.Join(staging, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("fake"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	doc2, err := f.svc.store.GetDocument(textID)
	if err != nil {
		t.Fatal(err)
	}
	meta.TextID = textID
	meta.Status = StatusAudioReady
	_, _, audioSt := AudioStats(doc2, staging)
	meta.AudioStatus = audioSt
	if err := f.svc.store.SaveDraft(meta, doc2); err != nil {
		t.Fatal(err)
	}

	audioMeta := DraftMeta{
		TextID: "audio_action_draft", CourseCode: "en_ru", Title: "Audio action",
		Level: "A2", Format: "dialogue", TargetLanguage: "en",
		Status: StatusApproved, Origin: OriginManualText,
	}
	audioDoc := saveTestDraft(t, f.svc, audioMeta, 1)
	rec := doRequest(t, handler, http.MethodPost, "/api/drafts/"+audioDoc.ID+"/audio", nil)
	if rec.Code != http.StatusOK && rec.Code != http.StatusBadRequest {
		t.Fatalf("audio action status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/drafts/missing_draft/audio", nil)
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusBadRequest {
		t.Fatalf("audio missing draft status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/drafts/"+textID+"/publish", map[string]any{
		"sync_bundle": false,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("publish status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/drafts/"+textID+"/publish", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("republish status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodGet, "/api/drafts/"+textID+"/publish", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("publish get status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/drafts/missing_draft/publish", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("publish missing status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/drafts/"+textID+"/cover", map[string]any{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("cover post status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/published/cover", map[string]any{
		"course_code": "en_ru",
		"text_id":     "missing_pub",
	})
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusBadRequest {
		t.Fatalf("published cover missing status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodPatch, "/api/published", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("published patch status=%d", rec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/published/sync", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("sync bad json status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodGet, "/api/course-audio/en_ru/x", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("course audio bad path status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodGet, "/api/course-audio/en_ru/nope/seg.mp3", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("course audio missing status=%d", rec.Code)
	}

	course, _ := f.svc.paths.Course("en_ru")
	imgDir := filepath.Join(course.GrammarDir, "assets", "reading", "img_demo")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imgDir, "cover.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec = doRequest(t, handler, http.MethodGet, "/api/course-images/en_ru/img_demo/cover.png", nil)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("course png status=%d ct=%q", rec.Code, rec.Header().Get("Content-Type"))
	}

	stagingID := "staging_png"
	stagingDir := filepath.Join(f.svc.paths.StagingDir(stagingID), "assets", "reading", stagingID)
	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stagingDir, "cover.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec = doRequest(t, handler, http.MethodGet, "/api/images/"+stagingID+"/cover.png", nil)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("staging png status=%d ct=%q", rec.Code, rec.Header().Get("Content-Type"))
	}

	rec = doRequest(t, handler, http.MethodPut, "/api/drafts/"+textID, map[string]any{"document": "bad"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("put bad document status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodPatch, "/api/drafts/"+textID, nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("patch draft status=%d", rec.Code)
	}

	rec = doRequest(t, handler, http.MethodGet, "/api/drafts?course_code=en_ru&audio=ready", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("drafts audio filter status=%d", rec.Code)
	}

	coverDraft := DraftMeta{
		TextID: "cover_delete_http", CourseCode: "en_ru", Title: "Cover delete",
		Level: "A2", TargetLanguage: "en", Status: StatusDraft,
	}
	coverDoc := saveTestDraft(t, f.svc, coverDraft, 0)
	rec = doRequest(t, handler, http.MethodDelete, "/api/drafts/"+coverDoc.ID+"/cover", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete draft cover status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, handler, http.MethodPost, "/api/drafts/"+coverDoc.ID+"/reject", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("reject draft status=%d", rec.Code)
	}

	pubTextID := "batch_cover_pub"
	pubDoc := &TextDocument{
		ID: pubTextID, Title: "Batch cover", Level: "A2", TargetLanguage: "es",
		CategoryID: "es_a2",
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{{"segment_id": "s1", "text": "Hola"}},
		},
	}
	courseES, _ := f.svc.paths.Course("es_ru")
	if err := writeTextToCourse(courseES.GrammarDir, pubDoc); err != nil {
		t.Fatal(err)
	}
	rec = doRequest(t, handler, http.MethodPost, "/api/covers/batch", map[string]any{
		"course_code": "es_ru",
		"limit":       1,
		"skip_llm":    true,
	})
	if rec.Code != http.StatusOK && rec.Code != http.StatusBadRequest {
		t.Fatalf("cover batch status=%d body=%s", rec.Code, rec.Body.String())
	}
}
