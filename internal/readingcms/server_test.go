package readingcms

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type cmsTestFixture struct {
	svc     *Service
	handler http.Handler
	root    string
}

func initCourseTree(t *testing.T, root, courseRel string) {
	t.Helper()
	courseDir := filepath.Join(root, courseRel)
	textsDir := filepath.Join(courseDir, "reading", "texts")
	if err := os.MkdirAll(textsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	idx := `{"version":"1.0.0","categories":{},"texts":{}}`
	if err := os.WriteFile(filepath.Join(courseDir, "reading", "index.json"), []byte(idx), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setupServerFixture(t *testing.T) *cmsTestFixture {
	t.Helper()
	root := t.TempDir()
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

func (f *cmsTestFixture) seedDraft(t *testing.T, meta DraftMeta, doc *TextDocument) {
	t.Helper()
	if meta.TextID == "" {
		meta.TextID = doc.ID
	}
	if err := f.svc.store.SaveDraft(meta, doc); err != nil {
		t.Fatal(err)
	}
}

func (f *cmsTestFixture) seedPublished(t *testing.T, courseCode, textID string, doc *TextDocument) {
	t.Helper()
	course, err := f.svc.paths.Course(courseCode)
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID == "" {
		doc.ID = textID
	}
	if err := writeTextToCourse(course.GrammarDir, doc); err != nil {
		t.Fatal(err)
	}
}

func doRequest(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeJSONBody(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(dst); err != nil {
		t.Fatalf("decode json: %v body=%q", err, rec.Body.String())
	}
}

func TestServerHealth(t *testing.T) {
	f := setupServerFixture(t)
	rec := doRequest(t, f.handler, http.MethodGet, "/api/health", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status=%d", rec.Code)
	}
}

func TestServerCoursesAndMethodNotAllowed(t *testing.T) {
	f := setupServerFixture(t)

	rec := doRequest(t, f.handler, http.MethodGet, "/api/courses", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("courses status=%d", rec.Code)
	}
	var coursesResp map[string][]map[string]string
	decodeJSONBody(t, rec, &coursesResp)
	if len(coursesResp["courses"]) != 2 {
		t.Fatalf("courses=%d", len(coursesResp["courses"]))
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/courses", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("post courses status=%d", rec.Code)
	}
}

func TestServerDraftsListAndFilters(t *testing.T) {
	f := setupServerFixture(t)
	meta := DraftMeta{
		TextID:         "draft_a2_en",
		CourseCode:     "en_ru",
		Title:          "Morning coffee",
		Level:          "A2",
		Format:         "dialogue",
		TargetLanguage: "en",
		Status:         StatusDraft,
		Origin:         OriginManualText,
		AudioStatus:    AudioNone,
	}
	doc := &TextDocument{
		ID:             meta.TextID,
		Title:          meta.Title,
		Level:          meta.Level,
		TargetLanguage: meta.TargetLanguage,
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedDraft(t, meta, doc)

	rec := doRequest(t, f.handler, http.MethodGet, "/api/drafts?course_code=en_ru&level=A2&search=coffee", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("drafts status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out map[string]interface{}
	decodeJSONBody(t, rec, &out)
	if int(out["total"].(float64)) != 1 {
		t.Fatalf("total=%v", out["total"])
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/drafts?search=missing", nil)
	decodeJSONBody(t, rec, &out)
	if int(out["total"].(float64)) != 0 {
		t.Fatalf("expected empty search, total=%v", out["total"])
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/drafts", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("post drafts status=%d", rec.Code)
	}
}

func TestServerDraftByIDLifecycle(t *testing.T) {
	f := setupServerFixture(t)
	meta := DraftMeta{
		TextID:         "draft_lifecycle",
		CourseCode:     "en_ru",
		Title:          "Lifecycle",
		Level:          "A2",
		Format:         "dialogue",
		TargetLanguage: "en",
		Status:         StatusDraft,
		Origin:         OriginManualText,
	}
	doc := &TextDocument{
		ID:             meta.TextID,
		Title:          meta.Title,
		Level:          meta.Level,
		TargetLanguage: meta.TargetLanguage,
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedDraft(t, meta, doc)

	rec := doRequest(t, f.handler, http.MethodGet, "/api/drafts/"+meta.TextID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get draft status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodPut, "/api/drafts/"+meta.TextID, map[string]any{
		"document": map[string]any{
			"id":              meta.TextID,
			"title":           "Updated title",
			"level":           "A2",
			"target_language": "en",
			"reading_passage": map[string]any{"segments": []any{}},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("put draft status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/drafts/"+meta.TextID+"/approve", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/drafts/"+meta.TextID+"/reject", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("reject status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/drafts/missing_draft", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing draft status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodDelete, "/api/drafts/"+meta.TextID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete draft status=%d", rec.Code)
	}
}

func TestServerDraftCoverDelete(t *testing.T) {
	f := setupServerFixture(t)
	textID := "draft_cover_http"
	meta := DraftMeta{
		TextID:         textID,
		CourseCode:     "en_ru",
		Title:          "Cover me",
		Level:          "A2",
		TargetLanguage: "en",
		Status:         StatusDraft,
	}
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	doc := &TextDocument{
		ID:                textID,
		Title:             meta.Title,
		Level:             meta.Level,
		TargetLanguage:    meta.TargetLanguage,
		CoverThumbRelPath: thumbRel,
		CoverImagePrompt:  "watercolor cafe",
		ReadingPassage:    map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedDraft(t, meta, doc)
	staging := f.svc.paths.StagingDir(textID)
	thumbAbs := filepath.Join(staging, filepath.FromSlash(thumbRel))
	if err := os.MkdirAll(filepath.Dir(thumbAbs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(thumbAbs, []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := doRequest(t, f.handler, http.MethodDelete, "/api/drafts/"+textID+"/cover", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete cover status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestServerPublishedListAndDetail(t *testing.T) {
	f := setupServerFixture(t)
	textID := "pub_es_a2_demo"
	doc := &TextDocument{
		ID:             textID,
		Title:          "Plaza walk",
		Level:          "A2",
		TargetLanguage: "es",
		CategoryID:     "es_a2",
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{{"segment_id": "s1", "text": "Hola"}},
		},
	}
	f.seedPublished(t, "es_ru", textID, doc)

	rec := doRequest(t, f.handler, http.MethodGet, "/api/published?course_code=es_ru&level=A2&search=plaza", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("published list status=%d", rec.Code)
	}
	var pub map[string]interface{}
	decodeJSONBody(t, rec, &pub)
	if int(pub["total"].(float64)) != 1 {
		t.Fatalf("total=%v", pub["total"])
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/published/detail?course_code=es_ru&text_id="+textID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("published detail status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/published/detail?course_code=es_ru", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("detail missing text_id status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/published/detail?course_code=es_ru&text_id=missing", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("detail missing text status=%d", rec.Code)
	}
}

func TestServerPublishedDelete(t *testing.T) {
	f := setupServerFixture(t)
	textID := "pub_delete_me"
	doc := &TextDocument{
		ID:             textID,
		Title:          "Delete me",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedPublished(t, "es_ru", textID, doc)

	rec := doRequest(t, f.handler, http.MethodDelete, "/api/published?course_code=es_ru&text_id="+textID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete published status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, f.handler, http.MethodDelete, "/api/published?course_code=es_ru&text_id="+textID, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("second delete status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodDelete, "/api/published?course_code=es_ru", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("delete missing params status=%d", rec.Code)
	}
}

func TestServerStaticAssets(t *testing.T) {
	f := setupServerFixture(t)
	textID := "asset_demo"
	staging := f.svc.paths.StagingDir(textID)
	assetDir := filepath.Join(staging, "assets", "reading", textID)
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetDir, "seg_01.mp3"), []byte("mp3"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetDir, "cover_thumb.webp"), []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := doRequest(t, f.handler, http.MethodGet, "/api/audio/"+textID+"/seg_01.mp3", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("audio status=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "audio/mpeg" {
		t.Fatalf("audio content-type=%q", ct)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/images/"+textID+"/cover_thumb.webp", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("image status=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("image content-type=%q", ct)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/audio/bad", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad audio path status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/images/"+textID+"/cover_thumb.webp", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("post image status=%d", rec.Code)
	}
}

func TestServerCourseStaticAssets(t *testing.T) {
	f := setupServerFixture(t)
	textID := "course_asset_demo"
	doc := &TextDocument{
		ID:             textID,
		Title:          "Course asset",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	f.seedPublished(t, "es_ru", textID, doc)
	course, err := f.svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	assetDir := filepath.Join(course.GrammarDir, "assets", "reading", textID)
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetDir, "seg_01.mp3"), []byte("mp3"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetDir, "cover_hero.jpeg"), []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := doRequest(t, f.handler, http.MethodGet, "/api/course-audio/es_ru/"+textID+"/seg_01.mp3", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("course audio status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/course-images/es_ru/"+textID+"/cover_hero.jpeg", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("course image status=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("jpeg content-type=%q", ct)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/course-images/xx_ru/"+textID+"/cover_hero.jpeg", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown course status=%d", rec.Code)
	}
}

func TestServerCoverProgressAndBatch(t *testing.T) {
	f := setupServerFixture(t)

	rec := doRequest(t, f.handler, http.MethodGet, "/api/cover-progress?course_code=es_ru&text_id=none", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("cover progress status=%d", rec.Code)
	}
	var prog map[string]interface{}
	decodeJSONBody(t, rec, &prog)
	if prog["running"].(bool) {
		t.Fatal("expected not running")
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/cover-progress?course_code=es_ru", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("cover progress missing text_id status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/cover-batch-progress?batch_id=", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("batch progress missing id status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/cover-batch-progress?batch_id=missing", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("batch progress missing job status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodPost, "/api/covers/batch", map[string]any{
		"course_code": "es_ru",
		"limit":       0,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("batch empty course texts status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestServerPublishedSync(t *testing.T) {
	f := setupServerFixture(t)
	textID := "sync_from_course"
	doc := &TextDocument{
		ID:             textID,
		Title:          "Sync from course",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{{"segment_id": "s1", "text": "Hola"}},
		},
	}
	f.seedPublished(t, "es_ru", textID, doc)

	rec := doRequest(t, f.handler, http.MethodPost, "/api/published/sync", map[string]any{
		"course_code": "es_ru",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("published sync status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestServerImportJSONBadRequest(t *testing.T) {
	f := setupServerFixture(t)
	rec := doRequest(t, f.handler, http.MethodPost, "/api/drafts/import-json", map[string]any{
		"course_code": "en_ru",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("import json missing document status=%d", rec.Code)
	}
}

func TestServerPublishedCoverDelete(t *testing.T) {
	f := setupServerFixture(t)
	textID := "pub_cover_del"
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	doc := &TextDocument{
		ID:                textID,
		Title:             "Cover delete",
		Level:             "A2",
		TargetLanguage:    "es",
		CoverThumbRelPath: thumbRel,
		CoverImagePrompt:  "prompt",
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

	rec := doRequest(t, f.handler, http.MethodDelete, "/api/published/cover", map[string]any{
		"course_code": "es_ru",
		"text_id":     textID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("delete published cover status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestServerBadJSONAndSubroutes(t *testing.T) {
	f := setupServerFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/drafts/generate", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/drafts/", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("empty draft subroute status=%d", rec.Code)
	}

	rec = doRequest(t, f.handler, http.MethodGet, "/api/drafts/demo/unknown", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown draft action status=%d", rec.Code)
	}
}

func TestServerHandlerWithWebRoot(t *testing.T) {
	root := t.TempDir()
	webRoot := filepath.Join(root, "web")
	if err := os.MkdirAll(webRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(webRoot, "index.html"), []byte("cms"), 0o644); err != nil {
		t.Fatal(err)
	}
	initCourseTree(t, root, "courses/english-grammar")
	svc, err := NewService(root)
	if err != nil {
		t.Fatal(err)
	}
	h := NewServer(svc, webRoot).Handler()
	rec := doRequest(t, h, http.MethodGet, "/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("web root status=%d", rec.Code)
	}
}

func TestMetaLastJobLog(t *testing.T) {
	if metaLastJobLog(nil) != "" {
		t.Fatal("nil meta")
	}
	if metaLastJobLog(&DraftMeta{LastJobLog: "x"}) != "x" {
		t.Fatal("expected log")
	}
}

func TestReadJSONEmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	var dst struct{}
	if err := readJSON(req, &dst); err != nil {
		t.Fatalf("empty body: %v", err)
	}
}
