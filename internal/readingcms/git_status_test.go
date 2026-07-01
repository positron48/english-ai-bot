package readingcms

import (
	"os/exec"
	"testing"
)

func TestListPublishedMarksNewUncommittedText(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	f := setupServerFixture(t)
	if out, err := exec.Command("git", "-C", f.root, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v output=%s", err, string(out))
	}

	textID := "pub_es_a2_new_uncommitted"
	doc := &TextDocument{
		ID:             textID,
		Title:          "Plaza nueva",
		Level:          "A2",
		TargetLanguage: "es",
		CategoryID:     "es_a2",
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{{"segment_id": "s1", "text": "Hola"}},
		},
	}
	f.seedPublished(t, "es_ru", textID, doc)

	items, err := f.svc.ListPublished("es_ru", "A2", "plaza", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%d", len(items))
	}
	if items[0].GitStatus != "untracked" {
		t.Fatalf("git_status=%q", items[0].GitStatus)
	}
	if !items[0].IsNewUncommitted {
		t.Fatalf("is_new_uncommitted=false")
	}

	items, err = f.svc.ListPublished("es_ru", "A2", "plaza", "", "committed")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("committed filter items=%d", len(items))
	}

	items, err = f.svc.ListPublished("es_ru", "A2", "plaza", "", "new")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("new filter items=%d", len(items))
	}
}
