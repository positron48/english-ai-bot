package repository

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestRealEsTrainingPackOnDiskLoads(t *testing.T) {
	t.Parallel()
	// CWD = package dir (internal/repository) when running package tests.
	root := filepath.Join("..", "grammartrainingpack", "es")
	if _, err := os.Stat(filepath.Join(root, "index.json")); err != nil {
		t.Skip("embedded grammar training pack not on disk in this environment")
	}
	r := NewGrammarTrainingPackRepositoryWithFS(os.DirFS(root), nil)
	ok, n, err := r.HasAnyQuestions()
	if err != nil {
		t.Fatalf("HasAnyQuestions: %v", err)
	}
	if !ok || n < 10 {
		t.Fatalf("expected many ES training questions, ok=%v n=%d", ok, n)
	}
	_, err = r.GetIndex()
	if err != nil {
		t.Fatalf("GetIndex: %v", err)
	}
}

func TestParseTrainingPackIndex_chapterString(t *testing.T) {
	raw := []byte(`{
  "version": "1.0.0",
  "language": "en",
  "chapters": { "c1": "a.json" }
}`)
	idx, err := parseTrainingPackIndex(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if idx.Chapters["c1"] != "a.json" {
		t.Fatalf("legacy chapters: %v", idx.Chapters)
	}
	if len(idx.chapterFiles["c1"]) != 1 || idx.chapterFiles["c1"][0] != "a.json" {
		t.Fatalf("chapterFiles: %#v", idx.chapterFiles)
	}
}

func TestParseTrainingPackIndex_chapterStringArray(t *testing.T) {
	raw := []byte(`{
  "chapters": { "c1": ["a.json", "b.json"] }
}`)
	idx, err := parseTrainingPackIndex(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(idx.chapterFiles["c1"]) != 2 {
		t.Fatalf("expected 2 files, got %#v", idx.chapterFiles["c1"])
	}
}

func TestParseTrainingPackIndex_blocks(t *testing.T) {
	raw := []byte(`{
  "chapters": { "orphan": ["044/x.json"] },
  "blocks": { "c1::b1": "001/b1.questions.json" }
}`)
	idx, err := parseTrainingPackIndex(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if idx.blockFiles["c1::b1"] != "001/b1.questions.json" {
		t.Fatalf("blocks: %#v", idx.blockFiles)
	}
}

func TestCollectTrainingPackFilePaths_blocksWin(t *testing.T) {
	idx, err := parseTrainingPackIndex([]byte(`{
  "chapters": { "c0": "legacy.json" },
  "blocks": { "a::b1": "001/a.json" }
}`))
	if err != nil {
		t.Fatal(err)
	}
	paths := collectTrainingPackFilePaths(idx)
	if len(paths) != 1 || paths[0] != "001/a.json" {
		t.Fatalf("expected blocks-only path, got %v", paths)
	}
}

func TestGetAllQuestions_v2(t *testing.T) {
	indexJSON := `{
  "version": "1.0.0",
  "language": "es",
  "chapters": {},
  "blocks": {
    "es.grammar.demo.chapter::b1_theory_x": "001/demo.questions.json"
  }
}`
	qfile := `{
  "chapter_id": "es.grammar.demo.chapter",
  "theory_block_id": "b1_theory_x",
  "questions": [
    {
      "id": "q1",
      "theory_block_id": "b1_theory_x",
      "chapter_id": "es.grammar.demo.chapter"
    }
  ]
}`
	mfs := fstest.MapFS{
		"index.json":            &fstest.MapFile{Data: []byte(indexJSON), Mode: 0o644},
		"chapters/001/demo.questions.json": &fstest.MapFile{Data: []byte(qfile), Mode: 0o644},
	}
	r := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
	qs, err := r.GetAllQuestions()
	if err != nil {
		t.Fatalf("GetAllQuestions: %v", err)
	}
	if len(qs) != 1 {
		t.Fatalf("len=%d", len(qs))
	}
	wantID := "es.grammar.demo.chapter::b1_theory_x::q1"
	if id, _ := qs[0]["id"].(string); id != wantID {
		t.Fatalf("id=%q want %q", id, wantID)
	}
	by, err := r.QuestionsByTheoryBlock()
	if err != nil {
		t.Fatalf("QuestionsByTheoryBlock: %v", err)
	}
	if len(by["b1_theory_x"]) != 1 {
		t.Fatalf("by block: %v", by)
	}
}

func TestGetChapterQuestions_v2(t *testing.T) {
	indexJSON := `{
  "version": "1.0.0",
  "chapters": {},
  "blocks": {
    "c1::b1": "001/x.json",
    "c1::b2": "001/y.json",
    "c2::b1": "002/z.json"
  }
}`
	mfs := fstest.MapFS{
		"index.json": &fstest.MapFile{Data: []byte(indexJSON), Mode: 0o644},
		"chapters/001/x.json": &fstest.MapFile{Data: []byte(`{"questions":[{"id":1,"chapter_id":"c1","theory_block_id":"b1"}]}`), Mode: 0o644},
		"chapters/001/y.json": &fstest.MapFile{Data: []byte(`{"questions":[{"id":2,"chapter_id":"c1","theory_block_id":"b2"}]}`), Mode: 0o644},
		"chapters/002/z.json": &fstest.MapFile{Data: []byte(`{"questions":[{"id":3,"chapter_id":"c2","theory_block_id":"b1"}]}`), Mode: 0o644},
	}
	r := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
	qs, err := r.GetChapterQuestions("c1")
	if err != nil {
		t.Fatalf("GetChapterQuestions: %v", err)
	}
	if len(qs) != 2 {
		t.Fatalf("len c1=%d", len(qs))
	}
}

func TestGetAllQuestions_emptyChapters(t *testing.T) {
	mfs := fstest.MapFS{
		"index.json": &fstest.MapFile{Data: []byte(`{
  "version": "1.0.0",
  "language": "en",
  "chapters": {}
}`), Mode: 0o644},
	}
	r := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
	qs, err := r.GetAllQuestions()
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 0 {
		t.Fatalf("expected 0 questions, got %d", len(qs))
	}
}

func TestGetAllQuestions_v1(t *testing.T) {
	indexJSON := `{
  "chapters": { "c1": "all.json" }
}`
	q := `{
  "chapter_id": "c1",
  "questions": [
    { "id": "q1", "chapter_id": "c1", "theory_block_id": "t1" }
  ]
}`
	mfs := fstest.MapFS{
		"index.json":     &fstest.MapFile{Data: []byte(indexJSON), Mode: 0o644},
		"chapters/all.json": &fstest.MapFile{Data: []byte(q), Mode: 0o644},
	}
	r := NewGrammarTrainingPackRepositoryWithFS(mfs, nil)
	qs, err := r.GetAllQuestions()
	if err != nil {
		t.Fatalf("GetAllQuestions: %v", err)
	}
	if len(qs) != 1 {
		t.Fatalf("len=%d", len(qs))
	}
	wantID := "c1::t1::q1"
	if id, _ := qs[0]["id"].(string); id != wantID {
		t.Fatalf("id=%q", id)
	}
}
