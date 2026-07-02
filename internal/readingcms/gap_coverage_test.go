//go:build test

package readingcms

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const gapCoverageBaseID = 900001

func gapTextID(suffix string) string {
	return fmt.Sprintf("gap%d_%s", gapCoverageBaseID, suffix)
}

func installGapMockCoverScript(t *testing.T, root string) {
	installMockCoverScript(t, root)
}

func setupGapFixture(t *testing.T) *cmsTestFixture {
	t.Helper()
	t.Setenv("READING_COVER_PROMPT_MOCK", "1")
	root := t.TempDir()
	installGapMockCoverScript(t, root)
	initCourseTree(t, root, "courses/english-grammar")
	initCourseTree(t, root, "courses/spanish-grammar")
	svc, err := NewService(root)
	if err != nil {
		t.Fatal(err)
	}
	return &cmsTestFixture{
		svc:     svc,
		handler: NewServer(svc, "").Handler(),
		root:    root,
	}
}

func gapDraftDoc(textID string) (*DraftMeta, *TextDocument) {
	meta := DraftMeta{
		TextID:         textID,
		CourseCode:     "es_ru",
		Title:          "Gap coverage draft",
		Level:          "A2",
		Format:         "dialogue",
		TargetLanguage: "es",
		Status:         StatusApproved,
		Origin:         OriginManualText,
		AudioStatus:    AudioNone,
	}
	doc := &TextDocument{
		ID:             textID,
		Title:          meta.Title,
		Level:          meta.Level,
		TargetLanguage: meta.TargetLanguage,
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{
				{"segment_id": "s1", "text": "Hola", "speaker_id": "narrator"},
			},
		},
	}
	return &meta, doc
}

func TestGapCoverageStoreUpdateMeta(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("update_meta")
	meta, doc := gapDraftDoc(textID)
	f.seedDraft(t, *meta, doc)

	if err := f.svc.store.UpdateMeta(textID, func(m *DraftMeta) {
		m.Title = "Updated via UpdateMeta"
		m.LastJobLog = "gap update"
	}); err != nil {
		t.Fatal(err)
	}
	got, ok, err := f.svc.store.GetMeta(textID)
	if err != nil || !ok {
		t.Fatalf("GetMeta: ok=%v err=%v", ok, err)
	}
	if got.Title != "Updated via UpdateMeta" || got.LastJobLog != "gap update" {
		t.Fatalf("meta=%+v", got)
	}

	if err := f.svc.store.UpdateMeta("missing_"+textID, func(m *DraftMeta) {
		m.Title = "nope"
	}); err == nil || !strings.Contains(err.Error(), "draft not found") {
		t.Fatalf("UpdateMeta missing: err=%v", err)
	}
}

func TestGapCoverageStoreStagingAssetsDir(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("staging_assets")
	want := filepath.Join(f.svc.paths.StagingDir(textID), "assets", "reading", textID)
	if got := f.svc.store.StagingAssetsDir(textID); got != want {
		t.Fatalf("StagingAssetsDir=%q want %q", got, want)
	}
}

func TestGapCoverageWriteStagingCatalogForCover(t *testing.T) {
	textID := gapTextID("staging_catalog")
	doc := &TextDocument{
		ID:             textID,
		Title:          "Catalog",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	root := t.TempDir()
	if err := writeStagingCatalogForCover(root, doc); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "reading", "index.json")); err != nil {
		t.Fatalf("index.json: %v", err)
	}
	loaded, err := readStagingTextDoc(root, textID)
	if err != nil || loaded.Title != "Catalog" {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
}

