package readingcms

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncPublishedToCMS(t *testing.T) {
	root := t.TempDir()
	courseDir := filepath.Join(root, "courses", "spanish-grammar")
	textsDir := filepath.Join(courseDir, "reading", "texts")
	if err := os.MkdirAll(textsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	textID := "free_es_a2_test_sync"
	docJSON := `{
  "id": "` + textID + `",
  "title": "Sync test",
  "level": "A2",
  "target_language": "es",
  "category_id": "es_a2",
  "reading_passage": {"segments": [{"id":"s1","text":"Hola"}]}
}`
	if err := os.WriteFile(filepath.Join(textsDir, textID+".json"), []byte(docJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	idx := `{"version":"1.0.0","categories":{},"texts":{"` + textID + `":"texts/` + textID + `.json"}}`
	if err := os.WriteFile(filepath.Join(courseDir, "reading", "index.json"), []byte(idx), 0o644); err != nil {
		t.Fatal(err)
	}

	svc, err := NewService(root)
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.SyncPublishedToCMS(PublishedSyncRequest{CourseCode: "es_ru"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Imported != 1 || res.Total != 1 {
		t.Fatalf("imported=%d total=%d", res.Imported, res.Total)
	}
	meta, ok, err := svc.store.GetMeta(textID)
	if err != nil || !ok {
		t.Fatalf("meta missing: ok=%v err=%v", ok, err)
	}
	if meta.Status != StatusPublished || meta.Origin != OriginCourseImport {
		t.Fatalf("meta status=%s origin=%s", meta.Status, meta.Origin)
	}
	res2, err := svc.SyncPublishedToCMS(PublishedSyncRequest{CourseCode: "es_ru"})
	if err != nil {
		t.Fatal(err)
	}
	if res2.Skipped != 1 || res2.Imported != 0 {
		t.Fatalf("second sync skipped=%d imported=%d", res2.Skipped, res2.Imported)
	}
}