func TestGapCoverageSyncPublishedCoverToDraft(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("sync_cover")
	meta, doc := gapDraftDoc(textID)
	f.seedDraft(t, *meta, doc)

	course, err := f.svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	heroRel := "assets/reading/" + textID + "/cover_hero.webp"
	pubDoc := &TextDocument{
		ID:                textID,
		Title:             doc.Title,
		Level:             doc.Level,
		TargetLanguage:    doc.TargetLanguage,
		CoverThumbRelPath: thumbRel,
		CoverHeroRelPath:  heroRel,
		CoverImagePrompt:  "synced prompt",
		ReadingPassage:    doc.ReadingPassage,
	}
	if err := writeTextToCourse(course.GrammarDir, pubDoc); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{thumbRel, heroRel} {
		abs := filepath.Join(course.GrammarDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("webp"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.svc.syncPublishedCoverToDraft(textID, pubDoc, course.GrammarDir); err != nil {
		t.Fatal(err)
	}
	stagingAssets := f.svc.store.StagingAssetsDir(textID)
	if _, err := os.Stat(filepath.Join(stagingAssets, "cover_thumb.webp")); err != nil {
		t.Fatalf("staging thumb: %v", err)
	}
	gotMeta, _, err := f.svc.GetDraft(textID)
	if err != nil {
		t.Fatal(err)
	}
	if gotMeta.CoverStatus != CoverReady || gotMeta.CoverImagePrompt != "synced prompt" {
		t.Fatalf("meta=%+v", gotMeta)
	}
}

func TestGapCoverageGenerateCover(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("gen_cover")
	meta, doc := gapDraftDoc(textID)
	f.seedDraft(t, *meta, doc)

	out, err := f.svc.GenerateCover(context.Background(), textID, CoverGenerateOpts{
		SkipLLM: true,
		Prompt:  "watercolor street scene at dusk",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.CoverStatus != CoverReady {
		t.Fatalf("status=%s log=%q", out.CoverStatus, out.LastJobLog)
	}
	_, doc2, err := f.svc.GetDraft(textID)
	if err != nil || strings.TrimSpace(doc2.CoverThumbRelPath) == "" {
		t.Fatalf("doc cover paths not set: %+v err=%v", doc2, err)
	}

	failID := gapTextID("gen_cover_fail")
	meta2, doc2 := gapDraftDoc(failID)
	meta2.TextID = failID
	doc2.ID = failID
	f.seedDraft(t, *meta2, doc2)
	_, err = f.svc.GenerateCover(context.Background(), failID, CoverGenerateOpts{
		SkipLLM: true,
		Prompt:  "__MOCK_COVER_FAIL__",
	})
	if err == nil {
		t.Fatal("expected cover script failure")
	}
}

func TestGapCoverageGeneratePublishedCover(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("pub_gen_cover")
	doc := &TextDocument{
		ID:             textID,
		Title:          "Published cover gen",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedPublished(t, "es_ru", textID, doc)

	item, log, err := f.svc.GeneratePublishedCover(context.Background(), "es_ru", textID, CoverGenerateOpts{
		SkipLLM: true,
		Prompt:  "mock published cover prompt",
	})
	if err != nil {
		t.Fatalf("GeneratePublishedCover: %v log=%q", err, log)
	}
	if item.CoverStatus != CoverReady || log == "" {
		t.Fatalf("item=%+v log=%q", item, log)
	}

	meta, docSeed := gapDraftDoc(textID)
	f.seedDraft(t, *meta, docSeed)
	syncID := gapTextID("pub_gen_sync")
	syncDoc := &TextDocument{
		ID:             syncID,
		Title:          "Sync after publish cover",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedPublished(t, "es_ru", syncID, syncDoc)
	syncMeta, syncDraftDoc := gapDraftDoc(syncID)
	syncMeta.TextID = syncID
	syncDraftDoc.ID = syncID
	f.seedDraft(t, *syncMeta, syncDraftDoc)
	if _, _, err := f.svc.GeneratePublishedCover(context.Background(), "es_ru", syncID, CoverGenerateOpts{
		SkipLLM: true,
		Prompt:  "sync draft after cover",
	}); err != nil {
		t.Fatal(err)
	}
	gotMeta, _, err := f.svc.GetDraft(syncID)
	if err != nil || gotMeta.CoverStatus != CoverReady {
		t.Fatalf("draft sync after published cover: meta=%+v err=%v", gotMeta, err)
	}
}

func TestGapCoverageDeletePublishedCover(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("pub_del_cover")
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	doc := &TextDocument{
		ID:                textID,
		Title:             "Delete published cover",
		Level:             "A2",
		TargetLanguage:    "es",
		CoverThumbRelPath: thumbRel,
		CoverImagePrompt:  "to delete",
		ReadingPassage:    map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedPublished(t, "es_ru", textID, doc)
	course, err := f.svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	thumbAbs := filepath.Join(course.GrammarDir, filepath.FromSlash(thumbRel))
	if err := os.MkdirAll(filepath.Dir(thumbAbs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(thumbAbs, []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}

	meta, draftDoc := gapDraftDoc(textID)
	meta.TextID = textID
	draftDoc.ID = textID
	draftDoc.CoverThumbRelPath = thumbRel
	draftDoc.CoverImagePrompt = "draft copy"
	f.seedDraft(t, *meta, draftDoc)
	stagingThumb := filepath.Join(f.svc.store.StagingAssetsDir(textID), "cover_thumb.webp")
	if err := os.MkdirAll(filepath.Dir(stagingThumb), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stagingThumb, []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}

	item, err := f.svc.DeletePublishedCover("es_ru", textID)
	if err != nil {
		t.Fatal(err)
	}
	if item.CoverStatus != CoverNone {
		t.Fatalf("item status=%s", item.CoverStatus)
	}
	if _, err := os.Stat(thumbAbs); !os.IsNotExist(err) {
		t.Fatalf("course thumb still exists: %v", err)
	}
	gotMeta, gotDoc, err := f.svc.GetDraft(textID)
	if err != nil {
		t.Fatal(err)
	}
	if gotMeta.CoverStatus != CoverNone || gotDoc.CoverThumbRelPath != "" {
		t.Fatalf("draft not synced: meta=%+v doc=%+v", gotMeta, gotDoc)
	}
}

func TestGapCoverageSyncPublishedCoverToDraftNoDraft(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("sync_no_draft")
	course, err := f.svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	doc := &TextDocument{
		ID:             textID,
		Title:          "No draft",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	if err := f.svc.syncPublishedCoverToDraft(textID, doc, course.GrammarDir); err != nil {
		t.Fatalf("expected nil when draft missing, got %v", err)
	}
}

func TestGapCoverageGenerateCoverNotFound(t *testing.T) {
	f := setupGapFixture(t)
	if _, err := f.svc.GenerateCover(context.Background(), gapTextID("missing"), CoverGenerateOpts{}); err == nil {
		t.Fatal("expected draft not found")
	}
}

func TestGapCoverageDeletePublishedCoverValidation(t *testing.T) {
	f := setupGapFixture(t)
	if _, err := f.svc.DeletePublishedCover("", gapTextID("bad")); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGapCoverageHandleDraftAction(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("draft_action")
	meta, doc := gapDraftDoc(textID)
	meta.Status = StatusDraft
	f.seedDraft(t, *meta, doc)

	rec := doRequest(t, f.handler, http.MethodPost, "/api/drafts/"+textID+"/approve", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/drafts/"+textID+"/reject", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("reject status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/drafts/"+textID+"/audio", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("audio status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/drafts/"+textID+"/approve", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("approve GET status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/drafts/missing_"+textID+"/approve", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing approve status=%d", rec.Code)
	}
}

func TestGapCoverageHandleDraftCover(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("draft_cover_http")
	meta, doc := gapDraftDoc(textID)
	f.seedDraft(t, *meta, doc)

	rec := doRequest(t, f.handler, http.MethodPost, "/api/drafts/"+textID+"/cover", map[string]any{
		"skip_llm": true,
		"prompt":   "HTTP draft cover prompt",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("post cover status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/drafts/"+textID+"/cover", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("get cover status=%d", rec.Code)
	}

	failID := gapTextID("draft_cover_fail")
	meta2, doc2 := gapDraftDoc(failID)
	meta2.TextID = failID
	doc2.ID = failID
	f.seedDraft(t, *meta2, doc2)
	rec = doRequest(t, f.handler, http.MethodPost, "/api/drafts/"+failID+"/cover", map[string]any{
		"skip_llm": true,
		"prompt":   "__MOCK_COVER_FAIL__",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("cover fail status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodDelete, "/api/drafts/missing_"+textID+"/cover", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing cover status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGapCoverageHandlePublishedCover(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("pub_cover_http")
	doc := &TextDocument{
		ID:             textID,
		Title:          "HTTP published cover",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedPublished(t, "es_ru", textID, doc)

	rec := doRequest(t, f.handler, http.MethodPost, "/api/published/cover", map[string]any{
		"course_code": "es_ru",
		"text_id":     textID,
		"skip_llm":    true,
		"prompt":      "HTTP published cover prompt",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("post published cover status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodDelete, "/api/published/cover", map[string]any{
		"course_code": "es_ru",
		"text_id":     textID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("delete published cover status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/published/cover", map[string]any{
		"course_code": "es_ru",
		"text_id":     "missing_" + textID,
		"skip_llm":    true,
		"prompt":      "nope",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing published cover status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodDelete, "/api/published/cover", map[string]any{
		"course_code": "es_ru",
		"text_id":     "missing_" + textID,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("delete missing published cover status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/published/cover", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET published cover status=%d", rec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/published/cover", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json status=%d", rec.Code)
	}
}

func TestGapCoverageHandleCoverBatch(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("cover_batch")
	f.seedPublished(t, "es_ru", textID, &TextDocument{
		ID:             textID,
		Title:          "Batch item",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	})

	rec := doRequest(t, f.handler, http.MethodPost, "/api/covers/batch", map[string]any{
		"course_code": "es_ru",
		"limit":       1,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("cover batch status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out map[string]interface{}
	decodeJSONBody(t, rec, &out)
	// total counts operations, not texts: one text without a saved prompt = prompt + image = 2 ops.
	if out["started"] != true || int(out["total"].(float64)) != 2 {
		t.Fatalf("batch response=%v", out)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/covers/batch", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET batch status=%d", rec.Code)
	}
}

func TestGapCoverageHandlePublishedSync(t *testing.T) {
	f := setupGapFixture(t)
	textID := gapTextID("published_sync")
	f.seedPublished(t, "es_ru", textID, &TextDocument{
		ID:             textID,
		Title:          "Sync me",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{{"segment_id": "s1", "text": "Hola"}},
		},
	})

	rec := doRequest(t, f.handler, http.MethodPost, "/api/published/sync", map[string]any{
		"course_code": "es_ru",
		"force":       true,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("sync status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/published/sync", map[string]any{
		"course_code": "bad_course",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad course sync status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/published/sync", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET sync status=%d", rec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/published/sync", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("sync bad json status=%d", rec.Code)
	}
}
